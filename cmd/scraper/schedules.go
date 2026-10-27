package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func onward(d *goquery.Document) *goquery.Selection {
	if d.Find("#ReturnSchedule").Length() == 0 {
		return d.Selection
	}
	copy := d.Selection.Clone()
	stopped := false
	var trim func(*html.Node)
	trim = func(n *html.Node) {
		for child := n.FirstChild; child != nil; {
			next := child.NextSibling
			for _, a := range child.Attr {
				if a.Key == "id" && a.Val == "ReturnSchedule" {
					stopped = true
				}
			}
			if stopped {
				n.RemoveChild(child)
			} else {
				trim(child)
			}
			child = next
		}
	}
	for _, n := range copy.Nodes {
		trim(n)
	}
	return copy
}
func weekday(s string) string { return strings.TrimSuffix(strings.ToUpper(clean(s)), "S") }

var monthToken = regexp.MustCompile(`(?i)\b(jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:t(?:ember)?)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\b`)
var dayToken = regexp.MustCompile(`\b\d{1,2}\b`)

// Both month-first lists (Sep 11 & 25) and day-first lists (18 Sep).
func mentionedDates(note string) map[string]bool {
	dates := map[string]bool{}
	months := monthToken.FindAllStringIndex(note, -1)
	for i, p := range months {
		m := strings.ToLower(note[p[0]:p[1]])[:3]
		mon := map[string]int{"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6, "jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12}[m]
		end := len(note)
		if i+1 < len(months) {
			end = months[i+1][0]
		}
		tail := note[p[1]:end]
		numbers := dayToken.FindAllString(tail, -1)
		// Day-first notation is detected from a number immediately before the month.
		before := regexp.MustCompile(`(\d{1,2})\s*$`).FindStringSubmatch(note[:p[0]])
		if len(before) > 0 {
			numbers = []string{before[1]}
		} else {
			tail = regexp.MustCompile(`^[\s\d,&]+`).FindString(tail)
			numbers = dayToken.FindAllString(tail, -1)
		}
		for _, n := range numbers {
			day, _ := strconv.Atoi(n)
			if day > 0 && day <= 31 {
				dates[strconv.Itoa(mon)+"-"+strconv.Itoa(day)] = true
			}
		}
	}
	return dates
}
func permitted(note string, day time.Time) (bool, bool) {
	lower := strings.ToLower(note)
	if strings.Contains(lower, "dangerous goods only") || strings.Contains(lower, "no passengers permitted") {
		return false, true
	}
	only := strings.Contains(lower, "only on") || strings.Contains(lower, "only sails on")
	except := strings.Contains(lower, "except") || strings.Contains(lower, "not available on")
	if !only && !except {
		return true, true
	}
	dates := mentionedDates(note)
	if len(dates) == 0 {
		return false, false
	}
	matches := dates[strconv.Itoa(int(day.Month()))+"-"+strconv.Itoa(day.Day())]
	return (!only || matches) && (!except || !matches), true
}

func parseStops(s *goquery.Selection) []models.Stop {
	out := []models.Stop{}
	s.Find(".schedule-stop-details-first-word p").Each(func(_ int, p *goquery.Selection) {
		text := clean(p.Text())
		for _, kind := range []string{"Non-stop", "Transfer", "Stop", "Thru fare"} {
			if strings.HasPrefix(text, kind) {
				name := strings.TrimSpace(strings.TrimPrefix(text, kind))
				name = strings.TrimPrefix(name, "at ")
				out = append(out, models.Stop{Type: strings.ToLower(strings.ReplaceAll(kind, " ", "-")), TerminalName: name})
				break
			}
		}
	})
	s.Find(".progtrckr ul").Each(func(_ int, p *goquery.Selection) {
		text := clean(p.Text())
		for _, kind := range []string{"Non-stop", "Transfer", "Stop", "Thru fare"} {
			if strings.HasPrefix(text, kind) {
				name := strings.TrimSpace(strings.TrimPrefix(text, kind))
				name = strings.TrimPrefix(name, "at ")
				out = append(out, models.Stop{Type: strings.ToLower(strings.ReplaceAll(kind, " ", "-")), TerminalName: name})
				break
			}
		}
	})
	if via := clean(s.Find("#stop-names").Text()); via != "" {
		out = append(out, models.Stop{Type: "itinerary", TerminalName: via})
	}
	return out
}

