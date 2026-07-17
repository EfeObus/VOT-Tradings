# VOT Tradings — Web Client

React + TypeScript + Vite dashboard for the VOT Tradings gateway (`cmd/main_gateway`). Read-only for now: system health and the unified cross-broker balance view.

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

## Structure

- `src/lib/api.ts`, `src/lib/types.ts` — typed client for the gateway's JSON API; keep in sync with `internal/httpapi/httpapi.go`
- `src/components/` — `Header` (brand logo, served by the gateway at `/logo.png`), `Dashboard`, `StatusPill`, `StatTile`, `BrokerCard`
- `src/hooks/usePolling.ts` — fixed-interval polling hook used for health and balance data

## Current scope

Only what the gateway currently exposes: `/healthz` and `/api/v1/balance`. There's no order entry or position management UI yet — the gateway itself doesn't expose those endpoints (see root `README.md` for what's implemented vs. planned).
