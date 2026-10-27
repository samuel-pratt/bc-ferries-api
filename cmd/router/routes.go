package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"github.com/samuel-pratt/bc-ferries-api/cmd/staticdata"
	"strings"
)

/**************/
/* V2 Structs */
/**************/

type AllDataResponse struct {
	CapacityRoutes    []models.CapacityRoute    `json:"capacityRoutes"`
	NonCapacityRoutes []models.NonCapacityRoute `json:"nonCapacityRoutes"`
	Meta              models.ResponseMeta       `json:"meta"`
	SpecialServices   []models.SpecialService   `json:"specialServices"`
}

type CapacityResponse struct {
	Routes []models.CapacityRoute `json:"routes"`
	Meta   models.ResponseMeta    `json:"meta"`
}

/*************/
/* V2 Routes */
/*************/

/*
 * GetCapacityAndNonCapacitySailings
 *
 * Returns data for all capacity and non capacity routes
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func GetCapacityAndNonCapacitySailings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	capacityRoute := db.GetCapacitySailings()
	nonCapacityRoute := db.GetNonCapacitySailings()

	response := AllDataResponse{
		CapacityRoutes:    capacityRoute,
		NonCapacityRoutes: nonCapacityRoute,
		Meta:              buildResponseMeta(capacityRoute, nonCapacityRoute, true, true),
		SpecialServices:   db.GetSpecialServices(),
	}
	checkSpecialServices(&response.Meta, response.SpecialServices)
	checkCatalogCoverage(&response.Meta, capacityRoute, nonCapacityRoute, true, true)

	jsonString, _ := json.Marshal(response)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)

}

func buildResponseMeta(capacity []models.CapacityRoute, nonCapacity []models.NonCapacityRoute, requireCapacity, requireSchedule bool) models.ResponseMeta {
	now := time.Now().UTC()
	oldestCapacity := oldestCapacityObservation(capacity)
	oldestSchedule := oldestScheduleObservation(nonCapacity)
	stale := (requireCapacity && (oldestCapacity == nil || now.Sub(*oldestCapacity) > 10*time.Minute)) ||
		(requireSchedule && (oldestSchedule == nil || now.Sub(*oldestSchedule) > 8*time.Hour))
	sources := make(map[string]models.SourceHealth)
	if requireCapacity {
		sources["currentConditions"] = sourceHealth(oldestCapacity, now, 10*time.Minute)
	}
	if requireSchedule {
		sources["schedules"] = sourceHealth(oldestSchedule, now, 8*time.Hour)
	}
	issues := []string{}
	loc, _ := time.LoadLocation("America/Vancouver")
	today := now.In(loc).Format("2006-01-02")
	if requireCapacity {
		present := map[string]bool{}
		for _, r := range capacity {
			present[r.RouteCode] = true
			if r.DataStatus != "available" || len(r.Warnings) > 0 {
				issues = append(issues, "current conditions incomplete: "+r.RouteCode)
			}
			if r.ServiceDate != today {
				issues = append(issues, "current conditions service date is not today: "+r.RouteCode)
			}
			for _, s := range r.Sailings {
				if s.DepartureTime == "" || s.VesselName == "" {
					issues = append(issues, "invalid sailing: "+r.RouteCode)
					break
				}
			}
		}
		missing := 0
		for _, p := range staticdata.CapacityPairs() {
			if !present[strings.ReplaceAll(p, "-", "")] {
				missing++
			}
		}
		if missing > 0 {
			issues = append(issues, fmt.Sprintf("missing capacity routes: %d", missing))
		}
	}
	if requireSchedule {
		present := map[string]bool{}
		for _, r := range nonCapacity {
			present[r.RouteCode] = true
			if (r.DataStatus != "available" && r.DataStatus != "no-service") || len(r.Warnings) > 0 {
				issues = append(issues, "schedule unavailable: "+r.RouteCode)
			}
			if r.ServiceDate != today {
				issues = append(issues, "schedule service date is not today: "+r.RouteCode)
			}
			if r.CurrentConditions != nil && (r.CurrentConditions.DataStatus == "unavailable" || len(r.CurrentConditions.Warnings) > 0 || now.Sub(r.CurrentConditions.ObservedAt) > 10*time.Minute) {
				issues = append(issues, "current conditions unavailable: "+r.RouteCode)
			}
		}
		missing := 0
		for _, p := range staticdata.SchedulePairs() {
			if !present[strings.ReplaceAll(p, "-", "")] {
				missing++
			}
		}
		if missing > 0 {
			issues = append(issues, fmt.Sprintf("missing schedule routes: %d", missing))
		}
	}
	return models.ResponseMeta{
		SchemaVersion: "2.2",
		GeneratedAt:   now,
		Stale:         stale,
		Sources:       sources,
		Complete:      !stale && len(issues) == 0,
		Issues:        issues,
	}
}

func sourceHealth(observed *time.Time, now time.Time, maxAge time.Duration) models.SourceHealth {
	if observed == nil {
		return models.SourceHealth{Status: "unavailable"}
	}
	status := "fresh"
	if now.Sub(*observed) > maxAge {
		status = "stale"
	}
	return models.SourceHealth{Status: status, ObservedAt: observed, LastSuccessfulAt: observed}
}

func oldestCapacityObservation(routes []models.CapacityRoute) *time.Time {
	var oldest *time.Time
	for _, route := range routes {
		if route.ObservedAt.IsZero() {
			return nil
		}
		value := route.ObservedAt
		if oldest == nil || value.Before(*oldest) {
			oldest = &value
		}
	}
	return oldest
}

func oldestScheduleObservation(routes []models.NonCapacityRoute) *time.Time {
	var oldest *time.Time
	for _, route := range routes {
		if route.ObservedAt.IsZero() {
			return nil
		}
		value := route.ObservedAt
		if oldest == nil || value.Before(*oldest) {
			oldest = &value
		}
	}
	return oldest
}

/*
 * GetCapacitySailings
 *
 * Returns sailing data for all capacity routes
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func GetCapacitySailings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routes := db.GetCapacitySailings()

	response := CapacityResponse{
		Routes: routes,
		Meta:   buildResponseMeta(routes, nil, true, false),
	}
	checkCatalogCoverage(&response.Meta, routes, nil, true, false)

	usable := false
	for _, route := range routes {
		if len(route.Sailings) > 0 {
			usable = true
			break
		}
	}
	if !usable {
		jsonString, _ := json.Marshal("BC Ferries Data Currently Down")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonString)
	} else {
		jsonString, _ := json.Marshal(response)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonString)
	}
}

/*
 * GetNonCapacitySailings
 *
 * Returns sailing data for all non capacity routes
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func GetNonCapacitySailings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routes := db.GetNonCapacitySailings()

	response := models.NonCapacityResponse{
		Routes:          routes,
		Meta:            buildResponseMeta(nil, routes, false, true),
		SpecialServices: db.GetSpecialServices(),
	}
	checkSpecialServices(&response.Meta, response.SpecialServices)
	checkCatalogCoverage(&response.Meta, nil, routes, false, true)

	jsonString, _ := json.Marshal(response)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)

}

/**************/
/* V1 Structs */
/**************/

