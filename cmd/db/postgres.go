package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/samuel-pratt/bc-ferries-api/cmd/config"
)

var Conn *sql.DB

func Init() {
	var err error
	Conn, err = sql.Open("postgres", config.DB.URL)
	if err != nil {
		panic(err)
	}
	Conn.SetMaxOpenConns(10)
	Conn.SetMaxIdleConns(5)
	Conn.SetConnMaxLifetime(30 * time.Minute)
}

// Keep the original route tables unchanged: legacy binaries SELECT * into five fields.
func EnsureSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	tx, err := Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, statement := range []string{
		"SET LOCAL lock_timeout = '5s'",
		"SET LOCAL statement_timeout = '15s'",
		`CREATE TABLE IF NOT EXISTS route_operational_data (
    kind TEXT NOT NULL CHECK (kind IN ('capacity','noncapacity')),
    route_code VARCHAR(6) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    PRIMARY KEY (kind,route_code)
  )`,
		`CREATE TABLE IF NOT EXISTS source_snapshots (
    kind TEXT NOT NULL,
    source_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    PRIMARY KEY (kind,source_key)
  )`,
	} {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("ensure schema: %w", err)
		}
	}
	return tx.Commit()
}

// Write the legacy data and its observation metadata in the same transaction.
func SaveRoute(kind, code, from, to, duration string, sailings, operational []byte) error {
	table := "capacity_routes"
	if kind == "noncapacity" {
		table = "non_capacity_routes"
	} else if kind != "capacity" {
		return fmt.Errorf("invalid route kind")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "INSERT INTO "+table+` (route_code,from_terminal_code,to_terminal_code,sailing_duration,sailings)
 VALUES ($1,$2,$3,$4,$5) ON CONFLICT(route_code) DO UPDATE SET
 from_terminal_code=EXCLUDED.from_terminal_code,to_terminal_code=EXCLUDED.to_terminal_code,
 sailing_duration=EXCLUDED.sailing_duration,sailings=EXCLUDED.sailings`, code, from, to, duration, sailings)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO route_operational_data(kind,route_code,payload) VALUES($1,$2,$3)
 ON CONFLICT(kind,route_code) DO UPDATE SET payload=EXCLUDED.payload`, kind, code, operational)
	if err != nil {
		return err
	}
	return tx.Commit()
}
