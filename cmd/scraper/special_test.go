package scraper

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestGambierDates(t *testing.T) {
	for _, tc := range []struct {
		name, header, start, end string
	}{
		{"shared year", "October 13 to December 16, 2026", "2026-10-13", "2026-12-16"},
		{"explicit years", "September 8, 2026 to October 12, 2026", "2026-09-08", "2026-10-12"},
		{"shared year crossing January", "December 17 to January 5, 2027", "2026-12-17", "2027-01-05"},
		{"explicit years crossing January", "December 17, 2026 to January 5, 2027", "2026-12-17", "2027-01-05"},
		{"same day", "October 13 to October 13, 2026", "2026-10-13", "2026-10-13"},
		{"whitespace", "October\u00a013\n to December 16, 2026", "2026-10-13", "2026-12-16"},
		{"missing end year", "October 13 to December 16", "", ""},
		{"invalid start", "October 32 to December 16, 2026", "", ""},
		{"invalid end", "October 13 to December 32, 2026", "", ""},
		{"invalid month", "Octember 13 to December 16, 2026", "", ""},
		{"explicit reversed dates", "December 17, 2027 to January 5, 2027", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end, err := gambierDates("Schedule in effect: " + tc.header)
			if tc.start == "" {
				if err == nil {
					t.Fatal("invalid date range accepted")
				}
				return
			}
			if err != nil || start.Format("2006-01-02") != tc.start || end.Format("2006-01-02") != tc.end {
				t.Fatalf("got %v to %v, error %v", start, end, err)
			}
		})
	}
}

func TestGambierSharedYearSchedule(t *testing.T) {
	b, err := os.ReadFile("testdata/gambier_keats.html")
	if err != nil {
		t.Fatal(err)
	}
	// Replay the saved table with the exact date header published October 26.
	html := strings.Replace(string(b), "September 8, 2026 to October 12, 2026", "October 13 to December 16, 2026", 1)
	for _, tc := range []struct {
		name, at, status string
	}{
		{"before Pacific start", "2026-10-13T06:59:59Z", "unavailable"},
		{"Pacific start", "2026-10-13T07:00:00Z", "available"},
		{"reported production failure", "2026-10-27T04:00:00Z", "available"},
		{"Pacific last second", "2026-12-17T07:59:59Z", "available"},
		{"expired in Pacific", "2026-12-17T08:00:00Z", "unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			at, err := time.Parse(time.RFC3339, tc.at)
			if err != nil {
				t.Fatal(err)
			}
			s := parseGambier(document(html), at)
			if s.DataStatus != tc.status || s.ValidFrom != "2026-10-13" || s.ValidThrough != "2026-12-16" || len(s.Rows) != 52 || len(s.Columns) != 5 || len(s.Notes) < 4 {
				t.Fatalf("unexpected schedule: %+v", s)
			}
			if s.CheckedAt == nil || !s.CheckedAt.Equal(at) {
				t.Fatal("missing check timestamp")
			}
			if tc.status == "available" {
				if !s.ObservedAt.Equal(at) || len(s.Warnings) != 0 {
					t.Fatal("available schedule missing observation or has warnings")
				}
			} else if !s.ObservedAt.IsZero() || len(s.Warnings) == 0 {
				t.Fatal("invalid schedule must explain failure without a fresh observation")
			}
		})
	}
}

func TestGambierInvalidMarkup(t *testing.T) {
	for _, html := range []string{
		"<html>Waiting room</html>",
		"<table><tr><td>Schedule in effect: unknown</td></tr></table>",
		"<table><tr><td>Schedule in effect: October 13 to December 16, 2026</td></tr></table>",
	} {
		s := parseGambier(document(html), time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
		if s.DataStatus != "unavailable" || len(s.Warnings) == 0 || !s.ObservedAt.IsZero() {
			t.Fatalf("invalid page accepted or unexplained: %+v", s)
		}
	}
}
