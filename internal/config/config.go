// Package config loads VOT Tradings runtime configuration from environment
// variables (populated via .env in local development).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	Env      string
	LogLevel string

	// USDCADRate is a static fallback CAD->USD conversion rate used by the
	// unified balance engine until a live spot-rate feed (e.g. from OANDA's
	// USD_CAD pricing stream) is wired in.
	USDCADRate float64

	// CORSAllowedOrigins lists the exact Origin values the web client is
	// served from, so the gateway can allow its browser fetches through.
	CORSAllowedOrigins []string

	// CredentialEncryptionKey is a base64-encoded 32-byte AES-256 key used
	// to encrypt each user's stored brokerage credentials at rest (see
	// pkg/crypto). There is no safe default — Load panics if it's unset,
	// since running without it would mean either refusing to start or
	// silently storing secrets in plaintext, and the latter is worse.
	CredentialEncryptionKey string

	Postgres  PostgresConfig
	Redis     RedisConfig
	Alpaca    AlpacaConfig
	OANDA     OANDAConfig
	Questrade QuestradeConfig
	DLEngine  DLEngineConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	SSLMode  string
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DB, p.SSLMode)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type AlpacaConfig struct {
	APIKeyID  string
	SecretKey string
	BaseURL   string
	DataURL   string
}

type OANDAConfig struct {
	AccountID   string
	AccessToken string
	BaseURL     string
}

type QuestradeConfig struct {
	RefreshToken string
	AuthURL      string
}

type DLEngineConfig struct {
	URL        string
	TimeoutMS  int
}

// Load reads configuration from environment variables, applying sane local
// development defaults for anything not explicitly set. It first loads a
// local .env file if one is present (see .env.example); real env vars
// already set in the process environment always take precedence.
func Load() Config {
	_ = godotenv.Load()

	credentialKey := getEnv("CREDENTIAL_ENCRYPTION_KEY", "")
	if credentialKey == "" {
		fmt.Fprintln(os.Stderr, "config: CREDENTIAL_ENCRYPTION_KEY is not set.")
		fmt.Fprintln(os.Stderr, "Generate one with: openssl rand -base64 32")
		fmt.Fprintln(os.Stderr, "and set it in .env — refusing to start rather than store broker credentials in plaintext.")
		os.Exit(1)
	}

	return Config{
		Port:                    getEnv("PORT", "8080"),
		Env:                     getEnv("ENV", "development"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		USDCADRate:              getEnvFloat("USD_CAD_RATE", 0.73),
		CORSAllowedOrigins:      getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		CredentialEncryptionKey: credentialKey,
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "vot"),
			Password: getEnv("POSTGRES_PASSWORD", "vot_dev_password"),
			DB:       getEnv("POSTGRES_DB", "vot_tradings"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Alpaca: AlpacaConfig{
			APIKeyID:  getEnv("ALPACA_API_KEY_ID", ""),
			SecretKey: getEnv("ALPACA_SECRET_KEY", ""),
			BaseURL:   getEnv("ALPACA_BASE_URL", "https://paper-api.alpaca.markets"),
			DataURL:   getEnv("ALPACA_DATA_URL", "https://data.alpaca.markets"),
		},
		OANDA: OANDAConfig{
			AccountID:   getEnv("OANDA_ACCOUNT_ID", ""),
			AccessToken: getEnv("OANDA_ACCESS_TOKEN", ""),
			BaseURL:     getEnv("OANDA_BASE_URL", "https://api-fxpractice.oanda.com"),
		},
		Questrade: QuestradeConfig{
			RefreshToken: getEnv("QUESTRADE_REFRESH_TOKEN", ""),
			AuthURL:      getEnv("QUESTRADE_AUTH_URL", "https://login.questrade.com/oauth2/token"),
		},
		DLEngine: DLEngineConfig{
			URL:       getEnv("DL_ENGINE_URL", "http://localhost:5000"),
			TimeoutMS: getEnvInt("DL_ENGINE_TIMEOUT_MS", 2000),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
