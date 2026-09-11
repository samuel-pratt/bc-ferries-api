CREATE TABLE capacity_routes (
    route_code VARCHAR(6) PRIMARY KEY,
    from_terminal_code VARCHAR(3) NOT NULL,
    to_terminal_code VARCHAR(3) NOT NULL,
    sailing_duration VARCHAR(7) NOT NULL,
    sailings JSONB NOT NULL
);

CREATE TABLE non_capacity_routes (
    route_code VARCHAR(6) PRIMARY KEY,
    from_terminal_code VARCHAR(3) NOT NULL,
    to_terminal_code VARCHAR(3) NOT NULL,
    sailing_duration VARCHAR(7) NOT NULL,
    sailings JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS route_operational_data (
 kind TEXT NOT NULL CHECK (kind IN ('capacity','noncapacity')),
 route_code VARCHAR(6) NOT NULL,
 payload JSONB NOT NULL DEFAULT '{}',
 PRIMARY KEY(kind,route_code)
);