func parseScheduleRow(row *goquery.Selection, at time.Time) (models.NonCapacitySailing, bool, bool) {
	cells := row.ChildrenFiltered("td")
	s := models.NonCapacitySailing{}
	if cells.Length() < 4 {
		return s, false, false
	}
	departureCell := cells.Eq(1).Clone()
	departureCell.Find("p").Remove()
	dep, arr := clockPattern.FindString(clean(departureCell.Text())), clockPattern.FindString(clean(cells.Eq(2).Text()))
	notes := []string{}
	cells.Eq(1).Find("p").Each(func(_ int, p *goquery.Selection) {
		if v := clean(p.Text()); v != "" {
			notes = append(notes, v)
		}
	})
	s.VesselStatus = strings.Join(uniqueStrings(notes), " | ")
	if strings.Contains(strings.ToLower(row.Text()), "no sailings available") {
		return s, false, true
	}
	include, recognized := permitted(s.VesselStatus, pacific(at))
	if !include {
		return s, false, recognized
	}
	if dep == "" || arr == "" {
		return s, false, false
	}
	s.DepartureTime = dep
	s.ArrivalTime = arr
	s.Status = "scheduled"
	s.ObservedAt = at
	s.ServiceDate = pacific(at).Format("2006-01-02")
	if raw := cells.Eq(1).AttrOr("data-sort", ""); raw != "" {
		sourceDate, err := time.Parse("01/02/2006 15:04:05", raw)
		if err == nil && sourceDate.Format("2006-01-02") != s.ServiceDate {
			return s, false, false
		}
	}
	s.ScheduledTime = dep
	s.ArrivalDate = arrivalDate(s.ServiceDate, dep, arr)
	s.Duration = clean(cells.Eq(3).Text())
	s.Stops = parseStops(row)
	return s, true, true
}

func parseSchedule(d *goquery.Document, from, to string, daily bool, at time.Time) (models.NonCapacityRoute, bool) {
	r := models.NonCapacityRoute{DataState: models.DataState{DataStatus: "available", ServiceDate: pacific(at).Format("2006-01-02"), CheckedAt: &at, SourceURL: MakeScheduleLink(from, to)}, RouteCode: from + to, FromTerminalCode: from, ToTerminalCode: to, ObservedAt: at, Sailings: []models.NonCapacitySailing{}}
	scope := onward(d)
	r.Notices = parseServiceNotices(d, r.RouteCode, from, at)
	var rows *goquery.Selection
	if daily {
		table := scope.Find("#dailyScheduleTableOnward")
		if table.Length() == 0 {
			return r, false
		}
		rows = table.ChildrenFiltered("tbody").ChildrenFiltered("tr.schedule-table-row")
		if rows.Length() == 0 {
			return r, false
		}
	} else {
		r.SourceURL = MakeSeasonalScheduleLink(from, to)
		if strings.Contains(strings.ToLower(scope.Text()), "seasonal schedules have not been posted") {
			return r, false
		}
		if !validSeason(scope, at) {
			return r, false
		}
		table := scope.Find("table.table-seasonal-schedule").FilterFunction(func(_ int, t *goquery.Selection) bool { return t.Find("thead [data-schedule-day]").Length() > 0 }).First()
		if table.Length() == 0 {
			return r, false
		}
		headers := table.ChildrenFiltered("thead")
		var body *goquery.Selection
		headers.Each(func(_ int, h *goquery.Selection) {
			if weekday(h.Find("[data-schedule-day]").First().AttrOr("data-schedule-day", "")) == weekday(pacific(at).Weekday().String()) {
				body = h.NextFiltered("tbody")
			}
		})
		if body == nil {
			r.DataStatus = "no-service"
			return r, true
		}
		rows = body.ChildrenFiltered("tr.schedule-table-row")
		if rows.Length() == 0 {
			return r, false
		}
	}
	valid := true
	seen := map[string]bool{}
	rows.Each(func(_ int, row *goquery.Selection) {
		s, include, ok := parseScheduleRow(row, at)
		if !ok {
			valid = false
		}
		if include {
			key := strings.ToLower(s.DepartureTime + "|" + s.ArrivalTime)
			if !seen[key] {
				r.Sailings = append(r.Sailings, s)
				seen[key] = true
				if r.SailingDuration == "" {
					r.SailingDuration = s.Duration
				}
			}
		}
	})
	if !valid {
		r.Warnings = append(r.Warnings, "unrecognized schedule row or date restriction")
		return r, false
	}
	if len(r.Sailings) == 0 {
		r.DataStatus = "no-service"
	}
	return r, true
}

var seasonRange = regexp.MustCompile(`(?i)([a-z]{3} \d{1,2}, \d{4})\s*-\s*([a-z]{3} \d{1,2}, \d{4})`)

func validSeason(scope *goquery.Selection, at time.Time) bool {
	m := seasonRange.FindStringSubmatch(clean(scope.Text()))
	if len(m) == 0 {
		return true
	}
	a, e := time.Parse("Jan 2, 2006", m[1])
	b, e2 := time.Parse("Jan 2, 2006", m[2])
	day := pacific(at).Format("2006-01-02")
	return e == nil && e2 == nil && day >= a.Format("2006-01-02") && day <= b.Format("2006-01-02")
}

// Kept as a small pure helper for existing callers/tests.
func parseDailyScheduleSailings(d *goquery.Document) ([]models.NonCapacitySailing, string, bool) {
	at := time.Now().UTC()
	if raw := d.Find("#dailyScheduleTableOnward [data-sort]").First().AttrOr("data-sort", ""); raw != "" {
		loc, _ := time.LoadLocation("America/Vancouver")
		if date, e := time.ParseInLocation("01/02/2006 15:04:05", raw, loc); e == nil {
			at = date
		}
	}
	r, ok := parseSchedule(d, "", "", true, at)
	return r.Sailings, r.SailingDuration, ok && len(r.Sailings) > 0
}
