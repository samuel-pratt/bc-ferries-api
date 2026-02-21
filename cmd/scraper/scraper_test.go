package scraper

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseDailyScheduleSailings_ParsesUpdatedSWBTSAOnwardTimes(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "html", "daily_schedule.html")
	html, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("fixture not found at %s: %v", fixturePath, err)
	}

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse fixture HTML: %v", err)
	}

	sailings, duration, ok := parseDailyScheduleSailings(document)
	if !ok {
		t.Fatalf("expected daily sailings to be parsed")
	}

	if len(sailings) != 10 {
		t.Fatalf("expected 10 onward sailings, got %d", len(sailings))
	}

	if duration != "01:35" {
		t.Fatalf("expected duration 01:35, got %q", duration)
	}

	seen := make(map[string]bool, len(sailings))
	for _, sailing := range sailings {
		seen[sailing.DepartureTime] = true
	}

	if !seen["8:00 am"] {
		t.Fatalf("expected parsed sailings to include 8:00 am")
	}

	if !seen["12:00 pm"] {
		t.Fatalf("expected parsed sailings to include 12:00 pm")
	}
}
