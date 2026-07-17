-- VOT Tradings primary relational ledger schema.
-- Applied automatically on Postgres container init (see docker-compose.yml),
-- or manually via: psql "$DATABASE_URL" -f internal/db/schema.sql

CREATE TABLE IF NOT EXISTS accounts (
    id                 TEXT PRIMARY KEY,
    broker             TEXT NOT NULL,
    currency           TEXT NOT NULL,
    equity             NUMERIC(18, 4) NOT NULL DEFAULT 0,
    buying_power       NUMERIC(18, 4) NOT NULL DEFAULT 0,
    cash               NUMERIC(18, 4) NOT NULL DEFAULT 0,
    pattern_day_trader BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS positions (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    broker          TEXT NOT NULL,
    symbol          TEXT NOT NULL,
    asset_class     TEXT NOT NULL,
    quantity        NUMERIC(18, 8) NOT NULL,
    avg_entry_price NUMERIC(18, 8) NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, symbol)
);

CREATE TABLE IF NOT EXISTS orders (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    broker          TEXT NOT NULL,
    symbol          TEXT NOT NULL,
    side            TEXT NOT NULL,
    type            TEXT NOT NULL,
    quantity        NUMERIC(18, 8) NOT NULL,
    limit_price     NUMERIC(18, 8),
    status          TEXT NOT NULL,
    broker_order_id TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_account_id ON orders (account_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

-- Self-correcting feedback loop ledger: one row per DL engine forecast,
-- reconciled with the realized outcome once the horizon window elapses.
CREATE TABLE IF NOT EXISTS predictions (
    inference_id      TEXT PRIMARY KEY,
    symbol            TEXT NOT NULL,
    asset_class       TEXT NOT NULL,
    horizon_minutes   INTEGER NOT NULL,
    predicted_price   NUMERIC(18, 8) NOT NULL,
    predicted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    actual_price      NUMERIC(18, 8),
    resolved_at       TIMESTAMPTZ,
    abs_error         NUMERIC(18, 8),
    direction_correct BOOLEAN,
    model_version     TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_predictions_symbol ON predictions (symbol);
CREATE INDEX IF NOT EXISTS idx_predictions_unresolved ON predictions (resolved_at) WHERE resolved_at IS NULL;

-- Pattern Day Trader compliance: one row per round-trip (same symbol bought
-- and sold within the same session) used to enforce the rolling 5-business
-- day day-trade count for margin accounts under $25,000 USD equity.
CREATE TABLE IF NOT EXISTS day_trades (
    id         TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    symbol     TEXT NOT NULL,
    trade_date DATE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_day_trades_account_date ON day_trades (account_id, trade_date);

-- OAuth refresh-token persistence for brokers whose tokens rotate on every
-- exchange (currently Questrade: the previous refresh token is invalidated
-- the moment a new one is issued). Without this, a process restart burns
-- access and requires re-authorizing through the broker's website.
CREATE TABLE IF NOT EXISTS broker_oauth_tokens (
    broker        TEXT PRIMARY KEY,
    refresh_token TEXT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
