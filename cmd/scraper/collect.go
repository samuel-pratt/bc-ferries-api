package scraper

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"github.com/samuel-pratt/bc-ferries-api/cmd/staticdata"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var sourceClient = &http.Client{Timeout: 12 * time.Second}

var scheduleScrape sync.Mutex

func fetchDocument(url string) (*goquery.Document, error) {
	req, e := http.NewRequest(http.MethodGet, url, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BCFerriesAPI)")
	response, e := sourceClient.Do(req)
	if e != nil {
		return nil, e
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("source HTTP %d", response.StatusCode)
	}
	b, e := io.ReadAll(io.LimitReader(response.Body, (8<<20)+1))
	if e != nil {
		return nil, e
	}
	if len(b) > 8<<20 {
		return nil, fmt.Errorf("source document too large")
	}
	return goquery.NewDocumentFromReader(strings.NewReader(string(b)))
}

type routeCatalog struct {
	models.DataState
	Pairs []string `json:"pairs"`
}

var pairPattern = regexp.MustCompile(`/(?:current-conditions|schedules/(?:daily|seasonal))/([A-Z]{3}-[A-Z]{3})(?:[/?#]|$)`)

func discoverPairs(d *goquery.Document) []string {
	out := []string{}
	d.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
		if m := pairPattern.FindStringSubmatch(a.AttrOr("href", "")); len(m) > 0 {
			out = append(out, m[1])
		}
	})
	return uniqueStrings(out)
}
func catalog(kind string, fallback []string, d *goquery.Document) []string {
	pairs := append([]string{}, fallback...)
	var previous routeCatalog
	if b := db.Snapshots("catalog")[kind]; len(b) > 0 {
		_ = json.Unmarshal(b, &previous)
		pairs = append(pairs, previous.Pairs...)
	}
	status := "unavailable"
	if d != nil {
		found := discoverPairs(d)
		if len(found) >= len(fallback)/2 {
			pairs = append(pairs, found...)
			status = "available"
		}
	}
	pairs = uniqueStrings(pairs)
	sort.Strings(pairs)
	at := time.Now().UTC()
	if e := db.SaveSnapshot("catalog", kind, routeCatalog{DataState: models.DataState{DataStatus: status, CheckedAt: &at}, Pairs: pairs}); e != nil {
		log.Print(e)
	}
	return pairs
}

