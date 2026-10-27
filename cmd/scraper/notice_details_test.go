package scraper

import (
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func TestNoticeBodyOnlyIncludesArticle(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`<div class="advisoryDescription"><p>Other route alert</p></div><div class="container service-notice"><h2 class="title">Sailing Delay</h2><h3 class="service-region">Posted yesterday</h3><p><p>Updated: 8:46 am</p><p>The <a href="/route">10:00 am sailing</a> is delayed&nbsp;by 20 minutes.</p><ul><li>Arrive at the usual check-in time.</li></ul><script>untrusted()</script></p></div><footer><p>Subscribe</p></footer>`))
	want := "Updated: 8:46 am\n\nThe 10:00 am sailing is delayed by 20 minutes.\n\nArrive at the usual check-in time."
	if got := parseNoticeBody(doc); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	empty, _ := goquery.NewDocumentFromReader(strings.NewReader(`<main><p>Unavailable</p></main>`))
	if got := parseNoticeBody(empty); got != "" {
		t.Fatal(got)
	}
}

func TestNoticeDetailsCacheAndFailure(t *testing.T) {
	cache := noticeDetailCache{entries: make(map[string]noticeDetail)}
	link := "https://www.bcferries.com/current-conditions/service-notices?serviceNoticeCode=123"
	var calls atomic.Int32
	fetch := func(string) (*goquery.Document, error) {
		calls.Add(1)
		return goquery.NewDocumentFromReader(strings.NewReader(`<div class="container service-notice"><p>Sailing delayed.</p></div>`))
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cache.get(link, fetch).body != "Sailing delayed." {
				t.Error("missing body")
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("fetched %d times", calls.Load())
	}
	entry := cache.entries[link]
	entry.expires = time.Now().Add(-time.Second)
	cache.entries[link] = entry
	failed := cache.get(link, func(string) (*goquery.Document, error) { return nil, errors.New("offline") })
	if failed.body != "" || !failed.observedAt.IsZero() {
		t.Fatal("failed fetch must not present old notice text as current")
	}
	cache.get("https://example.com/current-conditions/service-notices?serviceNoticeCode=123", fetch)
	if calls.Load() != 1 {
		t.Fatal("non-official URL fetched")
	}
}

func TestCapturedLiveNoticeBody(t *testing.T) {
	path := os.Getenv("BCF_NOTICE_HTML")
	if path == "" {
		t.Skip("optional captured upstream notice")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	body := parseNoticeBody(doc)
	if !strings.Contains(body, "departed 41 minutes behind schedule") || !strings.Contains(body, "10:00 am departing Tsawwassen") {
		t.Fatalf("missing captured delay explanation: %q", body)
	}
	if strings.Contains(body, "Subscribe to service notices") {
		t.Fatal("included unrelated page content")
	}
}
