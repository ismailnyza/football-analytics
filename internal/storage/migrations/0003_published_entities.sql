CREATE TABLE IF NOT EXISTS published_entities (
    id INTEGER PRIMARY KEY,
    source_code TEXT NOT NULL,
    external_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    resolved_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    attributes_json TEXT NOT NULL,
    validation_json TEXT NOT NULL,
    published_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_published_entities_source
    ON published_entities(source_code, entity_type, published_at);
