package scraper

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var clockPattern = regexp.MustCompile(`(?i)\b\d{1,2}:\d{2}\s*[ap]m\b`)
var departedPattern = regexp.MustCompile(`(?i)Departed\s+(\d{1,2}:\d{2}\s*[ap]m)`)
var arrivalPattern = regexp.MustCompile(`(?i)(Arrived|ETA)\s*:\s*(\d{1,2}:\d{2}\s*[ap]m|Variable)`)
var percentPattern = regexp.MustCompile(`^(\d{1,3})%$`)

func clean(s string) string { return strings.Join(strings.Fields(s), " ") }
func flag(v bool) *bool     { return &v }
func pacific(t time.Time) time.Time {
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		panic(err)
	}
	return t.In(loc)
}
func minutes(s string) (int, bool) {
	t, e := time.Parse("3:04pm", strings.ReplaceAll(strings.ToLower(s), " ", ""))
	return t.Hour()*60 + t.Minute(), e == nil
}
func arrivalDate(date, dep, arr string) string {
	d, e := time.Parse("2006-01-02", date)
	a, ok := minutes(arr)
	b, ok2 := minutes(dep)
	if e != nil || !ok || !ok2 {
		return ""
	}
	if a < b {
		d = d.AddDate(0, 0, 1)
	}
	return d.Format("2006-01-02")
}
func parseFull(s string) (int, bool) {
	s = clean(s)
	if strings.EqualFold(s, "full") {
		return 100, true
	}
	m := percentPattern.FindStringSubmatch(s)
	if len(m) != 2 {
		return 0, false
	}
	n, e := strconv.Atoi(m[1])
	return 100 - n, e == nil && n >= 0 && n <= 100
}

type detailFetcher func(string) (*goquery.Document, error)

