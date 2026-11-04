CREATE TABLE IF NOT EXISTS price_snapshots_hourly (
    item_slug       TEXT        NOT NULL REFERENCES items(slug),
    hour            TIMESTAMPTZ NOT NULL,
    sell_min        INT         NOT NULL DEFAULT 0,
    sell_median     INT         NOT NULL DEFAULT 0,
    buy_max         INT         NOT NULL DEFAULT 0,
    avg_sell_orders INT         NOT NULL DEFAULT 0,
    avg_buy_orders  INT         NOT NULL DEFAULT 0,
    sample_count    INT         NOT NULL DEFAULT 0,
    PRIMARY KEY (item_slug, hour)
);

CREATE TABLE IF NOT EXISTS price_snapshots_daily (
    item_slug       TEXT        NOT NULL REFERENCES items(slug),
    day             DATE        NOT NULL,
    sell_min        INT         NOT NULL DEFAULT 0,
    sell_median     INT         NOT NULL DEFAULT 0,
    buy_max         INT         NOT NULL DEFAULT 0,
    avg_sell_orders INT         NOT NULL DEFAULT 0,
    avg_buy_orders  INT         NOT NULL DEFAULT 0,
    sample_count    INT         NOT NULL DEFAULT 0,
    PRIMARY KEY (item_slug, day)
);

CREATE INDEX IF NOT EXISTS idx_hourly_item_time ON price_snapshots_hourly (item_slug, hour DESC);
CREATE INDEX IF NOT EXISTS idx_daily_item_time  ON price_snapshots_daily  (item_slug, day  DESC);
