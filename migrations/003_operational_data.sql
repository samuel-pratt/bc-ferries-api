-- Does not change legacy table shapes; old binaries remain usable after rollback.
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '15s';
CREATE TABLE IF NOT EXISTS route_operational_data (
    kind TEXT NOT NULL CHECK (kind IN ('capacity', 'noncapacity')),
    route_code VARCHAR(6) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    PRIMARY KEY (kind, route_code)
);
COMMIT;
