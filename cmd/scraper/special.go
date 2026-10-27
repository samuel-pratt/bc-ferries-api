package scraper

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
)

const gambierURL = "https://www.bcferries.com/routes-fares/schedules/gambier-keats"

var gambierDateRange = regexp.MustCompile(`Schedule in effect:\s*([A-Za-z]+ \d{1,2})(, \d{4})?\s+to\s+([A-Za-z]+ \d{1,2}, \d{4})`)

func gambierDates(text string) (time.Time, time.Time, error) {
	match := gambierDateRange.FindStringSubmatch(clean(text))
	if len(match) != 4 {
		return time.Time{}, time.Time{}, fmt.Errorf("schedule date range not recognized")
	}
	end, err := time.Parse("January 2, 2006", match[3])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid schedule end date: %w", err)
	}
	startText := match[1] + match[2]
	if match[2] == "" {
		// The published end year also applies to the start, unless the range
		// crosses New Year. Never infer the year from the server's clock.
		startText = fmt.Sprintf("%s, %d", match[1], end.Year())
	}
	start, err := time.Parse("January 2, 2006", startText)
	if err == nil && match[2] == "" && start.After(end) {
		start, err = time.Parse("January 2, 2006", fmt.Sprintf("%s, %d", match[1], end.Year()-1))
	}
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid schedule start date: %w", err)
	}
	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("schedule starts after its end date")
	}
	return start, end, nil
}

func parseGambier(d *goquery.Document, at time.Time) models.SpecialService {
	r := models.SpecialService{ID: "gambier-keats", Name: "Langdale, Gambier Island and Keats Island", DataState: models.DataState{DataStatus: "unavailable", CheckedAt: &at, SourceURL: gambierURL}, Rows: [][]string{}, Columns: []string{}, Notes: []string{}}
	table := d.Find("table").FilterFunction(func(_ int, t *goquery.Selection) bool { return strings.Contains(t.Text(), "Schedule in effect:") }).First()
	if table.Length() == 0 {
		r.Warnings = append(r.Warnings, "schedule table not found")
		return r
	}
	start, end, err := gambierDates(table.Text())
	if err != nil {
		r.Warnings = append(r.Warnings, err.Error())
		return r
	}
	r.ValidFrom = start.Format("2006-01-02")
	r.ValidThrough = end.Format("2006-01-02")
	table.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		cells := []string{}
		tr.ChildrenFiltered("td,th").Each(func(_ int, c *goquery.Selection) { cells = append(cells, clean(c.Text())) })
		if len(cells) == 1 && strings.HasPrefix(cells[0], "Additional trip") {
			r.Notes = append(r.Notes, cells[0])
		}
		if len(cells) != 5 {
			return
		}
		if strings.Contains(strings.Join(cells, " "), "Friday") {
			r.Columns = cells
			return
		}
		if strings.HasPrefix(cells[0], "Lv.") || strings.HasPrefix(cells[0], "Arr.") {
			r.Rows = append(r.Rows, cells)
		}
	})
	d.Find("p").Each(func(_ int, p *goquery.Selection) {
		v := clean(p.Text())
		if strings.HasPrefix(v, "▲") || strings.HasPrefix(v, "★") || strings.HasPrefix(v, "◆") || strings.Contains(v, "independently operated") || strings.Contains(v, "Passenger service only") || strings.Contains(v, "Some departures from Langdale") {
			r.Notes = append(r.Notes, v)
		}
	})
	r.Notes = uniqueStrings(r.Notes)
	day := pacific(at).Format("2006-01-02")
	if len(r.Columns) == 5 && len(r.Rows) > 0 && len(r.Notes) >= 4 && day >= r.ValidFrom && day <= r.ValidThrough {
		r.DataStatus = "available"
		r.ObservedAt = at
	} else if day < r.ValidFrom || day > r.ValidThrough {
		r.Warnings = append(r.Warnings, "schedule is not valid for today's Pacific date")
	} else {
		r.Warnings = append(r.Warnings, "schedule table or notes incomplete")
	}
	return r
}
