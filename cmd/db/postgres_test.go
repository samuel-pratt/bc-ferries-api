package db

import (
	"database/sql"
	"net/url"
	"os"
	"testing"
)

// Only runs against the disposable release database created by compose.test.yml.
func TestMigrationPreservesLegacyAndAtomicWrites(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL using scripts/verify.sh")
	}
	var err error
	Conn, err = sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer Conn.Close()
	var name string
	if err = Conn.QueryRow("SELECT current_database()").Scan(&name); err != nil || name != "release_test" {
		t.Fatalf("requires isolated release_test database: %q %v", name, err)
	}
	for _, table := range []string{"capacity_routes", "non_capacity_routes"} {
		if _, err = Conn.Exec("CREATE TABLE " + table + " (route_code VARCHAR(6) PRIMARY KEY,from_terminal_code VARCHAR(3) NOT NULL,to_terminal_code VARCHAR(3) NOT NULL,sailing_duration VARCHAR(7) NOT NULL,sailings JSONB NOT NULL)"); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err = EnsureSchema(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = Conn.Exec(`CREATE ROLE release_app LOGIN PASSWORD 'disposable-test-only' NOSUPERUSER; GRANT USAGE,CREATE ON SCHEMA public TO release_app; GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA public TO release_app`); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	u.User = url.UserPassword("release_app", "disposable-test-only")
	Conn, err = sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer Conn.Close()
	if err = EnsureSchema(); err != nil {
		t.Fatalf("non-superuser cannot restart: %v", err)
	}
	if err = SaveRoute("capacity", "TSASWB", "TSA", "SWB", "01:35", []byte(`[{"time":"8:00 am"}]`), []byte(`{"observedAt":"2026-09-11T12:00:00Z"}`)); err != nil {
		t.Fatal(err)
	}
	var code, from, to, duration string
	var sailings []byte
	if err = Conn.QueryRow("SELECT * FROM capacity_routes").Scan(&code, &from, &to, &duration, &sailings); err != nil {
		t.Fatalf("old binary cannot read migrated schema: %v", err)
	}
	if code != "TSASWB" || len(GetCapacitySailings()) != 1 {
		t.Fatal("route round trip failed")
	}
	if err = SaveRoute("capacity", "TSASWB", "TSA", "SWB", "02:00", []byte(`[]`), []byte(`invalid json`)); err == nil {
		t.Fatal("invalid metadata should fail")
	}
	if err = Conn.QueryRow("SELECT sailing_duration FROM capacity_routes").Scan(&duration); err != nil || duration != "01:35" {
		t.Fatalf("partial write escaped transaction: %s %v", duration, err)
	}
}
