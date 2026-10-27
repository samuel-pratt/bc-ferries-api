package scraper

import "testing"

func TestNoticeRouteScope(t *testing.T) {
	for _, tc := range []struct {
		title, route string
		want         bool
	}{
		{"Otter Bay Terminal Life Extension Project", "TSASWB", false},
		{"Otter Bay Terminal Life Extension Project", "TSAPOB", true},
		{"Otter Bay Terminal Life Extension Project", "POBSWB", true},
		{"Otter Bay Terminal Life Extension Project", "TSASGI", true},
		{"Tsawwassen - Southern Gulf Islands revised schedule", "TSASWB", false},
		{"Tsawwassen - Southern Gulf Islands revised schedule", "TSASGI", true},
		{"Tsawwassen - Swartz Bay cancellation", "TSASWB", true},
		{"Swartz Bay parking update", "TSASWB", true},
		{"Weather disruption", "TSASWB", true},
	} {
		_, got := noticeScope(tc.title, tc.route, tc.route[:3])
		if got != tc.want {
			t.Errorf("%s / %s: %v", tc.title, tc.route, got)
		}
	}
}
