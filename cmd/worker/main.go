package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ads-aggregator/ads-aggregator/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().
		Str("version", Version).
		Str("build_time", BuildTime).
		Str("git_commit", GitCommit).
		Msg("Starting Ads Analytics Worker")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health check server
	healthPort := 8081
	healthServer := startHealthServer(healthPort)

	// TODO: Initialize database connection
	// TODO: Initialize Redis connection
	// TODO: Initialize scheduler with cron jobs
	// TODO: Start data sync workers
	// TODO: Start token refresh workers

	log.Info().
		Str("app", cfg.App.Name).
		Str("env", cfg.App.Env).
		Bool("scheduler_enabled", cfg.Scheduler.Enabled).
		Msg("Worker started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down worker...")

	// Graceful shutdown
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Health server shutdown error")
	}

	// Give ongoing jobs time to complete
	time.Sleep(5 * time.Second)

	log.Info().Msg("Worker exited")
	_ = ctx // silence unused variable
}

// startHealthServer starts an HTTP server for health checks
func startHealthServer(port int) *http.Server {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"worker","version":"%s"}`, Version)
	})

	// Readiness check
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Check database connection
		// TODO: Check Redis connection
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ready"}`)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Info().Int("port", port).Msg("Health server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Health server error")
		}
	}()

	return server
}
