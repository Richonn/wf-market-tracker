-- Add rank to raw snapshots
ALTER TABLE price_snapshots ADD COLUMN rank INT NOT NULL DEFAULT 0;
ALTER TABLE price_snapshots DROP CONSTRAINT IF EXISTS price_snapshots_item_slug_collected_at_key;
ALTER TABLE price_snapshots ADD CONSTRAINT price_snapshots_item_slug_collected_at_rank_key UNIQUE (item_slug, collected_at, rank);

-- Add rank to hourly aggregates
ALTER TABLE price_snapshots_hourly ADD COLUMN rank INT NOT NULL DEFAULT 0;
ALTER TABLE price_snapshots_hourly DROP CONSTRAINT price_snapshots_hourly_pkey;
ALTER TABLE price_snapshots_hourly ADD PRIMARY KEY (item_slug, hour, rank);

-- Add rank to daily aggregates
ALTER TABLE price_snapshots_daily ADD COLUMN rank INT NOT NULL DEFAULT 0;
ALTER TABLE price_snapshots_daily DROP CONSTRAINT price_snapshots_daily_pkey;
ALTER TABLE price_snapshots_daily ADD PRIMARY KEY (item_slug, day, rank);