func parseConditions(d *goquery.Document, from, to string, at time.Time, fetch detailFetcher) models.ConditionsSnapshot {
	result := models.ConditionsSnapshot{DataState: models.DataState{DataStatus: "available", ServiceDate: pacific(at).Format("2006-01-02"), CheckedAt: &at, SourceURL: MakeCurrentConditionsLink(from, to)}, ObservedAt: at, Sailings: []models.CapacitySailing{}}
	result.Notices = parseServiceNotices(d, from+to, from, at)
	result.Terminal = parseTerminalInfo(d, from, at)
	table := d.Find("table.detail-departure-table")
	if table.Length() == 0 {
		result.DataStatus = "not-published"
		if result.Terminal == nil && !strings.Contains(strings.ToLower(d.Text()), "current conditions") {
			result.DataStatus = "unavailable"
		}
		return result
	}
	table.Find("tr.mobile-friendly-row").Each(func(i int, row *goquery.Selection) {
		cells := row.ChildrenFiltered("td")
		left := clean(cells.First().Text())
		right := clean(cells.Eq(1).Text())
		s := models.CapacitySailing{SailingDetails: models.SailingDetails{ServiceDate: result.ServiceDate, CapacityKnown: flag(false), CarCapacityKnown: flag(false), OversizeCapacityKnown: flag(false), StatusObservedAt: &at}, ObservedAt: at}
		timeIndexes := clockPattern.FindAllStringIndex(left, -1)
		if cells.Length() < 2 || len(timeIndexes) == 0 {
			// A missing sailing is not a complete observation. Keep the parsed
			// rows for diagnostics, but never replace a good route with this list.
			result.DataStatus = "unavailable"
			result.Warnings = append(result.Warnings, fmt.Sprintf("unrecognized sailing row %d", i))
			return
		}
		s.DepartureTime = left[timeIndexes[0][0]:timeIndexes[0][1]]
		if _, valid := minutes(s.DepartureTime); !valid {
			result.DataStatus = "unavailable"
			result.Warnings = append(result.Warnings, fmt.Sprintf("invalid departure time in sailing row %d", i))
			return
		}
		s.ScheduledTime = s.DepartureTime
		cut := timeIndexes[0][1]
		if strings.Contains(left, "(Tomorrow)") {
			s.ServiceDate = pacific(at).AddDate(0, 0, 1).Format("2006-01-02")
		}
		s.SailingStatus = "future"
		s.Status = "scheduled"
		if m := departedPattern.FindStringSubmatch(left); len(m) > 0 {
			s.ActualDepartureTime = m[1]
			s.DepartureTime = m[1]
			s.SailingStatus = "current"
			s.Status = "departed"
			if len(timeIndexes) > 1 {
				cut = timeIndexes[1][1]
			}
		}
		s.VesselName = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(left[cut:]), "(Tomorrow)"))
		if m := arrivalPattern.FindStringSubmatch(right); len(m) > 0 {
			s.ArrivalTime = m[2]
			if strings.EqualFold(m[1], "Arrived") {
				s.ActualArrivalTime = m[2]
				s.SailingStatus = "past"
				s.Status = "arrived"
			} else {
				s.EstimatedArrivalTime = m[2]
			}
		}
		if strings.Contains(strings.ToLower(right), "cancelled") {
			s.SailingStatus = "cancelled"
			s.Status = "cancelled"
		}
		if strings.Contains(right, "...") && s.ActualDepartureTime == "" {
			s.SailingStatus = "unknown"
			s.Status = "unknown"
			s.ArrivalTime = "..."
		}
		// Preserve operational explanations, without duplicating the mobile time/vessel.
		messages := []string{}
		cells.Eq(1).Find(".cc-message-updates p, .text-red p, p.text-red").Each(func(_ int, p *goquery.Selection) {
			v := clean(p.Text())
			label := strings.ToLower(v)
			if v != "" && !arrivalPattern.MatchString(v) && label != "cancelled" && label != "estimated vehicle space available" && label != "upcoming" && label != "full" && !percentPattern.MatchString(v) {
				messages = append(messages, v)
			}
		})
		s.VesselStatus = strings.Join(uniqueStrings(messages), " | ")
		s.StatusText = s.VesselStatus
		if s.Status == "scheduled" {
			href := row.Find("a[href*='sailing-availability']").First().AttrOr("href", "")
			// Main-page percentage remains usable when a detail request fails.
			if n, ok := parseFull(clean(cells.Eq(1).Find(".cc-vessel-percent-full").First().Text())); ok {
				s.Fill = n
				s.CapacityKnown = flag(true)
			}
			if strings.EqualFold(right, "full") || strings.EqualFold(clean(cells.Eq(1).Find(".cc-message-updates").Text()), "full") {
				s.Fill = 100
				s.CarFill = 100
				s.OversizeFill = 100
				s.CapacityKnown = flag(true)
				// Preserve legacy numbers, but a route-wide Full label does not
				// supply independent measurements for either vehicle class.
			}
			if href != "" {
				details, err := fetch("https://www.bcferries.com" + strings.ReplaceAll(href, " ", "%20"))
				if err != nil {
					result.Warnings = append(result.Warnings, "vehicle capacity details unavailable")
				} else {
					values := []*int{&s.Fill, &s.CarFill, &s.OversizeFill}
					known := []**bool{&s.CapacityKnown, &s.CarCapacityKnown, &s.OversizeCapacityKnown}
					p := details.Find("p.vehicle-icon-text")
					if p.Length() != 3 {
						result.Warnings = append(result.Warnings, "unrecognized vehicle capacity details")
					}
					p.Each(func(j int, v *goquery.Selection) {
						if j >= 3 {
							return
						}
						if n, ok := parseFull(v.Text()); ok {
							*values[j] = n
							*known[j] = flag(true)
						} else {
							result.Warnings = append(result.Warnings, "invalid vehicle capacity percentage")
						}
					})
				}
			}
			if s.VesselStatus != "" {
				normalizeCapacityStatus(&s)
			}
		}
		s.ArrivalDate = arrivalDate(s.ServiceDate, s.DepartureTime, s.ArrivalTime)
		// SGI itineraries are published in the following detail row.
		detail := row.Next()
		if !detail.HasClass("mobile-friendly-row") {
			s.Stops = parseStops(detail)
		}
		result.Sailings = append(result.Sailings, s)
	})
	if len(result.Sailings) == 0 {
		result.DataStatus = "unavailable"
		result.Warnings = append(result.Warnings, "no recognizable current-condition sailings")
	}
	result.Warnings = uniqueStrings(result.Warnings)
	return result
}

func uniqueStrings(in []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, s := range in {
		if !seen[s] {
			out = append(out, s)
			seen[s] = true
		}
	}
	return out
}
