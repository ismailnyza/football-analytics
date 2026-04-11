ALTER TABLE raw_payloads ADD COLUMN processed_at TEXT;

CREATE INDEX IF NOT EXISTS idx_raw_payloads_source_processed
    ON raw_payloads(source_id, processed_at, fetched_at);
