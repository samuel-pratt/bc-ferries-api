package scraper

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"github.com/samuel-pratt/bc-ferries-api/cmd/staticdata"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var auditTime = time.Date(2026, 9, 11, 21, 30, 0, 0, time.UTC)

func document(s string) *goquery.Document {
	d, e := goquery.NewDocumentFromReader(strings.NewReader(s))
	if e != nil {
		panic(e)
	}
	return d
}
func noDetails(string) (*goquery.Document, error) { return nil, fmt.Errorf("detail unavailable") }

// Reduced from the HSB-NAN, SWB-SGI and DUK-TSA live markup captured Sep 11.
func TestLiveConditionEdgeCases(t *testing.T) {
	d := document(`<table class="detail-departure-table"><tbody>
 <tr class="mobile-friendly-row"><td><p>6:20 pm <a>Queen of Cowichan</a></p></td><td><p>6:20 pm Queen of Cowichan</p><div>Upcoming</div></td></tr>
 <tr class="mobile-friendly-row"><td><p>6:15 am (Tomorrow) <a>Queen of Oak Bay</a></p></td><td>Upcoming</td></tr>
 <tr class="mobile-friendly-row"><td><p>5:00 am <a>Queen of Cumberland</a></p></td><td><div class="percentage-full"><div>...</div></div></td></tr>
 <tr class="sgi-row"><td><span id="stop-names">via Mayne Island to Pender Island</span></td></tr>
 <tr class="mobile-friendly-row"><td><p>11:45 pm Departed 11:52 pm Vessel</p></td><td><div class="cc-message-updates">ETA : 1:30 am</div><p class="text-red">We are loading as many vehicles as possible</p></td></tr>
 <tr class="mobile-friendly-row"><td><p>3:15 pm Vessel</p></td><td><p>3:15 pm Vessel</p><div class="cc-message-updates"><span>Full</span></div></td></tr>
 </tbody></table>`)
	c := parseConditions(d, "HSB", "NAN", auditTime, noDetails)
	if len(c.Sailings) != 5 {
		t.Fatal(c)
	}
	a, b, u, m, f := c.Sailings[0], c.Sailings[1], c.Sailings[2], c.Sailings[3], c.Sailings[4]
	if a.DepartureTime != "6:20 pm" || a.VesselName != "Queen of Cowichan" || a.Status != "scheduled" || a.CapacityKnown == nil || *a.CapacityKnown {
		t.Fatal(a)
	}
	if b.ServiceDate != "2026-09-12" || b.ScheduledTime != "6:15 am" {
		t.Fatal(b)
	}
	if u.Status != "unknown" || u.SailingStatus != "unknown" || u.DepartureTime != "5:00 am" || len(u.Stops) != 1 {
		t.Fatal(u)
	}
	if m.DepartureTime != "11:52 pm" || m.ScheduledTime != "11:45 pm" || m.ActualDepartureTime != "11:52 pm" || m.ArrivalDate != "2026-09-12" || m.VesselStatus == "" {
		t.Fatal(m)
	}
	if f.Fill != 100 || f.CarFill != 100 || f.OversizeFill != 100 || !*f.CapacityKnown {
		t.Fatal(f)
	}
}
func TestMalformedCapacityIsUnknown(t *testing.T) {
	for _, v := range []string{"", "N/A", "101%", "-1%", "busy"} {
		if _, ok := parseFull(v); ok {
			t.Fatalf("accepted %q", v)
		}
	}
	d := document(`<table class="detail-departure-table"><tr class="mobile-friendly-row"><td>7:00 pm Vessel</td><td><span class="cc-vessel-percent-full">40%</span><a href="/sailing-availability?x=1">Details</a></td></tr></table>`)
	c := parseConditions(d, "A", "B", auditTime, noDetails)
	s := c.Sailings[0]
	if c.DataStatus != "available" || s.Fill != 60 || !*s.CapacityKnown || *s.CarCapacityKnown || len(c.Warnings) == 0 {
		t.Fatal(c)
	}
	c = parseConditions(d, "A", "B", auditTime, func(string) (*goquery.Document, error) {
		return document(`<p class="vehicle-icon-text">N/A</p><p class="vehicle-icon-text">Full</p><p class="vehicle-icon-text">50%</p>`), nil
	})
	s = c.Sailings[0]
	if c.DataStatus != "available" || s.Fill != 60 || s.CarFill != 100 || s.OversizeFill != 50 || len(c.Warnings) == 0 {
		t.Fatal(c)
	}
}