type Response struct {
	Schedule  map[string]map[string]models.Route `json:"schedule"`
	ScrapedAt time.Time                          `json:"scrapedAt"`
}

/*************/
/* V1 Routes */
/*************/
// V1 routes return data in a different format and only contain upcoming sailings for specific routes

/*
 * GetAllSailings
 *
 * Returns all sailing data
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func GetAllSailings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	capacityRoute := db.GetCapacitySailings()
	nonCapacityRoute := db.GetNonCapacitySailings()

	response := AllDataResponse{
		CapacityRoutes:    capacityRoute,
		NonCapacityRoutes: nonCapacityRoute,
	}

	jsonString, _ := json.Marshal(ConvertV1ResponseToV2Response(response))

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

/*
 * GetSailingsByDepartureTerminal
 *
 * Returns sailing data for given departure
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func GetSailingsByDepartureTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	departureTerminal := ps.ByName("departureTerminal")
	capacityRoute := db.GetCapacitySailings()
	nonCapacityRoute := db.GetNonCapacitySailings()

	allDataResponse := AllDataResponse{
		CapacityRoutes:    capacityRoute,
		NonCapacityRoutes: nonCapacityRoute,
	}

	jsonString, _ := json.Marshal(ConvertV1ResponseToV2Response(allDataResponse)[departureTerminal])

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

/*
 * GetSailingsByDepartureAndDestinationTerminals
 *
 * Returns sailing data for given departure and destination terminal
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func GetSailingsByDepartureAndDestinationTerminals(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	departureTerminal := ps.ByName("departureTerminal")
	destinationTerminal := ps.ByName("destinationTerminal")
	capacityRoute := db.GetCapacitySailings()
	nonCapacityRoute := db.GetNonCapacitySailings()

	allDataResponse := AllDataResponse{
		CapacityRoutes:    capacityRoute,
		NonCapacityRoutes: nonCapacityRoute,
	}

	jsonString, _ := json.Marshal(ConvertV1ResponseToV2Response(allDataResponse)[departureTerminal][destinationTerminal])

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

/****************/
/* Other Routes */
/****************/

