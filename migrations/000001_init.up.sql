CREATE TABLE IF NOT EXISTS items (
    id   TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS price_snapshots (
    id                  BIGSERIAL PRIMARY KEY,
    item_slug           TEXT        NOT NULL REFERENCES items(slug),
    collected_at        TIMESTAMPTZ NOT NULL,
    sell_min            INT         NOT NULL DEFAULT 0,
    sell_median         INT         NOT NULL DEFAULT 0,
    buy_max             INT         NOT NULL DEFAULT 0,
    active_sell_orders  INT         NOT NULL DEFAULT 0,
    active_buy_orders   INT         NOT NULL DEFAULT 0,
    UNIQUE (item_slug, collected_at)
);

CREATE INDEX IF NOT EXISTS idx_snapshots_item_time ON price_snapshots (item_slug, collected_at DESC);
