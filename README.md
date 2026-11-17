# wf-market-tracker

Track Warframe item prices from [warframe.market](https://warframe.market) — live order snapshots, price history charts, 24h/7d trends.

**Live site:** https://ugreen-nas.tailac801e.ts.net

## What it does

- Collects the top buy/sell orders for every tradeable item every 45 minutes
- Seeds historical data from warframe.market statistics on first run
- Exposes a read-only JSON API + static web frontend
- Displays price charts (sell median, sell min, buy max) with volume and time range filters
- Links directly to warframe.market listings and the official wiki for each item

## Stack

| Component | Tech |
|-----------|------|
| Collector | Go — polls warframe.market v2 API |
| Database  | PostgreSQL 17 |
| API       | Go — read-only HTTP server |
| Frontend  | Vanilla HTML/JS + [Lightweight Charts](https://github.com/tradingview/lightweight-charts) |
| Infra     | Docker Compose on a UGREEN NAS |
| Exposure  | Tailscale Funnel (HTTPS, no open ports) |

## Running locally

**Prerequisites:** Docker, Docker Compose

```sh
git clone https://github.com/Richonn/wf-market-tracker
cd wf-market-tracker

cp .env.example .env
# edit .env — set POSTGRES_PASSWORD and WFMT_RO_PASSWORD at minimum

docker-compose up -d
```

The API will be available at http://localhost:8080.

Bootstrap (first run) seeds historical data for all items — this takes ~15 minutes due to API rate limits.

## API

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `GET /items` | All items with latest price and 24h/7d change |
| `GET /items/{slug}/candles?interval=1h\|1d` | Price candles for one item |
| `GET /movers` | Top 20 items by 24h price change |

## Project structure

```
cmd/collector/        — collector entrypoint
cmd/api/              — API server entrypoint
internal/api/         — warframe.market API client (v1 + v2)
internal/collector/   — snapshot collection logic
internal/db/          — PostgreSQL queries
internal/alert/       — ntfy + Discord alert delivery
internal/model/       — shared types
migrations/           — SQL migrations (embedded)
web/                  — static frontend
grafana/              — Grafana provisioning
postgres/             — DB init scripts
scripts/              — Tailscale Funnel helpers
```

## Configuration

All config via environment variables (see `.env.example`):

| Variable | Required | Description |
|----------|----------|-------------|
| `POSTGRES_PASSWORD` | yes | Postgres wfmt user password |
| `WFMT_RO_PASSWORD` | yes | Postgres wfmt_ro read-only user password |
| `HEALTHCHECKS_URL` | no | healthchecks.io ping URL |
| `NTFY_URL` | no | ntfy endpoint for price alerts |
| `NTFY_TOKEN` | no | ntfy auth token |
| `DISCORD_WEBHOOK_URL` | no | Discord webhook for price alerts |
| `GRAFANA_PASSWORD` | no | Grafana admin password |

## Disclaimer

Not affiliated with Digital Extremes or warframe.market. Data sourced from the warframe.market public API — use responsibly and respect rate limits.
