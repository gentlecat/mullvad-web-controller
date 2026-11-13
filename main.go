package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.roman.zone/mullvad-web-controller/api"
	"go.roman.zone/mullvad-web-controller/config"
)

//go:embed content/*.html content/static/*
var content embed.FS

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	static, err := fs.Sub(content, "content")
	if err != nil {
		return fmt.Errorf("failed to load static content: %w", err)
	}

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: api.NewServer(cfg.DryRun).Routes(static),
		// No ReadTimeout: request bodies are tiny and capped, so the header timeout is enough.
		ReadHeaderTimeout: 10 * time.Second,
		// Upstream calls retry, so leave room for a slow relay list download.
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "url", "http://"+cfg.Addr, "dryRun", cfg.DryRun)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
