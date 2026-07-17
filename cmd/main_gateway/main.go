// Command main_gateway is the VOT Tradings middleware core: it loads
// configuration, connects to Postgres and Redis, wires up the brokerage
// drivers, and serves the JSON gateway described in the project README.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/brokerage/alpaca"
	"vot-tradings/internal/brokerage/oanda"
	"vot-tradings/internal/brokerage/questrade"
	"vot-tradings/internal/cache"
	"vot-tradings/internal/config"
	"vot-tradings/internal/db"
	"vot-tradings/internal/httpapi"
	"vot-tradings/pkg/logger"
)

const (
	schemaPath = "internal/db/schema.sql"
	assetsDir  = "assets"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Error("read schema", "error", err)
		os.Exit(1)
	}

	pgPool, err := db.Connect(ctx, cfg.Postgres, string(schema))
	if err != nil {
		log.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	redisClient, err := cache.Connect(ctx, cfg.Redis)
	if err != nil {
		log.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	tokenStore := db.NewOAuthTokenStore(pgPool)
	questradeClient := questrade.New(cfg.Questrade, tokenStore)
	questradeClient.Logger = log

	brokers := []brokerage.Broker{
		alpaca.New(cfg.Alpaca),
		oanda.New(cfg.OANDA),
		questradeClient,
	}

	srv := &httpapi.Server{
		Brokers:        brokers,
		DB:             pgPool,
		Cache:          redisClient,
		Logger:         log,
		USDCADRate:     cfg.USDCADRate,
		AssetsDir:      assetsDir,
		AllowedOrigins: cfg.CORSAllowedOrigins,
	}

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("gateway listening", "port", cfg.Port, "env", cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown", "error", err)
	}
}
