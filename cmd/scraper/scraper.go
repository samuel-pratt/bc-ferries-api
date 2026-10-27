package scraper

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
)

/*
 * MakeCurrentConditionsLink
 *
 * Makes a link to the current conditions page for a given departure and destination
 *
 * @param string departure
 * @param string destination
 *
 * @return string
 */
func MakeCurrentConditionsLink(departure, destination string) string {
	return "https://www.bcferries.com/current-conditions/" + departure + "-" + destination
}

/*
 * MakeScheduleLink
 *
 * Builds a link to the DAILY schedule page for a given departure and destination.
 * Daily pages reflect route-specific service updates in effect for the current day.
 *
 * @param string departure
 * @param string destination
 *
 * @return string
 */
func MakeScheduleLink(departure, destination string) string {
	return "https://www.bcferries.com/routes-fares/schedules/daily/" + departure + "-" + destination
}

/*
 * MakeSeasonalScheduleLink
 *
 * Builds a link to the SEASONAL schedule page for a given departure and destination.
 *
 * @param string departure
 * @param string destination
 *
 * @return string
 */
func MakeSeasonalScheduleLink(departure, destination string) string {
	return "https://www.bcferries.com/routes-fares/schedules/seasonal/" + departure + "-" + destination
}

/*
 * ScrapeCapacityRoutes
 *
 * Scrapes capacity routes
 *
 * @return void
 */
func normalizeCapacityStatus(sailing *models.CapacitySailing) {
	sailing.StatusText = strings.TrimSpace(sailing.VesselStatus)
	switch sailing.SailingStatus {
	case "cancelled":
		sailing.Status = "cancelled"
	case "past":
		sailing.Status = "arrived"
	case "current":
		sailing.Status = "departed"
	default:
		sailing.Status = "scheduled"
	}

	combined := strings.TrimSpace(sailing.VesselName + " " + sailing.VesselStatus)
	delayRe := regexp.MustCompile(`(?i)delay(?:ed)?(?:\s+approx(?:imately)?\.?)?\s*(\d+)\s*(?:m|min|minutes?)`)
	if match := delayRe.FindStringSubmatch(combined); len(match) == 2 && sailing.Status == "scheduled" {
		if minutes, err := strconv.Atoi(match[1]); err == nil {
			sailing.DelayMinutes = &minutes
			sailing.Status = "delayed"
			if sailing.StatusText == "" {
				sailing.StatusText = match[0]
			}
		}
	}
}

func parseServiceNotices(document *goquery.Document, routeCode, terminalCode string, observedAt time.Time) []models.ServiceNotice {
	seen := make(map[string]struct{})
	notices := make([]models.ServiceNotice, 0)
	onward(document).Find(".cc-service-notice a[href], a[href*='serviceNoticeCode=']").Each(func(_ int, link *goquery.Selection) {
		href, ok := link.Attr("href")
		if !ok || strings.TrimSpace(href) == "" {
			return
		}
		absoluteURL := href
		if strings.HasPrefix(href, "/") {
			absoluteURL = "https://www.bcferries.com" + href
		}
		id := ""
		if match := regexp.MustCompile(`serviceNoticeCode=([0-9]+)`).FindStringSubmatch(href); len(match) == 2 {
			id = match[1]
		}
		if id == "" {
			id = strings.ToLower(strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return -1
			}, strings.TrimSpace(link.Text())))
		}
		if _, exists := seen[id]; exists || id == "" {
			return
		}
		seen[id] = struct{}{}
		title := strings.Join(strings.Fields(link.Text()), " ")
		terminalCodes, relevant := noticeScope(title, routeCode, terminalCode)
		if !relevant {
			return
		}
		noticeType, severity := classifyNotice(title)
		notices = append(notices, models.ServiceNotice{
			ID: id, Title: title, URL: absoluteURL, Type: noticeType, Severity: severity,
			TerminalCodes: terminalCodes, RouteCodes: []string{routeCode}, ObservedAt: observedAt,
		})
	})
	return notices
}

