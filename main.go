package main

import (
	"context"
	"fmt"
	"log/slog"
	"minecraft-proxy/internal"
	"os"
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

	listenAddr, err := internal.GetListenAddr(logger)
	if err != nil {
		return fmt.Errorf("failed to get listen addr: %w", err)
	}

	mapping, err := internal.GetServerMapping()
	if err != nil {
		return fmt.Errorf("failed to get server mapping: %w", err)
	}

	ctx := context.Background()
	app := internal.NewApp(logger, listenAddr, mapping)
	return app.Run(ctx)
}
