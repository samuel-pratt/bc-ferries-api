package scraper

import "testing"

func TestConditionsFullLabelWithWarning(t *testing.T) {
	// Reduced from the live TSA-SWB 6pm row on Nov 1, 2026.
	d := document(`<table class="detail-departure-table"><tr class="mobile-friendly-row">
<td><p><span>6:00 pm</span>
<a>Coastal Celebration</a></p></td>
<td class="cc-status-td"><p class="mobile-paragraph">6:00 pm Coastal Celebration</p>
<div class="cc-percentage"><div class="percentage-full"><div class="cc-message-updates">
<span class="font-size-22"> Full </span>
<p class="text-red fnt-14 margin-bottom-0 italic-style">An earlier incident has caused some delays</p>
</div></div></div></td></tr></table>`)
	c := parseConditions(d, "TSA", "SWB", auditTime, noDetails)
	if c.DataStatus != "available" || len(c.Sailings) != 1 {
		t.Fatalf("unexpected snapshot: %#v", c)
	}
	s := c.Sailings[0]
	if s.CapacityKnown == nil || !*s.CapacityKnown || s.Fill != 100 {
		t.Fatalf("Full label lost beside warning: %#v", s)
	}
	if s.StatusText != "An earlier incident has caused some delays" || s.VesselStatus != s.StatusText {
		t.Fatalf("operational warning lost or changed: %#v", s)
	}
	if s.CarCapacityKnown == nil || *s.CarCapacityKnown || s.OversizeCapacityKnown == nil || *s.OversizeCapacityKnown {
		t.Fatalf("route total must not claim independently measured deck capacities: %#v", s)
	}
}

func TestConditionsDoesNotInferFullFromWarning(t *testing.T) {
	for _, markup := range []string{
		`<span>Not full</span><p>Loading continues</p>`,
		`<span>Upcoming</span><p>Earlier sailings were full</p>`,
		`<span>Upcoming</span><p>Full</p>`,
	} {
		d := document(`<table class="detail-departure-table"><tr class="mobile-friendly-row"><td>6:00 pm Coastal Celebration</td><td><div class="cc-message-updates">` + markup + `</div></td></tr></table>`)
		c := parseConditions(d, "TSA", "SWB", auditTime, noDetails)
		if len(c.Sailings) != 1 || c.Sailings[0].CapacityKnown == nil || *c.Sailings[0].CapacityKnown {
			t.Fatalf("warning mistaken for a Full label: %s: %#v", markup, c)
		}
	}
}