/*
 * HealthCheck
 *
 * Returns a simple response indicating the server is running.
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func HealthCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonString, _ := json.Marshal("Server OK")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

/********************/
/* Helper Functions */
/********************/

/*
 * ConvertV1ResponseToV2Response
 *
 * Converts the V2 API response format into the legacy V1 structure,
 * organizing sailings by departure and destination terminals.
 *
 * Filters only allowed terminal pairs as defined by internal maps.
 *
 * @param AllDataResponse allData - the combined capacity and non-capacity data
 *
 * @return map[string]map[string]models.Route - nested route data
 */
func ConvertV1ResponseToV2Response(allData AllDataResponse) map[string]map[string]models.Route {
	schedule := make(map[string]map[string]models.Route)

	// Define the allowed terminal codes for CapacityRoutes and NonCapacityRoutes
	capacityRoutesFilter := map[string][]string{
		"TSA": {"SWB", "SGI", "DUK"},
		"SWB": {"TSA", "FUL", "SGI"},
		"HSB": {"NAN", "LNG", "BOW"},
		"DUK": {"TSA"},
		"LNG": {"HSB"},
		"NAN": {"HSB"},
	}

	nonCapacityRoutesFilter := map[string][]string{
		"FUL": {"SWB"},
		"BOW": {"HSB"},
	}

	for _, capRoute := range allData.CapacityRoutes {
		fromTerminal := capRoute.FromTerminalCode
		toTerminal := capRoute.ToTerminalCode

		if allowedDestinations, ok := capacityRoutesFilter[fromTerminal]; ok {
			if contains(allowedDestinations, toTerminal) {
				route := models.Route{
					SailingDuration: capRoute.SailingDuration,
					Sailings:        []models.Sailing{},
				}

				for _, capSailing := range capRoute.Sailings {
					if capSailing.SailingStatus == "future" || capSailing.SailingStatus == "cancelled" {
						route.Sailings = append(route.Sailings, models.Sailing{
							DepartureTime: capSailing.DepartureTime,
							ArrivalTime:   capSailing.ArrivalTime,
							IsCancelled:   capSailing.SailingStatus == "cancelled",
							Fill:          capSailing.Fill,
							CarFill:       capSailing.CarFill,
							OversizeFill:  capSailing.OversizeFill,
							VesselName:    capSailing.VesselName,
							VesselStatus:  capSailing.VesselStatus,
						})
					}
				}

				if len(route.Sailings) > 0 {
					if _, ok := schedule[fromTerminal]; !ok {
						schedule[fromTerminal] = make(map[string]models.Route)
					}
					schedule[fromTerminal][toTerminal] = route
				}
			}
		}
	}

	for _, nonCapRoute := range allData.NonCapacityRoutes {
		fromTerminal := nonCapRoute.FromTerminalCode
		toTerminal := nonCapRoute.ToTerminalCode

		if allowedDestinations, ok := nonCapacityRoutesFilter[fromTerminal]; ok {
			if contains(allowedDestinations, toTerminal) {
				route := models.Route{
					SailingDuration: nonCapRoute.SailingDuration,
					Sailings:        []models.Sailing{},
				}

				for _, nonCapSailing := range nonCapRoute.Sailings {
					route.Sailings = append(route.Sailings, models.Sailing{
						DepartureTime: nonCapSailing.DepartureTime,
						ArrivalTime:   nonCapSailing.ArrivalTime,
						IsCancelled:   nonCapSailing.Status == "cancelled",
						Fill:          0,
						CarFill:       0,
						OversizeFill:  0,
						VesselName:    nonCapSailing.VesselName,
						VesselStatus:  nonCapSailing.VesselStatus,
					})
				}

				if len(route.Sailings) > 0 {
					if _, ok := schedule[fromTerminal]; !ok {
						schedule[fromTerminal] = make(map[string]models.Route)
					}
					schedule[fromTerminal][toTerminal] = route
				}
			}
		}
	}

	return schedule
}

/*
 * contains
 *
 * Utility function to check if a string slice contains a given string.
 *
 * @param []string s - the slice to search
 * @param string str - the string to look for
 *
 * @return bool - true if str is found in s
 */
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Readiness checks database access as well as usable cached route data.
func ReadyCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	var usable bool
	err := db.Conn.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM capacity_routes WHERE jsonb_array_length(sailings)>0)").Scan(&usable)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil || !usable {
		w.WriteHeader(503)
		w.Write([]byte("{\"status\":\"unavailable\"}"))
		return
	}
	w.Write([]byte("{\"status\":\"ready\"}"))
}
