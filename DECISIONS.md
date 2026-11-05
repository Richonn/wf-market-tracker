# DECISIONS

## API warframe.market — v1 bootstrap + v2 snapshots

**Decision: v1 for historical seed, v2 for ongoing snapshots.**

### Historical bootstrap (one-shot)
Use the v1 `/statistics` endpoint to seed the database with past price history before the
collector starts. This avoids starting from zero and gives immediately useful charts.

- `GET https://api.warframe.market/v1/items/{slug}/statistics` — completed sales history
- Run once per item in the watchlist at first launch
- v1 is deprecated but still alive; if it dies, we just lose the seed, not the system

### Ongoing collection (v2)
`https://api.warframe.market/v2/`

Regular snapshots of active orders via v2. We build our own price history from there,
with full control over granularity and metrics.

Retained v2 endpoints:
- `GET /items` — full item list
- `GET /item/{slug}` — item details
- `GET /orders/item/{slug}/top` — top 5 buy/sell (online sellers only)
- `GET /orders/item/{slug}` — all active orders (if deeper data is needed)

Required headers on every request:
- `Platform: pc`
- `User-Agent: wf-market-tracker/0.x (github.com/Richonn/wf-market-tracker)`

Rate limit: 3 req/sec max (Cloudflare 429 if exceeded). Throttling + exponential backoff
required in the collector.

---

## Project structure

```
cmd/collector/main.go                — entrypoint, reads DATABASE_URL, runs bootstrap then loop
internal/api/client.go               — HTTP client, 400ms/req throttle, exponential backoff (x5)
internal/api/v2.go                   — GetItems(), GetTopOrders()
internal/api/v1.go                   — GetStatistics() for historical seed
internal/collector/collector.go      — Bootstrap() + Run() + computeSnapshot()
internal/db/db.go                    — PostgreSQL connection pool
internal/db/queries.go               — UpsertItem, InsertSnapshot, InsertHistoricalSnapshot
internal/alert/ntfy.go               — ntfy Send()
internal/alert/discord.go            — Discord webhook Send()
internal/model/model.go              — Item, Order, Snapshot types
migrations/001_init.sql              — items + price_snapshots tables
```

No `watchlist.yaml` — tradeable item catalog is loaded dynamically from `GET /v2/items`
at each bootstrap, filtered by `bulkTradable: true`.

---

## Tailscale Funnel — public exposure

**Decision: Tailscale Funnel on port 8080 (api container only).**

Only the `api` container is on the `public` Docker network. Funnel terminates TLS and
proxies to `localhost:8080`. No ports are exposed directly on the NAS.

```sh
# Enable (run once on the NAS)
./scripts/funnel-enable.sh

# Emergency stop
./scripts/funnel-disable.sh
# or: tailscale funnel off
```

The Funnel URL is shown by `tailscale funnel status` after enabling.

Public URL: https://ugreen-nas.tailac801e.ts.net

The API container must bind port 8080 to the host loopback (`127.0.0.1:8080:8080`) so
Tailscale Funnel can reach it via `proxy http://127.0.0.1:8080`.

---

## Public interface — dedicated front vs exposing Grafana

**Decision: dedicated read-only API + static front, Grafana stays private.**

Exposing Grafana publicly would increase the attack surface (Grafana has had CVEs, requires
keeping auth config tight, and leaks version info). A purpose-built API exposes exactly
three endpoints, returns no stack traces, runs as non-root with `cap_drop: ALL` and
`read_only: true`, and connects to the DB with a read-only user.

The front is a single static HTML file — no build step, no framework, no dependencies
beyond TradingView Lightweight Charts loaded from CDN.

Tailscale Funnel exposes only the `api` container (port 8080) on the `public` Docker
network. All other containers (db, collector, grafana, uptime-kuma, dozzle) are on the
`internal` network only.

---

## Collector language — Go vs Python

**Decision: Go.**

- Single static binary, ~10 MB, no runtime dependency
- Low memory footprint (~5-10 MB at rest) — important on the NAS
- Native concurrency (goroutines) well suited for throttled HTTP loops
- Trivial Docker image (`COPY binary .`, distroless base)

---

## Database — TimescaleDB vs InfluxDB vs PostgreSQL

**Decision: PostgreSQL.**

Expected volume: ~30 items × 48 snapshots/day ≈ 1,500 rows/day (~550k/year). PostgreSQL
handles this easily with no extra complexity.

- Lower RAM footprint than TimescaleDB (~50-100 MB vs ~200 MB) or InfluxDB (~300-500 MB)
- Standard SQL, native Grafana datasource
- Upgrade path to TimescaleDB is a single `CREATE EXTENSION` if needed later

---

## Collection frequency

**Decision: every 45 minutes.**

We track all tradeable items (filtered by `bulkTradable: true` from `GET /items`), which
amounts to roughly 1,500-2,000 items. At 3 req/sec that's ~11 minutes of requests per
cycle, leaving plenty of idle time before the next run.

A 15-min window would be too tight. 45 min is comfortable and still gives 32 data points
per day per item — enough for meaningful charts and trend detection.

---

## Alert channels — ntfy vs Discord vs both

**Decision: both ntfy and Discord.**

- **ntfy** — push notifications on mobile, no account required, self-hostable
- **Discord** — existing bot, good for richer messages and history

Both are independent delivery paths, which adds resilience (if one is down, the other fires).

---

## Infrastructure — Docker Compose vs K3s

**Decision: Docker Compose on the NAS.**

Single node, no real benefit from K3s in terms of self-healing or scaling. Limited RAM on
the NAS. Docker Compose is sufficient, simpler to operate and debug.

Single collector instance (no scaling) to avoid duplicate snapshots and stay within rate
limits.
