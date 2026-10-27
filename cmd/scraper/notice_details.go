package scraper

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
)

type noticeDetail struct {
	body       string
	observedAt time.Time
	expires    time.Time
}

// Shared between routes and scrape jobs, so a notice is fetched once, not once
// per route or app user. Failures leave the official link usable.
type noticeDetailCache struct {
	mu      sync.Mutex
	entries map[string]noticeDetail
}

var serviceNoticeDetails = noticeDetailCache{entries: make(map[string]noticeDetail)}

func (c *noticeDetailCache) get(rawURL string, fetch detailFetcher) noticeDetail {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Host != "www.bcferries.com" || u.Path != "/current-conditions/service-notices" || u.Query().Get("serviceNoticeCode") == "" {
		return noticeDetail{}
	}
	// Scraper workers already have bounded concurrency. Serialize detail fetches
	// to coalesce duplicate requests; the HTTP client has a 12-second timeout.
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UTC()
	if entry, ok := c.entries[rawURL]; ok && now.Before(entry.expires) {
		return entry
	}
	entry := noticeDetail{expires: now.Add(30 * time.Second)}
	if doc, err := fetch(rawURL); err == nil {
		entry.body = parseNoticeBody(doc)
		if entry.body != "" {
			entry.observedAt = time.Now().UTC()
			entry.expires = entry.observedAt.Add(2 * time.Minute)
		}
	}
	if len(c.entries) >= 256 {
		for key, value := range c.entries {
			if !now.Before(value.expires) {
				delete(c.entries, key)
			}
		}
		if len(c.entries) >= 256 {
			for key := range c.entries {
				delete(c.entries, key)
				break
			}
		}
	}
	c.entries[rawURL] = entry
	return entry
}

func parseNoticeBody(doc *goquery.Document) string {
	// Only the notice article, never navigation, a site-wide advisory, or footer.
	article := doc.Find(".container.service-notice").First().Clone()
	article.Find("script, style, h2.title, h3.service-region").Remove()
	article.Find("br").ReplaceWithHtml("\n")
	var paragraphs []string
	article.Find("p, li").Each(func(_ int, node *goquery.Selection) {
		if node.Find("p, li").Length() != 0 {
			return
		}
		if text := clean(node.Text()); text != "" {
			paragraphs = append(paragraphs, text)
		}
	})
	return strings.Join(paragraphs, "\n\n")
}

func enrichServiceNotices(notices []models.ServiceNotice) {
	for i := range notices {
		detail := serviceNoticeDetails.get(notices[i].URL, fetchDocument)
		notices[i].Body = detail.body
		if detail.body != "" {
			at := detail.observedAt
			notices[i].BodyObservedAt = &at
		}
	}
}
