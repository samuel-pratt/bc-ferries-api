package router

import (
	"encoding/json"
	"fmt"
	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"strings"
	"time"
)

func checkSpecialServices(meta *models.ResponseMeta, services []models.SpecialService) {
	available := false
	loc, _ := time.LoadLocation("America/Vancouver")
	today := time.Now().In(loc).Format("2006-01-02")
	for _, s := range services {
		if s.ID == "gambier-keats" && s.DataStatus == "available" && !s.ObservedAt.IsZero() && time.Since(s.ObservedAt) <= 8*time.Hour && today >= s.ValidFrom && today <= s.ValidThrough {
			available = true
		}
	}
	if !available {
		meta.Complete = false
		meta.Issues = append(meta.Issues, "Gambier/Keats schedule unavailable")
	}
}

// Catalog discovery is checked separately from cached record freshness. A source
// outage or a newly listed, uncollected direction must not masquerade as complete.
func checkCatalogCoverage(meta *models.ResponseMeta, capacity []models.CapacityRoute, schedules []models.NonCapacityRoute, wantCapacity, wantSchedules bool) {
	catalogs := db.Snapshots("catalog")
	conditions := db.Snapshots("conditions")
	kinds := []string{"conditions"}
	if wantSchedules {
		kinds = append(kinds, "schedules")
	}
	for _, kind := range kinds {
		var catalog struct {
			models.DataState
			Pairs []string `json:"pairs"`
		}
		if json.Unmarshal(catalogs[kind], &catalog) != nil || catalog.DataStatus != "available" || catalog.CheckedAt == nil || time.Since(*catalog.CheckedAt) > 8*time.Hour {
			meta.Issues = append(meta.Issues, kind+" route discovery unavailable")
			continue
		}
		missing := 0
		for _, pair := range catalog.Pairs {
			code := strings.ReplaceAll(pair, "-", "")
			if kind == "conditions" {
				var c models.ConditionsSnapshot
				if json.Unmarshal(conditions[code], &c) != nil || c.DataStatus == "unavailable" || time.Since(c.ObservedAt) > 10*time.Minute || len(c.Warnings) > 0 {
					missing++
				}
			} else {
				found := false
				for _, r := range schedules {
					if r.RouteCode == code {
						found = true
						break
					}
				}
				if !found {
					missing++
				}
			}
		}
		if missing > 0 {
			meta.Issues = append(meta.Issues, fmt.Sprintf("%s routes missing or unavailable: %d", kind, missing))
		}
	}
	if len(meta.Issues) > 0 {
		meta.Complete = false
	}
}