func TestPartialConditionsCannotReplaceCompleteRoute(t *testing.T) {
	good := `<tr class="mobile-friendly-row"><td>9:00 am Vessel</td><td>Upcoming</td></tr>`
	for name, broken := range map[string]string{
		"missing time": `<tr class="mobile-friendly-row"><td>Unknown sailing</td><td>Cancelled</td></tr>`,
		"invalid time": `<tr class="mobile-friendly-row"><td>29:99 pm Vessel</td><td>Upcoming</td></tr>`,
		"missing cell": `<tr class="mobile-friendly-row"><td>10:00 am Vessel</td></tr>`,
	} {
		t.Run(name, func(t *testing.T) {
			// An error in any row must survive subsequent successfully parsed rows.
			for _, rows := range []string{good + broken, broken + good} {
				c := parseConditions(document(`<table class="detail-departure-table">`+rows+`</table>`), "TSA", "SWB", auditTime, noDetails)
				if c.DataStatus != "unavailable" || len(c.Sailings) != 1 || len(c.Warnings) == 0 {
					t.Fatalf("partial conditions accepted as complete: %#v", c)
				}
			}
		})
	}
}

func TestFetchRejectsHTTPFailuresAndTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(time.Second):
			}
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`<p class="vehicle-icon-text">0%</p>`))
	}))
	defer server.Close()
	previous := sourceClient
	sourceClient = &http.Client{Timeout: 20 * time.Millisecond}
	defer func() { sourceClient = previous }()
	for _, p := range []string{"/error", "/slow"} {
		if d, e := fetchDocument(server.URL + p); e == nil || d != nil {
			t.Fatalf("accepted HTTP failure %s: %v", p, e)
		}
	}
}
func TestDateRestrictions(t *testing.T) {
	for note, want := range map[string]bool{"Only on Jun 5, 19, Jul 3, 17, 31, Aug 14, 28, Sep 11 & 25": true, "Only on: 18 Sep": false, "Only on: 11 Sep": true, "Not Available on: 11 Sep": false, "Except Dec 25, Jan 1": true, "Foot passengers only": true, "Dangerous goods only": false} {
		got, known := permitted(note, pacific(auditTime))
		if !known || got != want {
			t.Errorf("%q: %v %v", note, got, known)
		}
	}
	if _, known := permitted("Only on an unknown date", pacific(auditTime)); known {
		t.Fatal("unknown exception accepted")
	}
}
func TestScheduleNoServiceAndReturnIsolation(t *testing.T) {
	d := document(`<h2 id="OnwardSchedule">A to B</h2><table class="table-seasonal-schedule"><thead><tr data-schedule-day="Mondays"><th>Monday</th></tr></thead><tbody><tr class="schedule-table-row"><td></td><td>8:00 am</td><td>9:00 am</td><td>1h</td></tr></tbody></table><h2 id="ReturnSchedule">B to A</h2><table class="table-seasonal-schedule"><thead><tr data-schedule-day="Fridays"><th>Friday</th></tr></thead><tbody><tr class="schedule-table-row"><td></td><td>10:00 am</td><td>11:00 am</td><td>1h</td></tr></tbody></table>`)
	r, ok := parseSchedule(d, "A", "B", false, auditTime)
	if !ok || r.DataStatus != "no-service" || len(r.Sailings) != 0 {
		t.Fatal(r, ok)
	}
	if _, ok = parseSchedule(document(`<h2 id="OnwardSchedule">A to B</h2><p>Schedules unavailable</p><h2 id="ReturnSchedule">B to A</h2><table id="dailyScheduleTableOnward"></table>`), "A", "B", true, auditTime); ok {
		t.Fatal("unavailable treated as no service")
	}
}
func TestScheduleSailingDetails(t *testing.T) {
	d := document(`<table id="dailyScheduleTableOnward"><tbody><tr class="schedule-table-row"><td></td><td>10:00 pm<p class="text-black">Foot passengers only</p></td><td>5:00 am</td><td>07:00</td><td><div class="progtrckr"><ul><li>Transfer Mayne Island</li></ul></div></td></tr></tbody></table>`)
	r, ok := parseSchedule(d, "A", "B", true, auditTime)
	if !ok {
		t.Fatal(r)
	}
	s := r.Sailings[0]
	if s.Duration != "07:00" || s.ArrivalDate != "2026-09-12" || len(s.Stops) != 1 || s.VesselStatus != "Foot passengers only" {
		t.Fatal(s)
	}
}
func TestConfiguredCoverage(t *testing.T) {
	p := staticdata.SchedulePairs()
	if len(p) != 118 || len(staticdata.CapacityPairs()) != 13 || len(staticdata.ConditionPairs()) != 110 {
		t.Fatal(len(p), len(staticdata.ConditionPairs()))
	}
	all := strings.Join(p, " ")
	if !strings.Contains(all, "COR-HRB") || !strings.Contains(all, "TSA-NAN") {
		t.Fatal(all)
	}
	raw := 0
	for _, v := range staticdata.GetNonCapacityDestinationTerminals() {
		raw += len(v)
	}
	if raw != len(p) {
		t.Fatal("duplicate pairs")
	}
}

