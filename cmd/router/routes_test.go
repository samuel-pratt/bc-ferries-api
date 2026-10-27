package router

import (
	"encoding/json"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"github.com/samuel-pratt/bc-ferries-api/cmd/staticdata"
	"strings"
	"testing"
	"time"
)

func TestNoServiceCapacityHealth(t *testing.T) {
	at := time.Now().UTC()
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		t.Fatal(err)
	}
	today := at.In(loc).Format("2006-01-02")
	routes := []models.CapacityRoute{}
	target := -1
	for _, pair := range staticdata.CapacityPairs() {
		code := strings.ReplaceAll(pair, "-", "")
		routes = append(routes, models.CapacityRoute{
			RouteCode: code, DataState: models.DataState{DataStatus: "available", ServiceDate: today}, ObservedAt: at,
			Sailings: []models.CapacitySailing{{DepartureTime: "4:30 pm", VesselName: "Vessel"}},
		})
		if code == "TSANAN" {
			target = len(routes) - 1
		}
	}
	if target < 0 {
		t.Fatal("TSA-NAN missing from capacity catalog")
	}
	r := &routes[target]
	r.DataStatus, r.Sailings = "no-service", []models.CapacitySailing{}
	meta := buildResponseMeta(routes, nil, true, false)
	if meta.Stale || !meta.Complete || meta.Sources["currentConditions"].Status != "fresh" || len(meta.Issues) != 0 {
		t.Fatalf("fresh no-service day made fleet unhealthy: %#v", meta)
	}
	r.ObservedAt = at.Add(-11 * time.Minute)
	meta = buildResponseMeta(routes, nil, true, false)
	if !meta.Stale || meta.Complete {
		t.Fatal("old no-service observation must still expire")
	}
	r.ObservedAt = at
	r.ServiceDate = at.In(loc).AddDate(0, 0, -1).Format("2006-01-02")
	if buildResponseMeta(routes, nil, true, false).Complete {
		t.Fatal("yesterday's no-service observation treated as complete")
	}
	r.ServiceDate, r.DataStatus = today, "unavailable"
	if buildResponseMeta(routes, nil, true, false).Complete {
		t.Fatal("genuine source failure hidden")
	}
}

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

func TestFreshDoesNotMeanComplete(t *testing.T) {
	at := time.Now().UTC()
	r := []models.CapacityRoute{{RouteCode: "HSBNAN", ObservedAt: at, Sailings: []models.CapacitySailing{{DepartureTime: "", VesselName: ""}}}}
	meta := buildResponseMeta(r, nil, true, false)
	if meta.Stale || meta.Complete || len(meta.Issues) == 0 {
		t.Fatal(meta)
	}
}
func TestLegacyResponseHasNoNewFields(t *testing.T) {
	at := time.Now().UTC()
	data := AllDataResponse{CapacityRoutes: []models.CapacityRoute{{RouteCode: "TSASWB", FromTerminalCode: "TSA", ToTerminalCode: "SWB", SailingDuration: "1h 35m", Sailings: []models.CapacitySailing{{SailingDetails: models.SailingDetails{ServiceDate: "2026-09-11", ScheduledTime: "6:00 am", CapacityKnown: new(bool)}, DepartureTime: "6:00 am", SailingStatus: "future", Fill: 50, VesselName: "Vessel", ObservedAt: at}}}}}
	b, e := json.Marshal(ConvertV1ResponseToV2Response(data))
	if e != nil {
		t.Fatal(e)
	}
	want := `{"TSA":{"SWB":{"sailingDuration":"1h 35m","sailings":[{"time":"6:00 am","arrivalTime":"","isCancelled":false,"fill":50,"carFill":0,"oversizeFill":0,"vesselName":"Vessel","vesselStatus":""}]}}}`
	if string(b) != want {
		t.Fatalf("legacy JSON changed: %s", b)
	}
}

func TestLegacyNonCapacityCancellation(t *testing.T) {
	for _, from := range []string{"FUL", "BOW"} {
		to := "SWB"
		if from == "BOW" {
			to = "HSB"
		}
		for _, status := range []string{"cancelled", "scheduled", "delayed", ""} {
			t.Run(from+to+"/"+status, func(t *testing.T) {
				data := AllDataResponse{NonCapacityRoutes: []models.NonCapacityRoute{{
					FromTerminalCode: from, ToTerminalCode: to,
					Sailings: []models.NonCapacitySailing{{DepartureTime: "8:00 am", Status: status}},
				}}}
				b, err := json.Marshal(ConvertV1ResponseToV2Response(data))
				if err != nil {
					t.Fatal(err)
				}
				var legacy map[string]map[string]models.Route
				if err := json.Unmarshal(b, &legacy); err != nil {
					t.Fatal(err)
				}
				sailings := legacy[from][to].Sailings
				if len(sailings) != 1 || sailings[0].IsCancelled != (status == "cancelled") {
					t.Fatalf("wrong legacy cancellation: %s", b)
				}
			})
		}
	}
}
