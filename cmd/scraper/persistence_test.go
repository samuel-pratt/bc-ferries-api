package scraper

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"os"
	"testing"

	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
)

func TestPartialScrapePreservesCachedSailings(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("run scripts/verify.sh for disposable database integration tests")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	var name string
	if err = admin.QueryRow("SELECT current_database()").Scan(&name); err != nil || name != "release_test" {
		t.Fatalf("requires isolated release_test database: %q %v", name, err)
	}
	// Other packages test legacy migrations concurrently in the public schema.
	// Keep this fixture and every production query under test in its own schema.
	if _, err = admin.Exec("CREATE SCHEMA scraper_regression"); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", "scraper_regression")
	u.RawQuery = q.Encode()
	conn, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	previous := db.Conn
	db.Conn = conn
	defer func() { db.Conn = previous }()
	if _, err = conn.Exec(`CREATE TABLE capacity_routes (
route_code VARCHAR(6) PRIMARY KEY, from_terminal_code VARCHAR(3) NOT NULL,
to_terminal_code VARCHAR(3) NOT NULL, sailing_duration VARCHAR(7) NOT NULL, sailings JSONB NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if err = db.EnsureSchema(); err != nil {
		t.Fatal(err)
	}
	good := `<table class="detail-departure-table">
<tr class="mobile-friendly-row"><td>9:00 am Vessel</td><td>Upcoming</td></tr>
<tr class="mobile-friendly-row"><td>10:00 am Vessel</td><td>Cancelled</td></tr></table>`
	ScrapeCapacityRoute(document(good), "TSA", "SWB")
	before := db.GetCapacitySailings()
	if len(before) != 1 || len(before[0].Sailings) != 2 || before[0].DataStatus != "available" {
		t.Fatalf("good fixture not stored: %#v", before)
	}
	broken := `<table class="detail-departure-table">
<tr class="mobile-friendly-row"><td>9:00 am Vessel</td><td>Upcoming</td></tr>
<tr class="mobile-friendly-row"><td>Unknown sailing</td><td>Cancelled</td></tr></table>`
	ScrapeCapacityRoute(document(broken), "TSA", "SWB")
	after := db.GetCapacitySailings()
	if len(after) != 1 || len(after[0].Sailings) != 2 || after[0].DataStatus != "unavailable" ||
		!after[0].ObservedAt.Equal(before[0].ObservedAt) || after[0].Sailings[1].SailingStatus != "cancelled" {
		t.Fatalf("partial scrape replaced cached sailings or observation: %#v", after)
	}
	var diagnostic models.ConditionsSnapshot
	if err = json.Unmarshal(db.Snapshots("conditions")["TSASWB"], &diagnostic); err != nil || diagnostic.DataStatus != "unavailable" || len(diagnostic.Warnings) == 0 {
		t.Fatalf("missing explicit parse failure: %#v, %v", diagnostic, err)
	}
	ScrapeCapacityRoute(document(good), "TSA", "SWB")
	recovered := db.GetCapacitySailings()
	if len(recovered) != 1 || recovered[0].DataStatus != "available" || len(recovered[0].Sailings) != 2 || !recovered[0].ObservedAt.After(before[0].ObservedAt) {
		t.Fatalf("fresh scrape did not recover: %#v", recovered)
	}
}
