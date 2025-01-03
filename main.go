package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"net"
	"os"
)

const (
	envVarNameListenAddr = "PROXY_LISTEN_ADDR"
	envVarNameMapping    = "PROXY_MAPPING"
	defaultListenAddr    = ":25565"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func mainErr() error {
	var logLevel slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var logger *slog.Logger
	if os.Getenv("LOG_FORMAT") == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	}

	listenAddr, err := getListenAddr(logger)
	if err != nil {
		return fmt.Errorf("failed to get listen addr: %w", err)
	}

	mapping, err := getServerMapping()
	if err != nil {
		return fmt.Errorf("failed to get server mapping: %w", err)
	}

	ctx := context.Background()
	app := App{logger: logger, listenAddr: listenAddr, mapping: mapping}

	return app.Run(ctx)
}

func getListenAddr(logger *slog.Logger) (string, error) {
	listenAddrStr, ok := os.LookupEnv(envVarNameListenAddr)
	if !ok {
		logger.Info(envVarNameListenAddr+" environment variable is not set, using default", "listen_addr", defaultListenAddr)
		return defaultListenAddr, nil
	}

	err := validateAddr(listenAddrStr)
	if err != nil {
		return "", fmt.Errorf("invalid "+envVarNameListenAddr+": %w", err)
	}

	return listenAddrStr, nil
}

func getServerMapping() (ServerMapping, error) {
	mappingJson, ok := os.LookupEnv(envVarNameMapping)
	if !ok {
		return ServerMapping{}, errors.New(envVarNameMapping + " environment variable is not set")
	}

	var mapping ServerMapping
	if err := json.Unmarshal([]byte(mappingJson), &mapping); err != nil {
		return ServerMapping{}, fmt.Errorf("failed to decode "+envVarNameMapping+" environment variable: %w", err)
	}

	if mapping.Default != "" {
		if err := validateAddr(mapping.Default); err != nil {
			return ServerMapping{}, err
		}
	}

	for domain, addr := range mapping.Servers {
		if domain == "" {
			return ServerMapping{}, errors.New("empty domain")
		}

		if err := validateAddr(addr); err != nil {
			return ServerMapping{}, fmt.Errorf("invalid server address for domain %q: %w", domain, err)
		}
	}

	return mapping, nil
}

type App struct {
	logger     *slog.Logger
	listenAddr string
	mapping    ServerMapping
}

func (a *App) Run(ctx context.Context) error {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", a.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			a.logger.Error("failed to close listener", "error", err)
		}
	}()

	a.logger.Info("Proxy is listening", "address", l.Addr().String())

	var connID int64
	g, ctx := errgroup.WithContext(ctx)
	for {
		c, err := l.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept: %w", err)
		}

		connID++

		currentConID := connID
		g.Go(func() error {
			return a.handleConnection(ctx, c, currentConID)
		})
	}
}

func (a *App) handleConnection(ctx context.Context, gameClient net.Conn, connID int64) error {
	logger := a.logger.With("connection_id", connID)
	defer func() {
		if err := gameClient.Close(); err != nil {
			logger.Error("failed to close connection to game client", "error", err)
		}
	}()

	logger = logger.With("client_addr", gameClient.RemoteAddr().String())
	logger.Info("Accepted connection")

	packet, firstBytes, err := readFirstPacket(gameClient)
	if err != nil {
		return fmt.Errorf("failed to read first packet: %w", err)
	}
	logger = logger.With("domain", packet.ServerAddress)
	logger.Debug("Received packet", "packet", packet)

	var remoteAddr string

	serverAddr, ok := a.mapping.Servers[packet.ServerAddress]
	if !ok {
		if a.mapping.Default == "" {
			logger.Warn("No server mapping found for domain", "domain", packet.ServerAddress)
			return nil
		}
		remoteAddr = a.mapping.Default
		logger.Info("No server mapping found for domain, using default server", "remote_addr", remoteAddr)
	} else {
		remoteAddr = serverAddr
		logger.Info("Found server mapping for domain", "remote_addr", remoteAddr)
	}

	minecraftServer, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer func() {
		if err := minecraftServer.Close(); err != nil {
			logger.Error("failed to close connection to minecraft server", "error", err)
		}
	}()
	logger = logger.With("server_addr", minecraftServer.RemoteAddr().String())
	logger.Debug("Connected to minecraft server")

	n, err := io.Copy(minecraftServer, &firstBytes)
	if err != nil {
		return fmt.Errorf("failed to copy first bytes: %w", err)
	}
	logger.Debug("Forwarded first bytes to server", "bytes", n)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if _, err := io.Copy(gameClient, minecraftServer); err != nil {
			return fmt.Errorf("failed to copy minecraftServer -> gameClient: %w", err)
		}

		logger.Debug("Stopped forwarding packets from server to client")
		return nil
	})

	g.Go(func() error {
		if _, err := io.Copy(minecraftServer, gameClient); err != nil {
			return fmt.Errorf("failed to copy gameClient -> minecraftServer: %w", err)
		}

		logger.Debug("Stopped forwarding packets from client to server")
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to wait: %w", err)
	}

	logger.Info("Connection closed")
	return nil
}

func readFirstPacket(reader io.Reader) (Packet, bytes.Buffer, error) {
	var firstBytes bytes.Buffer
	teeReader := io.TeeReader(reader, &firstBytes)

	packet, err := getNextPacket(teeReader)
	if err != nil {
		return Packet{}, firstBytes, fmt.Errorf("failed to get next packet: %w", err)
	}
	return packet, firstBytes, nil
}

type ServerMapping struct {
	Default string            `json:"default,omitempty"`
	Servers map[string]string `json:"servers,omitempty"`
}

func validateAddr(addr string) error {
	if addr == "" {
		return errors.New("empty address")
	}

	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid server address %q: %w", addr, err)
	}
	return err
}
