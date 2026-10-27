package scraper

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"github.com/samuel-pratt/bc-ferries-api/cmd/router"
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
	fixture, err := os.ReadFile("testdata/conditions_no_service.html")
	if err != nil {
		t.Fatal(err)
	}
	ScrapeCapacityRoute(document(string(fixture)), "TSA", "SWB")
	noService := db.GetCapacitySailings()
	if len(noService) != 1 || noService[0].DataStatus != "no-service" || noService[0].Sailings == nil || len(noService[0].Sailings) != 0 ||
		noService[0].ServiceDate == "" || !noService[0].ObservedAt.After(recovered[0].ObservedAt) {
		t.Fatalf("explicit no-service day failed to clear old sailings: %#v", noService)
	}
	// Even a fleet with no scheduled sailings must return the normal response,
	// not the legacy data-down string or a database-readiness failure.
	res := httptest.NewRecorder()
	router.GetCapacitySailings(res, httptest.NewRequest("GET", "/v2/capacity/", nil), nil)
	var response router.CapacityResponse
	if err = json.Unmarshal(res.Body.Bytes(), &response); err != nil || len(response.Routes) != 1 || response.Routes[0].DataStatus != "no-service" || response.Meta.Stale {
		t.Fatalf("no-service HTTP response invalid: %s (%v)", res.Body.String(), err)
	}
	res = httptest.NewRecorder()
	router.ReadyCheck(res, httptest.NewRequest("GET", "/readyz/", nil), nil)
	if res.Code != 200 {
		t.Fatalf("no-service snapshot failed readiness: %s", res.Body.String())
	}
	ScrapeCapacityRoute(document(`<h1>Service unavailable</h1>`), "TSA", "SWB")
	failed := db.GetCapacitySailings()
	if len(failed) != 1 || failed[0].DataStatus != "unavailable" || !failed[0].ObservedAt.Equal(noService[0].ObservedAt) {
		t.Fatalf("later failure falsely refreshed no-service observation: %#v", failed)
	}
	res = httptest.NewRecorder()
	router.ReadyCheck(res, httptest.NewRequest("GET", "/readyz/", nil), nil)
	if res.Code != 503 {
		t.Fatal("unavailable empty snapshot passed readiness")
	}
	ScrapeCapacityRoute(document(good), "TSA", "SWB")
	resumed := db.GetCapacitySailings()
	if len(resumed) != 1 || resumed[0].DataStatus != "available" || len(resumed[0].Sailings) != 2 {
		t.Fatalf("service failed to resume: %#v", resumed)
	}
}
