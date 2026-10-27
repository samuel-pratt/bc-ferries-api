package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"regexp"
	"strings"
	"time"
)

const gambierURL = "https://www.bcferries.com/routes-fares/schedules/gambier-keats"

func parseGambier(d *goquery.Document, at time.Time) models.SpecialService {
	r := models.SpecialService{ID: "gambier-keats", Name: "Langdale, Gambier Island and Keats Island", DataState: models.DataState{DataStatus: "unavailable", CheckedAt: &at, SourceURL: gambierURL}, Rows: [][]string{}, Columns: []string{}, Notes: []string{}}
	table := d.Find("table").FilterFunction(func(_ int, t *goquery.Selection) bool { return strings.Contains(t.Text(), "Schedule in effect:") }).First()
	if table.Length() == 0 {
		return r
	}
	pattern := regexp.MustCompile(`Schedule in effect:\s*([A-Za-z]+ \d{1,2}, \d{4})\s+to\s+([A-Za-z]+ \d{1,2}, \d{4})`)
	match := pattern.FindStringSubmatch(clean(table.Text()))
	if len(match) != 3 {
		return r
	}
	start, e := time.Parse("January 2, 2006", match[1])
	end, e2 := time.Parse("January 2, 2006", match[2])
	if e != nil || e2 != nil {
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
	}
	return r
}