func ScrapeCapacityRoutes() {
	index, _ := fetchDocument("https://www.bcferries.com/current-conditions")
	pairs := catalog("conditions", staticdata.ConditionPairs(), index)
	jobs := make(chan string)
	var wg sync.WaitGroup
	// Bounded concurrency: all conditions pages, including status-only/terminal
	// pages, share two workers. Per-request deadlines bound detail requests too.
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				parts := strings.Split(p, "-")
				d, e := fetchDocument(MakeCurrentConditionsLink(parts[0], parts[1]))
				if e != nil {
					at := time.Now().UTC()
					_ = db.SaveSnapshot("conditions", parts[0]+parts[1], models.ConditionsSnapshot{DataState: models.DataState{DataStatus: "unavailable", CheckedAt: &at, SourceURL: MakeCurrentConditionsLink(parts[0], parts[1])}, Sailings: []models.CapacitySailing{}})
					if isCapacityPair(p) {
						_ = db.MarkRouteUnavailable("capacity", parts[0], parts[1], MakeCurrentConditionsLink(parts[0], parts[1]), at)
					}
					continue
				}
				ScrapeCapacityRoute(d, parts[0], parts[1])
			}
		}()
	}
	for _, p := range pairs {
		jobs <- p
	}
	close(jobs)
	wg.Wait()
}
func isCapacityPair(pair string) bool {
	for _, p := range staticdata.CapacityPairs() {
		if p == pair {
			return true
		}
	}
	return false
}
func ScrapeCapacityRoute(d *goquery.Document, from, to string) {
	at := time.Now().UTC()
	c := parseConditions(d, from, to, at, fetchDocument)
	enrichServiceNotices(c.Notices)
	if e := db.SaveSnapshot("conditions", from+to, c); e != nil {
		log.Print(e)
	}
	capacity := isCapacityPair(from + "-" + to)
	for _, s := range c.Sailings {
		if s.CapacityKnown != nil && *s.CapacityKnown {
			capacity = true
		}
	}
	if !capacity {
		return
	}
	if c.DataStatus != "available" && c.DataStatus != "no-service" {
		_ = db.MarkRouteUnavailable("capacity", from, to, c.SourceURL, at)
		return
	}
	duration := ""
	d.Find("span").Each(func(_ int, s *goquery.Selection) {
		v := clean(s.Text())
		if duration == "" && strings.HasPrefix(strings.ToLower(v), "sailing duration:") {
			duration = strings.TrimSpace(v[len("sailing duration:"):])
		}
	})
	saveRoute("capacity", from, to, duration, c.Sailings, models.RouteOperationalData{DataState: c.DataState, Notices: c.Notices, Terminal: c.Terminal, ObservedAt: c.ObservedAt})
}
func saveRoute(kind, from, to, duration string, sailings any, op models.RouteOperationalData) bool {
	b, e := json.Marshal(sailings)
	if e != nil {
		return false
	}
	o, e := json.Marshal(op)
	if e != nil {
		return false
	}
	if e = db.SaveRoute(kind, from+to, from, to, duration, b, o); e != nil {
		log.Printf("save %s %s-%s: %v", kind, from, to, e)
		return false
	}
	return true
}
func ScrapeNonCapacityRoute(d *goquery.Document, from, to string, daily bool) bool {
	r, ok := parseSchedule(d, from, to, daily, time.Now().UTC())
	if !ok {
		return false
	}
	enrichServiceNotices(r.Notices)
	return saveRoute("noncapacity", from, to, r.SailingDuration, r.Sailings, models.RouteOperationalData{DataState: r.DataState, Notices: r.Notices, Terminal: r.Terminal, ObservedAt: r.ObservedAt})
}
func ScrapeNonCapacityRoutes() {
	if !scheduleScrape.TryLock() {
		return
	}
	defer scheduleScrape.Unlock()
	ctx, cancel := newBrowserContext()
	defer cancel()
	if e := chromedp.Run(ctx); e != nil {
		log.Printf("Chrome startup failed: %v", e)
		return
	}
	html, _, e := fetchWithChromedp(ctx, "https://www.bcferries.com/routes-fares/schedules")
	var index *goquery.Document
	if e == nil {
		index, _ = goquery.NewDocumentFromReader(strings.NewReader(html))
	}
	pairs := catalog("schedules", staticdata.SchedulePairs(), index)
	for _, p := range pairs {
		parts := strings.Split(p, "-")
		if parts[1] == "SGI" {
			continue
		}
		ok := false
		for _, url := range []string{MakeScheduleLink(parts[0], parts[1]), MakeSeasonalScheduleLink(parts[0], parts[1])} {
			html, final, e := fetchWithChromedp(ctx, url)
			if e != nil {
				continue
			}
			d, e := goquery.NewDocumentFromReader(strings.NewReader(html))
			if e == nil && ScrapeNonCapacityRoute(d, parts[0], parts[1], isDailySchedulePage(d, final)) {
				ok = true
				break
			}
			if strings.Contains(final, "/seasonal/") {
				break
			}
		}
		if !ok {
			_ = db.MarkRouteUnavailable("noncapacity", parts[0], parts[1], MakeScheduleLink(parts[0], parts[1]), time.Now().UTC())
		}
	}
	html, _, e = fetchWithChromedp(ctx, gambierURL)
	if e == nil {
		d, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
		service := parseGambier(d, time.Now().UTC())
		if e = db.SaveSnapshot("special", "gambier-keats", service); e != nil {
			log.Print(e)
		}
	} else {
		at := time.Now().UTC()
		_ = db.SaveSnapshot("special", "gambier-keats", models.SpecialService{ID: "gambier-keats", DataState: models.DataState{DataStatus: "unavailable", CheckedAt: &at, SourceURL: gambierURL}})
	}
}
