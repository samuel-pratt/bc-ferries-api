-- Additive: legacy route tables retain their original five columns.
CREATE TABLE IF NOT EXISTS source_snapshots (
    kind TEXT NOT NULL,
    source_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    PRIMARY KEY (kind, source_key)
);
