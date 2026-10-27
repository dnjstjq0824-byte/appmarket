package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/bookingnetwork/appmarket/backend/internal/platform/httpserver"
)

func requiredEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		slog.Error("required environment variable missing", "name", name)
		os.Exit(2)
	}
	return v
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	traceExporter, err := otlptracehttp.New(initCtx)
	if err != nil {
		logger.Error("OTel exporter initialization failed", "error", err)
		os.Exit(1)
	}
	res, err := resource.New(initCtx, resource.WithAttributes(semconv.ServiceName("bn-appmarket-api")))
	if err != nil {
		logger.Error("OTel resource initialization failed", "error", err)
		os.Exit(1)
	}
	tracer := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(res))
	otel.SetTracerProvider(tracer)
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = tracer.Shutdown(shutdownCtx)
	}()

	db, err := pgxpool.New(initCtx, requiredEnv("DATABASE_URL"))
	if err != nil {
		logger.Error("postgres configuration failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(initCtx); err != nil {
		logger.Error("postgres unavailable", "error", err)
		os.Exit(1)
	}
	redisOptions, err := redis.ParseURL(requiredEnv("REDIS_URL"))
	if err != nil {
		logger.Error("redis configuration failed", "error", err)
		os.Exit(1)
	}
	cache := redis.NewClient(redisOptions)
	defer cache.Close()
	if err := cache.Ping(initCtx).Err(); err != nil {
		logger.Error("redis unavailable", "error", err)
		os.Exit(1)
	}

	app, err := httpserver.New(httpserver.Config{
		DB: db, Redis: cache, OpenSearchURL: requiredEnv("OPENSEARCH_URL"),
		S3HealthURL: requiredEnv("S3_HEALTH_URL"), OpenSearchUser: os.Getenv("OPENSEARCH_USER"),
		OpenSearchPass: os.Getenv("OPENSEARCH_PASSWORD"), AllowedOrigin: os.Getenv("ALLOWED_ORIGIN"),
		TrustedProxyCIDRs: splitCIDRs(os.Getenv("TRUSTED_PROXY_CIDRS")), Logger: logger,
	})
	if err != nil {
		logger.Error("HTTP server configuration failed", "error", err)
		os.Exit(1)
	}

	server := &http.Server{Addr: ":8080", Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	logger.Info("API listening", "addr", server.Addr)
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("API listener stopped", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, c := context.WithTimeout(context.Background(), 30*time.Second)
		defer c()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
	}
}

func splitCIDRs(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}
