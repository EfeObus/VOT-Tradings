# VOT Tradings — Web Client

React + TypeScript + Tailwind v4 client for the VOT Tradings gateway (`cmd/main_gateway`). Five-page shell matching the platform's target IA — see [Current scope](#current-scope) for what's real vs. pending backend.

## Setup

```bash
cd web
cp .env.example .env   # point at your running gateway if not the default
npm install
npm run dev
```

Requires the Go gateway running and reachable (default `http://localhost:8080`), with `CORS_ALLOWED_ORIGINS` in the gateway's `.env` including this app's origin (default `http://localhost:5173`).

## Scripts

- `npm run dev` — dev server with HMR
- `npm run build` — type-check (`tsc -b`) then production build to `dist/`
- `npm run preview` — serve the production build locally
- `npm run lint` — Oxlint

## Brand tokens

Defined in `src/index.css` via Tailwind v4's `@theme`: `--color-canvas` (#0B0E14), `--color-surface` (#1A1D26), `--color-elevated` (#242936), `--color-fg` / `--color-fg-muted`, `--color-bull` (#0ECB81), `--color-bear` (#F6465D), `--color-accent` (#3B82F6). Generates standard Tailwind utilities (`bg-canvas`, `text-bull`, etc.) — no separate `tailwind.config.js` needed under v4's CSS-first config.

## Structure

- `src/pages/` — `Dashboard`, `Market`, `Intelligence`, `Trade`, `Settings`, routed in `App.tsx`
- `src/context/PortfolioContext.tsx` — single shared poll of `/healthz` + `/api/v1/balance`, consumed by any page via `usePortfolio()`
- `src/hooks/usePolling.ts` — the underlying fixed-interval polling hook
- `src/hooks/useAlpacaStream.ts`, `useOandaStream.ts`, `useInference.ts` — **stubs**. Each always returns `{ connected: false, reason: 'not_implemented' }`; they exist as the landing spot for real streaming/inference work later, not as working data sources today
- `src/components/layout/` — `NavBar`, `AppLayout`
- `src/components/ui/` — `Card`, `StatTile`, `StatusBadge`, `NotConnected` (the "this feature has no backend yet" panel used throughout Market/Intelligence/Trade)
- `src/components/charts/AllocationDonut.tsx` — per-broker USD allocation, computed from the gateway's `equity_usd` field (never a client-side guess at FX conversion)
- `src/components/trading/BrokerAccountCard.tsx` — per-broker connected/error card
- `src/lib/api.ts`, `src/lib/types.ts` — typed gateway client; keep in sync with `internal/httpapi/httpapi.go`

## Current scope

| Page | Status |
|---|---|
| `/dashboard` | Real — NAV, cross-border split, allocation chart, all from `/api/v1/balance` |
| `/settings` | Real — broker connectivity audit from the same endpoint |
| `/market/:symbol` | Layout only — candlesticks/L2/indicators all show `NotConnected`; needs a streaming backend |
| `/intelligence` | Layout only — needs the Python DL engine (`services/dl_engine`, currently empty) |
| `/trade` | Cash-by-currency is real; the order ticket and PDT shield show `NotConnected` — needs an order-execution HTTP endpoint |

Nothing in this app fabricates data for a feature the backend doesn't support — see each page's `NotConnected` panels for exactly what's missing and where.
