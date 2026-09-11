package router

import (
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"testing"
	"time"
)

func TestMetadataDoesNotHideStaleRoutes(t *testing.T) {
	fresh, old := time.Now().UTC(), time.Now().UTC().Add(-time.Hour)
	routes := []models.CapacityRoute{{ObservedAt: fresh}, {ObservedAt: old}}
	if !buildResponseMeta(routes, nil, true, false).Stale {
		t.Fatal("fresh route hid stale route")
	}
	routes[1].ObservedAt = time.Time{}
	if !buildResponseMeta(routes, nil, true, false).Stale {
		t.Fatal("legacy route falsely marked fresh")
	}
	if !buildResponseMeta(nil, nil, true, false).Stale {
		t.Fatal("empty snapshot falsely marked fresh")
	}
}
