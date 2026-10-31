package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://bn_app:bn_secure_password_2026@127.0.0.1:5432/bn_appmarket?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger.Info("connecting to postgres for migrations", "url", strings.Split(databaseURL, "@")[len(strings.Split(databaseURL, "@"))-1])
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	// 1. Ensure schema_migrations table exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`
	if _, err := pool.Exec(ctx, createTableSQL); err != nil {
		logger.Error("failed to create schema_migrations table", "error", err)
		os.Exit(1)
	}

	// 2. Discover .up.sql files in migrations dir
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "db/migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			migrationsDir = "backend/db/migrations"
		}
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil || len(files) == 0 {
		logger.Error("no migration files found", "dir", migrationsDir, "error", err)
		os.Exit(1)
	}
	sort.Strings(files)

	for _, file := range files {
		baseName := filepath.Base(file)
		version := strings.TrimSuffix(baseName, ".up.sql")

		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			logger.Error("failed to query migration status", "version", version, "error", err)
			os.Exit(1)
		}

		if exists {
			logger.Info("migration already applied, skipping", "version", version)
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			logger.Error("failed to read migration file", "file", file, "error", err)
			os.Exit(1)
		}

		logger.Info("applying migration", "version", version)
		tx, err := pool.Begin(ctx)
		if err != nil {
			logger.Error("failed to start transaction", "error", err)
			os.Exit(1)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			logger.Error("failed to execute migration SQL", "version", version, "error", err)
			os.Exit(1)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = tx.Rollback(ctx)
			logger.Error("failed to record migration", "version", version, "error", err)
			os.Exit(1)
		}

		if err := tx.Commit(ctx); err != nil {
			logger.Error("failed to commit migration transaction", "version", version, "error", err)
			os.Exit(1)
		}
		logger.Info("migration successfully applied", "version", version)
	}

	logger.Info("all database migrations completed successfully")
	fmt.Println("Database schema up to date.")
}
