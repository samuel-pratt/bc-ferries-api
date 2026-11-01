package scraper

import (
	"fmt"
	"testing"
	"time"
)

func TestConditionsExplicitServiceDates(t *testing.T) {
	// On Nov 1, TSA-SGI publishes two 9:55 sailings on different future days.
	// The Nov 3 sailing has 61% space available (39% full), not a sailing today.
	at := time.Date(2026, 11, 2, 0, 30, 0, 0, time.UTC)
	d := document(`<table class="detail-departure-table">
<tr class="mobile-friendly-row"><td><p><span>9:55 am&nbsp;</span> (Tomorrow)<br><a>Salish Eagle</a></p></td><td><span class="cc-vessel-percent-full">13%</span></td></tr>
<tr class="mobile-friendly-row"><td><p><span>9:55 am&nbsp;</span> (Nov 03, 2026)<br><a>Salish Eagle</a></p></td><td><span class="cc-vessel-percent-full">61%</span><a href="/next-sailing-availability?departureTime=2026-11-03 09:55:00&amp;routeCode=TSA-SGI">Details</a></td></tr>
<tr class="sgi-row"><td><span id="stop-names">via Galiano Island (Sturdies Bay) to Salt Spring Island (Long Harbour)</span></td></tr>
</table>`)
	c := parseConditions(d, "TSA", "SGI", at, noDetails)
	if c.DataStatus != "available" || c.ServiceDate != "2026-11-01" || len(c.Sailings) != 2 {
		t.Fatalf("unexpected snapshot: %#v", c)
	}
	for i, want := range []struct {
		date string
		fill int
	}{{"2026-11-02", 87}, {"2026-11-03", 39}} {
		s := c.Sailings[i]
		if s.ServiceDate != want.date || s.DepartureTime != "9:55 am" || s.ScheduledTime != "9:55 am" ||
			s.VesselName != "Salish Eagle" || s.Fill != want.fill || s.CapacityKnown == nil || !*s.CapacityKnown || s.Status != "scheduled" {
			t.Fatalf("sailing %d has wrong date, vessel or capacity: %#v", i, s)
		}
	}
	if len(c.Sailings[1].Stops) != 1 {
		t.Fatal("lost SGI itinerary")
	}
}

func TestConditionsDateLabels(t *testing.T) {
	at := time.Date(2026, 12, 31, 20, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		label string
		date  string
	}{
		{"", "2026-12-31"},
		{"(Tomorrow)", "2027-01-01"},
		{"(Jan 02, 2027)", "2027-01-02"},
		{"(January 2, 2027)", "2027-01-02"},
		{"(Feb 30, 2027)", ""},
	} {
		t.Run(test.label, func(t *testing.T) {
			d := document(fmt.Sprintf(`<table class="detail-departure-table"><tr class="mobile-friendly-row"><td>11:45 pm %s Departed 11:52 pm Salish Eagle</td><td>ETA : 1:30 am</td></tr></table>`, test.label))
			c := parseConditions(d, "TSA", "SGI", at, noDetails)
			if test.date == "" {
				if c.DataStatus != "unavailable" || len(c.Warnings) == 0 || len(c.Sailings) != 0 {
					t.Fatalf("invalid date accepted: %#v", c)
				}
				return
			}
			if c.DataStatus != "available" || len(c.Sailings) != 1 {
				t.Fatalf("unexpected snapshot: %#v", c)
			}
			s := c.Sailings[0]
			date, _ := time.Parse("2006-01-02", test.date)
			if s.ServiceDate != test.date || s.ArrivalDate != date.AddDate(0, 0, 1).Format("2006-01-02") ||
				s.VesselName != "Salish Eagle" || s.DepartureTime != "11:52 pm" || s.ScheduledTime != "11:45 pm" {
				t.Fatalf("incorrect dated overnight sailing: %#v", s)
			}
		})
	}
}
