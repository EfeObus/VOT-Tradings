# VOT Tradings

### Intelligent Multi-Asset Fintech Architecture for US & Canadian Capital Markets

VOT Tradings is an enterprise-ready, low-latency automated trading and asset management platform. Its architecture bridges North American equity markets (US & Canada) with global foreign exchange (Forex) venues, underwritten by a self-correcting, deep learning prediction and optimization engine.

The platform functions as an intelligent middleware layer. It unifies isolated brokerage APIs, streams high-frequency market data into localized time-series caches, and executes automated order routing strategies informed by deep learning forecasting models that continuously learn from market state changes and historical execution telemetry.

---

## Current Implementation Status

This section reflects what actually runs today, not the target architecture. See the sections below for the full design.

**Built and verified working:**
- Go middleware gateway (`cmd/main_gateway`) — loads config/`.env`, connects Postgres + Redis, applies the schema, serves HTTP
- Brokerage drivers for Alpaca, OANDA, and Questrade (`internal/brokerage/`) — account + quote reads for all three; order placement implemented for Alpaca and OANDA, not yet for Questrade (needs a live practice account to verify its two-step order-impact/commit flow)
- Cross-broker unified balance engine and FINRA Pattern Day Trader check (`internal/engine/`)
- Postgres-backed persistence for Questrade's rotating OAuth refresh token (`internal/db/oauth_tokens.go`) — without this, a process restart burns Questrade access, since its refresh tokens are single-use
- HTTP API: `GET /healthz`, `GET /api/v1/balance`, `GET /api/v1/quote`, `GET /logo.png` (see [Gateway HTTP API](#gateway-http-api))
- Web client (`web/`) — React + TypeScript + Tailwind, eight-page shell (`/profile`, `/dashboard`, `/funds`, `/market/:symbol`, `/trade`, `/forecasts`, `/reports`, `/settings`) matching the platform's target IA. **Dashboard** (unified NAV, cross-border split, per-broker USD allocation donut), **Settings** (broker connectivity audit), **Market data** (on-demand real quote lookup), and **Funds** (links to each broker's real funding portal) are fully real, backed by the gateway. **Trade**, **Forecasts**, **Reports**, and **Profile** render the correct layout but each panel explicitly states what backend piece it's waiting on — see below — rather than showing placeholder/fabricated numbers or a cosmetic login form.
- Local dev infra: `docker-compose.yml` (Postgres + Redis Stack)

**Not yet built** (described below as the target architecture, and stated inline in the web client's `NotConnected` panels):
- Authentication — the gateway has no user accounts, sessions, or per-user broker credentials; it's single-tenant, using one shared set of credentials from `.env` for everyone. Blocks: real Profile/register/login, and any notion of "this data belongs to this user."
- Python DL engine (`services/dl_engine/`) — LSTM/Transformer forecasters, RL execution optimizer, ONNX inference, self-correcting feedback loop. Blocks: Forecasts page's forecasting matrix, confidence tracking, self-correction log.
- Real-time streaming ingestion pipeline (`cmd/data_pipeline/`) — Kafka, WebSocket market data fan-out. Blocks: Market page's candlestick stream, Level 2 depth, indicator overlays. (Market data quote *lookups* are real and live today — see above; there's just no streaming tick feed yet.)
- Order-execution HTTP endpoint — `PlaceOrder` exists per-broker (`internal/brokerage/`) but the gateway doesn't expose it over HTTP, and `internal/engine/pdt.go`'s PDT check isn't wired to the API either. Blocks: Trade page's order ticket and PDT risk shield, and Reports (nothing persists to the `orders` table yet).
- Production safeguards described in [Production Considerations](#production-considerations-and-security-guardrails): CAD/USD slippage cushion, DL engine circuit breaker, at-rest key encryption

**Deliberately not planned, and why:** in-app deposit/withdrawal execution, and any form of direct market/FIX connectivity that would make VOT Tradings itself an exchange or clearing venue. See the note at the end of this section.

**Operational notes if you're running this locally:**
- OANDA's `PlaceOrder` is fully implemented — if `OANDA_BASE_URL` in your `.env` points at `api-fxtrade.oanda.com` (live) instead of `api-fxpractice.oanda.com` (practice), any code path that calls it executes real trades with real money.
- Questrade refresh tokens rotate on every use and invalidate the previous one. This repo persists the rotated token to Postgres automatically; if you ever hand-paste a fresh token into `.env` after the `broker_oauth_tokens` table already has a row, the *stored* value wins on next boot (see the comment on `ensureAuth` in `internal/brokerage/questrade/client.go`) — clear that row first if you need the `.env` value to take precedence.

**Why VOT Tradings routes through Alpaca/OANDA/Questrade instead of self-clearing:** it would be architecturally possible to write a direct FIX gateway, an in-memory margin ledger, and Kafka-replicated settlement — but doing that makes VOT Tradings itself a broker-dealer/exchange/clearing venue, which isn't a coding problem. It requires FINRA/SEC broker-dealer registration (or NFA/CFTC registration for FX dealing), clearing-member capital requirements that typically run into the millions of dollars, and a compliance/legal program before a single order could legally route. Operating unregistered would be a serious regulatory violation, not a launch-and-fix-later gap. Routing through already-regulated brokers is what makes the rest of this project legal to build as a solo/small-team effort. A shadow in-memory balance ledger on top of brokers VOT Tradings doesn't custody funds at is also a correctness hazard on its own terms: the real money and the real margin state live at Alpaca/OANDA/Questrade, so any local ledger that thinks it's authoritative will eventually drift from reality and approve or block trades based on stale state. Any future performance layer should be a *read-through cache* of broker-reported balances, refreshed frequently, advisory only — the broker's own real-time check at order submission stays the actual authority.

---

## Architectural Overview

```
                      ┌────────────────────────────────────────┐
                      │          VOT Tradings Client            │
                      │              Web (React)                │
                      └───────────────────┬────────────────────┘
                                          │ HTTPS (JSON)
                      ┌───────────────────▼────────────────────┐
                      │       VOT Middleware Core (Go)         │
                      │  - Session & Portfolio Orchestration  │
                      │  - Unified Buying Power Compute Engine │
                      └───────────┬───────────────┬────────────┘
                                  │               │
         ┌────────────────────────┘               └────────────────────────┐
         ▼                                                                 ▼
┌─────────────────────────────────┐                               ┌─────────────────────────────────┐
│     Deep Learning Engine        │                               │    Real-Time Data Pipeline      │
│  - TensorFlow / PyTorch (Python)│                               │  - Redis Time-Series Cache      │
│  - LSTM Price Forecasters       │                               │  - Kafka Event Streaming Engine │
│  - RL Execution Optimization    │                               │  - Order Routing Layer          │
│         (planned)                │                               │         (partially built)        │
└─────────────────────────────────┘                               └─────────────────────────────────┘
         │                                                                 │
         └────────────────────────────────┬────────────────────────────────┘
                                          │ Encrypted FIX / mTLS REST
         ┌────────────────────────────────┼────────────────────────────────┐
         ▼                                ▼                                ▼
┌──────────────────┐             ┌──────────────────┐             ┌──────────────────┐
│   Alpaca API     │             │    OANDA API     │             │  Questrade API   │
│ (US Equities/FX) │             │  (Global Forex)  │             │(Canadian TSX/NEO)│
└──────────────────┘             └──────────────────┘             └──────────────────┘

```

The system is designed around three decoupling loops; the first is live, the other two are the target design:

1. **The Ingestion and Distribution Loop (Go + Redis):** Maintains non-blocking WebSocket connections to the liquidity provider sandboxes, handles message normalization, updates the low-latency cache, and broadcasts to client subscribers. *(Planned — today, quotes are fetched synchronously via REST per broker, not streamed.)*
2. **The Deep Learning Inference and Feedback Loop (Python + ONNX Runtime):** Continuously ingests streaming multi-asset updates to generate rolling predictions. It records execution outcomes to structural logs to facilitate self-directed learning. *(Planned — the `predictions` table exists in the schema; nothing writes to it yet.)*
3. **The Order Routing and State Loop (Go + PostgreSQL):** Manages cross-broker risk management checks, dual-currency balance tracking (USD/CAD friction mitigation), portfolio accounting, and lifecycle executions. *(Live — balance aggregation and PDT checks are implemented; full order lifecycle persistence is not yet wired to the HTTP API.)*

---

## Technology Stack and Core Infrastructure

* **System Core and API Gateway:** Go (Golang) — Leveraging high-performance concurrency semantics (goroutines and native channels) to process incoming market depth feeds without blocking transactional logic.
* **Web Client:** React + TypeScript (Vite) — see `web/`.
* **Deep Learning Runtime (planned):** Python / ONNX Runtime — Models are trained using TensorFlow and PyTorch, optimized, and compiled into ONNX formats for ultra-low latency inference steps inside the production pipeline.
* **In-Memory Cache and Message Broker:** Redis Stack — Houses sub-millisecond market state caches, active order books, and handles local event distribution via Redis Pub/Sub.
* **Primary Relational Ledger:** PostgreSQL — Stores immutable transaction histories, user authentication records, configuration variables, and audited performance state with full ACID compliance.

---

## Deep Learning Architecture and Self-Correction (Planned)

Everything in this section describes the target design. None of it is implemented yet — `services/dl_engine/` is currently empty. It's documented here so the schema and integration points (the `predictions` table, `DL_ENGINE_URL`/`DL_ENGINE_TIMEOUT_MS` config, the 45ms circuit-breaker requirement) make sense in context.

VOT Tradings is designed to integrate a multi-tiered Artificial Intelligence layer processing asymmetric, volatile structural data streams across disparate asset classes.

### 1. Predictive Models (LSTM and Transformers)

* **Forex Engine:** LSTM networks optimized for continuous, highly liquid time-series tracking. Features map rolling historical ticks, volatility metrics, and multi-currency cross-rate variations to project future price distributions across T+5, T+15, and T+60 minute windows.
* **Equities Engine:** Spatial-temporal Transformer architectures mapping equity price motions alongside cross-asset correlation matrices (e.g., Crude Oil movements vs. the Canadian Dollar and TSX Energy stocks).

### 2. Execution Engine (Deep Reinforcement Learning)

* An A3C Reinforcement Learning network intended to act as an intelligent execution router — not deciding *what* to buy, but *how*.
* Analyzes market depth (Level 2 order books) to split orders dynamically into optimized micro-tranches, minimizing slippage, spread penalties, and market footprint.

### 3. The Self-Correcting Feedback Loop (Continuous Learning)

The schema already supports this (`predictions` table in `internal/db/schema.sql`), even though nothing populates it yet:

* Every generated prediction is tagged with a unique `inference_id`.
* Upon trade execution or window expiration, actual market results are married to the prediction record.
* RMSE and Directional Accuracy scores are tracked. If a model's delta exceeds a volatility threshold, an automated offline pipeline is meant to trigger localized parameter fine-tuning via policy gradient methods.

---

## API Integration Matrix

### 1. Alpaca Broker API (US Equities)

* **Role:** Custody, clearing, and execution engine for US National Market System (NMS) securities.
* **Configuration:** REST v2 for order lifecycle management. Implemented in `internal/brokerage/alpaca/`.
* **Sandbox URL:** `https://paper-api.alpaca.markets`

### 2. Questrade API (Canadian Equities)

* **Role:** Exposure to TSX, TSX-V, and NEO listed equities.
* **Configuration:** OAuth 2.0 refresh-token flow, with rotation persisted to Postgres (see `internal/db/oauth_tokens.go`). Implemented in `internal/brokerage/questrade/`. Account and quote reads work; order placement is stubbed pending live-sandbox verification.
* **API URL:** `https://api01.questrade.com/v1/` (resolved dynamically per-account after auth)

### 3. OANDA v20 API (Institutional Forex)

* **Role:** Direct market access for global Spot Foreign Exchange currency pairs.
* **Configuration:** REST execution against the v20 API. Implemented in `internal/brokerage/oanda/`, including order placement.
* **Sandbox URL:** `https://api-fxpractice.oanda.com` (use `https://api-fxtrade.oanda.com` for a live account — see the operational note above)

---

## Project Structure

```text
vot-tradings/
├── cmd/
│   └── main_gateway/         # Go middleware API gateway entry point (implemented)
│                              # cmd/data_pipeline/ (planned, not yet scaffolded)
├── internal/                 # Private Go application core
│   ├── brokerage/            # Broker drivers: Alpaca, OANDA, Questrade
│   ├── cache/                 # Redis connection + key/channel helpers
│   ├── config/                 # Env/.env-based configuration loader
│   ├── db/                     # Postgres connection, schema, OAuth token store
│   ├── engine/                 # Unified balance aggregation, PDT compliance
│   ├── httpapi/                # HTTP handlers (health, balance, logo)
│   └── models/                  # Shared domain types
├── pkg/
│   └── logger/                  # Structured slog logger
├── web/                        # React + TypeScript + Tailwind client (implemented)
│   └── src/
│       ├── pages/               # Dashboard, Market, Intelligence, Trade, Settings
│       ├── context/             # PortfolioContext — shared poll of health + balance
│       ├── hooks/                # usePolling (live); useAlpacaStream/useOandaStream/
│       │                          # useInference (stubs — see Current Implementation Status)
│       ├── components/           # layout/, ui/, charts/, trading/
│       └── lib/                  # Typed gateway API client
├── assets/
│   └── logo.png                 # Canonical app logo — served by the gateway at /logo.png
├── services/
│   └── dl_engine/                # Deep Learning subsystem (planned, currently empty)
├── docker-compose.yml            # Local Postgres + Redis Stack
└── .env.example
```

---

## Installation and Local Environment Orchestration

### Prerequisites

* **Go:** `v1.25+` (pinned by `go.mod`; a transitive dependency requires this floor)
* **Node.js:** `v20+` (for `web/`)
* **Docker and Docker Compose**
* Sandbox credentials for whichever of Alpaca / OANDA / Questrade you want to exercise — the gateway runs fine with any subset configured; unconfigured brokers just show as unavailable in `/api/v1/balance`.

### Step 1: Clone and configure environment variables

```bash
git clone https://github.com/EfeObus/VOT-Tradings.git
cd VOT-Tradings
cp .env.example .env
```

`internal/config.Load()` reads `.env` automatically (via `godotenv`) on top of real process env vars, so no separate `source`/`export` step is needed. Fill in whichever broker credentials you have; leave the rest as placeholders.

Postgres in `docker-compose.yml` defaults to host port **5433**, not 5432 — this avoids colliding with a locally-installed Postgres server, which is a common setup on developer machines. Adjust `POSTGRES_PORT` in `.env` if you remap it.

### Step 2: Start Postgres and Redis

```bash
docker compose up -d postgres redis
```

### Step 3: Run the Go gateway

```bash
go mod download
go run cmd/main_gateway/main.go
```

This connects to Postgres (auto-applying `internal/db/schema.sql`) and Redis, constructs the broker clients, and serves the gateway at `http://localhost:8080`.

### Step 4: Run the web client

```bash
cd web
cp .env.example .env
npm install
npm run dev
```

Serves the dashboard at `http://localhost:5173`. Make sure `CORS_ALLOWED_ORIGINS` in the gateway's `.env` includes this origin (it does by default).

### Step 5 (not yet available): Deep Learning subsystem

```bash
cd services/dl_engine
# Not implemented yet — see Current Implementation Status above.
```

---

## Gateway HTTP API

| Method | Path                | Description |
|--------|----------------------|--------------|
| GET    | `/healthz`            | Pings Postgres and Redis; `200` if both are reachable, `503` otherwise. |
| GET    | `/api/v1/balance`      | Fans out to every configured broker concurrently, returns each broker's account/error plus its USD-converted equity (`equity_usd`, via `internal/engine.USDRate`), and a USD-unified rollup (`internal/engine.AggregateBalances`). |
| GET    | `/api/v1/quote?broker=&symbol=` | One-shot REST quote lookup against a single broker's `GetQuote`. A synchronous snapshot, not a stream — see `web/`'s Market data page. |
| GET    | `/logo.png`            | Serves `assets/logo.png` — the single canonical app logo; clients should reference this endpoint rather than bundling their own copy. |

CORS is allow-listed by exact origin via `CORS_ALLOWED_ORIGINS` (comma-separated) — origins not on the list get no CORS headers and are blocked by the browser as usual.

---

## Production Considerations and Security Guardrails

Constraints to enforce before moving past sandbox parameters into live production capital execution:

1. **Pattern Day Trader (PDT) Safeguards** — ✅ implemented. `internal/engine/pdt.go` tracks rolling 5-business-day day-trade windows for accounts below $25,000 USD equity.
2. **CAD/USD Dual-Currency Slippage Control** — ⚠️ not yet implemented. The README's original spec called for a 50bps cushion on CAD/USD-crossing executions; `internal/engine/balance.go` currently does straight rate conversion with no cushion.
3. **Fail-Safe Circuit Breakers** — ⚠️ not yet implemented. Requires the DL engine to exist first; the 45ms-timeout fallback to deterministic VWAP/TWAP execution is target design, not current behavior.
4. **Transport Layer Encryption** — ⚠️ partial. Outbound calls to Alpaca/OANDA/Questrade are HTTPS. At-rest AES-GCM-256 encryption for stored secrets/trading keys is not implemented — secrets currently live in `.env` (gitignored) and, for Questrade's rotating token, in a plaintext Postgres column.

---

## License

This architecture is distributed under the terms of the MIT License. Review the `LICENSE` file for precise legal allowances.