func TestSavedNorthernAndGambierMarkup(t *testing.T) {
	b, e := os.ReadFile("testdata/northern_seasonal.html")
	if e != nil {
		t.Fatal(e)
	}
	r, ok := parseSchedule(document(string(b)), "PBB", "BEC", false, auditTime)
	if !ok || r.DataStatus != "no-service" {
		t.Fatal(r, ok)
	}
	r, ok = parseSchedule(document(string(b)), "PBB", "BEC", false, auditTime.AddDate(0, 0, 7))
	if !ok || len(r.Sailings) != 1 || r.Sailings[0].ServiceDate != "2026-09-18" {
		t.Fatal(r, ok)
	}
	b, e = os.ReadFile("testdata/gambier_keats.html")
	if e != nil {
		t.Fatal(e)
	}
	s := parseGambier(document(string(b)), auditTime)
	if s.DataStatus != "available" || len(s.Rows) != 52 || s.ValidThrough != "2026-10-12" {
		t.Fatal(s.DataState, len(s.Rows), s.Notes)
	}
	s = parseGambier(document(string(b)), auditTime.AddDate(0, 2, 0))
	if s.DataStatus != "unavailable" {
		t.Fatal("expired special schedule accepted")
	}
}

// Optional complete replay of the independently captured live audit. No network
// or DB calls. Golden times come from the independent source comparison.
func TestCapturedLiveNetwork(t *testing.T) {
	root := os.Getenv("BCF_AUDIT_DIR")
	if root == "" {
		t.Skip("set BCF_AUDIT_DIR to the saved Sep 11 live audit")
	}
	read := func(t *testing.T, name string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var schedules []struct {
		Pair       string `json:"pair"`
		Mode       string `json:"mode"`
		Result     string `json:"result"`
		SourceRows []struct {
			Time    string `json:"time"`
			Arrival string `json:"arrival"`
		} `json:"sourceRows"`
	}
	if e := json.Unmarshal(read(t, "reviewed-schedule-comparison.json"), &schedules); e != nil {
		t.Fatal(e)
	}
	for _, p := range schedules {
		t.Run(p.Pair, func(t *testing.T) {
			r, ok := parseSchedule(document(string(read(t, "routes-fares_schedules_daily_"+p.Pair+".html"))), p.Pair[:3], p.Pair[4:], p.Mode == "daily", auditTime)
			if !ok {
				file := "routes-fares_schedules_seasonal_" + p.Pair + ".html"
				if _, e := os.Stat(filepath.Join(root, file)); e == nil {
					r, ok = parseSchedule(document(string(read(t, file))), p.Pair[:3], p.Pair[4:], false, auditTime)
				}
			}
			if p.Result == "UPSTREAM_SCHEDULE_UNAVAILABLE" {
				if ok {
					t.Fatal("accepted unavailable source")
				}
				return
			}
			if !ok || len(r.Sailings) != len(p.SourceRows) {
				t.Fatalf("got %d (%s, ok=%v) want %d", len(r.Sailings), r.DataStatus, ok, len(p.SourceRows))
			}
			for i, s := range r.Sailings {
				if !strings.EqualFold(s.DepartureTime, p.SourceRows[i].Time) || !strings.EqualFold(s.ArrivalTime, p.SourceRows[i].Arrival) || s.ServiceDate != "2026-09-11" || s.Duration == "" {
					t.Fatal(s, p.SourceRows[i])
				}
			}
		})
	}
	var details []struct {
		URL  string `json:"url"`
		HTML string `json:"html"`
	}
	json.Unmarshal(read(t, "capacity-details.json"), &details)
	dm := map[string]string{}
	for _, v := range details {
		dm[v.URL] = v.HTML
	}
	fetch := func(u string) (*goquery.Document, error) {
		v, ok := dm[u]
		if !ok {
			return nil, fmt.Errorf("uncaptured URL %s", u)
		}
		return document(v), nil
	}
	var conditions []struct {
		Pair       string   `json:"pair"`
		SourceRows []string `json:"sourceRows"`
	}
	json.Unmarshal(read(t, "conditions-comparison.json"), &conditions)
	for _, p := range conditions {
		t.Run("conditions-"+p.Pair, func(t *testing.T) {
			c := parseConditions(document(string(read(t, "current-conditions_"+p.Pair+".html"))), p.Pair[:3], p.Pair[4:], auditTime, fetch)
			if len(c.Sailings) != len(p.SourceRows) || c.DataStatus == "unavailable" || len(c.Warnings) > 0 {
				t.Fatal(c.DataState, len(c.Sailings), len(p.SourceRows))
			}
			for _, s := range c.Sailings {
				if s.DepartureTime == "" || s.VesselName == "" || s.ServiceDate == "" {
					t.Fatal(s)
				}
			}
		})
	}
	var old struct {
		Capacity []models.CapacityRoute `json:"capacityRoutes"`
	}
	json.Unmarshal(read(t, "replay-api.json"), &old)
	for _, route := range old.Capacity {
		c := parseConditions(document(string(read(t, "current-conditions_"+route.FromTerminalCode+"-"+route.ToTerminalCode+".html"))), route.FromTerminalCode, route.ToTerminalCode, auditTime, fetch)
		for i, s := range c.Sailings {
			prior := route.Sailings[i]
			if s.Fill != prior.Fill || s.CarFill != prior.CarFill || s.OversizeFill != prior.OversizeFill {
				t.Errorf("capacity changed %s row %d: %d/%d/%d != %d/%d/%d", route.RouteCode, i, s.Fill, s.CarFill, s.OversizeFill, prior.Fill, prior.CarFill, prior.OversizeFill)
			}
		}
	}
	service := parseGambier(document(string(read(t, "routes-fares_schedules_gambier-keats.html"))), auditTime)
	if service.DataStatus != "available" || len(service.Rows) != 52 || service.ValidThrough != "2026-10-12" {
		t.Fatalf("special service: %s rows=%d notes=%v", service.DataStatus, len(service.Rows), service.Notes)
	}
}
