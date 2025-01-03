package internal

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"net"
)

type App struct {
	logger     *slog.Logger
	listenAddr string
	mapping    ServerMapping
}

func NewApp(logger *slog.Logger, listenAddr string, mapping ServerMapping) *App {
	return &App{logger: logger, listenAddr: listenAddr, mapping: mapping}
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