func classifyNotice(title string) (string, string) {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "cancel") || strings.Contains(lower, "closure"):
		return "cancellation", "critical"
	case strings.Contains(lower, "delay") || strings.Contains(lower, "revised schedule"):
		return "delay", "warning"
	case strings.Contains(lower, "terminal") || strings.Contains(lower, "parking"):
		return "terminal", "information"
	case strings.Contains(lower, "amenit") || strings.Contains(lower, "capacity"):
		return "service-change", "warning"
	default:
		return "information", "information"
	}
}

func parseTerminalInfo(document *goquery.Document, terminalCode string, observedAt time.Time) *models.TerminalInfo {
	info := &models.TerminalInfo{Code: terminalCode, ObservedAt: observedAt, Webcams: []models.Webcam{}}
	terminal := document.Find("#terminalDetails").First()
	info.Name = strings.Join(strings.Fields(terminal.Find("h4").First().Text()), " ")
	if href, ok := terminal.Find("a.terminal-detail-btn").First().Attr("href"); ok {
		if strings.HasPrefix(href, "/") {
			href = "https://www.bcferries.com" + href
		}
		info.DetailsURL = href
	}
	terminal.Find("a[href*='maps']").EachWithBreak(func(_ int, link *goquery.Selection) bool {
		info.DirectionsURL, _ = link.Attr("href")
		return false
	})
	document.Find(".webcam-wrapper img[src]").Each(func(index int, image *goquery.Selection) {
		src, ok := image.Attr("src")
		if !ok {
			return
		}
		container := image.Parent()
		label := strings.Join(strings.Fields(container.Find("p").First().Text()), " ")
		info.Webcams = append(info.Webcams, models.Webcam{
			ID:    fmt.Sprintf("%s-%d", strings.ToLower(terminalCode), index+1),
			Label: label, ImageURL: src, ObservedAt: observedAt,
		})
	})
	if info.Name == "" && len(info.Webcams) == 0 && info.DetailsURL == "" {
		return nil
	}
	return info
}

/********************/
/* Helper Functions */
/********************/

/*
 * fetchWithChromedp
 *
 * Uses a headless Chrome browser to fetch and render the full HTML content of a given URL.
 * This is used to bypass JavaScript-based protections like Queue-it by executing the page
 * in a real browser environment.
 *
 * @param string url - The URL to navigate to
 *
 * @return string - The full outer HTML of the rendered page
 * @return error - Any error encountered during navigation or retrieval
 */
func fetchWithChromedp(ctx context.Context, url string) (string, string, error) {
	tab, closeTab := chromedp.NewContext(ctx)
	defer closeTab()
	ctx, cancel := context.WithTimeout(tab, 45*time.Second)
	defer cancel()
	var html string
	var finalURL string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Poll(`!location.hostname.includes('queue-it.net') && document.readyState === 'complete'`, nil, chromedp.WithPollingInterval(time.Second)),
		chromedp.Location(&finalURL),
		chromedp.OuterHTML("html", &html),
	)

	return html, finalURL, err
}

func newBrowserContext() (context.Context, context.CancelFunc) {
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	// Docker runs Chrome as an unprivileged user, with all capabilities dropped.
	// Its nested namespace sandbox is unavailable on many Docker hosts.
	if os.Getenv("CHROME_NO_SANDBOX") == "true" {
		opts = append(opts, chromedp.NoSandbox)
	}
	alloc, closeAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, closeBrowser := chromedp.NewContext(alloc)
	return ctx, func() { closeBrowser(); closeAlloc() }
}

// isDailySchedulePage detects redirects from a daily URL to a seasonal page.
// The final browser URL is authoritative; canonical is a fallback for fixtures
// and pages whose client-side navigation leaves an ambiguous location.
func isDailySchedulePage(document *goquery.Document, finalURL string) bool {
	finalURL = strings.ToLower(finalURL)
	if strings.Contains(finalURL, "/schedules/seasonal/") {
		return false
	}
	if strings.Contains(finalURL, "/schedules/daily/") {
		return true
	}

	canonical, _ := document.Find(`link[rel="canonical"]`).First().Attr("href")
	canonical = strings.ToLower(canonical)
	return strings.Contains(canonical, "/schedules/daily/")
}
