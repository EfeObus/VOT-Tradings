// Command main_gateway is the VOT Tradings middleware core: it loads
// configuration, connects to Postgres and Redis, wires up auth and each
// user's own brokerage credentials, and serves the JSON gateway described
// in the project README.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vot-tradings/internal/auth"
	"vot-tradings/internal/cache"
	"vot-tradings/internal/config"
	"vot-tradings/internal/db"
	"vot-tradings/internal/httpapi"
	"vot-tradings/internal/userbrokers"
	"vot-tradings/pkg/crypto"
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

	credentialBox, err := crypto.NewBox(cfg.CredentialEncryptionKey)
	if err != nil {
		log.Error("init credential encryption", "error", err)
		os.Exit(1)
	}

	users := db.NewUserStore(pgPool)
	sessions := auth.NewSessionStore(redisClient)
	credentials := db.NewCredentialStore(pgPool, credentialBox)
	tokens := db.NewOAuthTokenStore(pgPool)
	brokerFactory := userbrokers.NewFactory(credentials, tokens)
	orders := db.NewOrderStore(pgPool)
	dayTrades := db.NewDayTradeStore(pgPool)

	srv := &httpapi.Server{
		DB:             pgPool,
		Cache:          redisClient,
		Logger:         log,
		Config:         cfg,
		Users:          users,
		Sessions:       sessions,
		Credentials:    credentials,
		Brokers:        brokerFactory,
		Orders:         orders,
		DayTrades:      dayTrades,
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
