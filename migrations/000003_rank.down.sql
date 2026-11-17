ALTER TABLE price_snapshots_daily DROP CONSTRAINT price_snapshots_daily_pkey;
ALTER TABLE price_snapshots_daily ADD PRIMARY KEY (item_slug, day);
ALTER TABLE price_snapshots_daily DROP COLUMN rank;

ALTER TABLE price_snapshots_hourly DROP CONSTRAINT price_snapshots_hourly_pkey;
ALTER TABLE price_snapshots_hourly ADD PRIMARY KEY (item_slug, hour);
ALTER TABLE price_snapshots_hourly DROP COLUMN rank;

ALTER TABLE price_snapshots DROP CONSTRAINT price_snapshots_item_slug_collected_at_rank_key;
ALTER TABLE price_snapshots ADD CONSTRAINT price_snapshots_item_slug_collected_at_key UNIQUE (item_slug, collected_at);
ALTER TABLE price_snapshots DROP COLUMN rank;
