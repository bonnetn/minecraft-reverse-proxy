package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
)

const (
	envVarNameListenAddr = "PROXY_LISTEN_ADDR"
	envVarNameMapping    = "PROXY_MAPPING"
	defaultListenAddr    = ":25565"
)

type ServerMapping struct {
	Default string            `json:"default,omitempty"`
	Servers map[string]string `json:"servers,omitempty"`
}

func GetListenAddr(logger *slog.Logger) (string, error) {
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

func GetServerMapping() (ServerMapping, error) {
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
