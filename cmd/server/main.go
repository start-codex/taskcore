package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/taskcode/internal/boards"
	"github.com/start-codex/taskcode/internal/issues"
	"github.com/start-codex/taskcode/internal/issuetypes"
	"github.com/start-codex/taskcode/internal/projects"
	"github.com/start-codex/taskcode/internal/statuses"
	"github.com/start-codex/taskcode/internal/users"
	"github.com/start-codex/taskcode/internal/workspaces"
	"github.com/start-codex/taskcode/migrations"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	if err := migrations.Up(ctx, db.DB); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// API routes live on a sub-mux so they're served under /api/*.
	// Each domain package registers paths like "POST /users" and the
	// StripPrefix removes "/api" before matching, so no domain package
	// needs to know about the prefix.
	api := http.NewServeMux()
	users.RegisterRoutes(api, db)
	workspaces.RegisterRoutes(api, db)
	projects.RegisterRoutes(api, db)
	statuses.RegisterRoutes(api, db)
	issuetypes.RegisterRoutes(api, db)
	boards.RegisterRoutes(api, db)
	issues.RegisterRoutes(api, db)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	registerUI(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      withRecover(withLogger(withRequestID(mux))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutting down gracefully")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
