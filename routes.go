package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

/**************/
/* V2 Structs */
/**************/

type AllDataResponse struct {
	CapacityRoutes    []CapacityRoute    `json:"capacityRoutes"`
	NonCapacityRoutes []NonCapacityRoute `json:"nonCapacityRoutes"`
}

type CapacityResponse struct {
	Routes []CapacityRoute `json:"routes"`
}

type CapacityRoute struct {
	RouteCode        string            `json:"routeCode"`
	FromTerminalCode string            `json:"fromTerminalCode"`
	ToTerminalCode   string            `json:"toTerminalCode"`
	SailingDuration  string            `json:"sailingDuration"`
	Sailings         []CapacitySailing `json:"sailings"`
}

type CapacitySailing struct {
	DepartureTime string `json:"time"`
	ArrivalTime   string `json:"arrivalTime"`
	SailingStatus string `json:"sailingStatus"`
	Fill          int    `json:"fill"`
	CarFill       int    `json:"carFill"`
	OversizeFill  int    `json:"oversizeFill"`
	VesselName    string `json:"vesselName"`
	VesselStatus  string `json:"vesselStatus"`
}

type NonCapacityResponse struct {
	Routes []NonCapacityRoute `json:"routes"`
}

type NonCapacityRoute struct {
	RouteCode        string               `json:"routeCode"`
	FromTerminalCode string               `json:"fromTerminalCode"`
	ToTerminalCode   string               `json:"toTerminalCode"`
	SailingDuration  string               `json:"sailingDuration"`
	Sailings         []NonCapacitySailing `json:"sailings"`
}

type NonCapacitySailing struct {
	DepartureTime string `json:"time"`
	ArrivalTime   string `json:"arrivalTime"`
	VesselName    string `json:"vesselName"`
	VesselStatus  string `json:"vesselStatus"`
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
func CapacityAndNonCapacitySailingsEndpoint(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	capacityRoute := GetCapacitySailings()
	nonCapacityRoute := GetNonCapacitySailings()

	response := AllDataResponse{
		CapacityRoutes:    capacityRoute,
		NonCapacityRoutes: nonCapacityRoute,
	}

	jsonString, _ := json.Marshal(response)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)

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
func CapacitySailingsEndpoint(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routes := GetCapacitySailings()

	response := CapacityResponse{
		Routes: routes,
	}

	if len(response.Routes[0].Sailings) == 0 {
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
func NonCapacitySailingsEndpoint(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routes := GetNonCapacitySailings()

	response := NonCapacityResponse{
		Routes: routes,
	}

	jsonString, _ := json.Marshal(response)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)

}

/**************/
/* V1 Structs */
/**************/

type Response struct {
	Schedule  map[string]map[string]Route `json:"schedule"`
	ScrapedAt time.Time                   `json:"scrapedAt"`
}

type Route struct {
	SailingDuration string    `json:"sailingDuration"`
	Sailings        []Sailing `json:"sailings"`
}

type Sailing struct {
	DepartureTime string `json:"time"`
	ArrivalTime   string `json:"arrivalTime"`
	IsCancelled   bool   `json:"isCancelled"`
	Fill          int    `json:"fill"`
	CarFill       int    `json:"carFill"`
	OversizeFill  int    `json:"oversizeFill"`
	VesselName    string `json:"vesselName"`
	VesselStatus  string `json:"vesselStatus"`
}

/*************/
/* V1 Routes */
/*************/
// V1 routes return data in a different format and only contain upcoming sailings for specific routes

/*
 * AllSailingsEndpoint
 *
 * Returns all sailing data
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func AllSailingsEndpoint(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	capacityRoute := GetCapacitySailings()
	nonCapacityRoute := GetNonCapacitySailings()

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
 * SailingsByDepartureTerminal
 *
 * Returns sailing data for given departure
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func SailingsByDepartureTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	departureTerminal := ps.ByName("departureTerminal")
	capacityRoute := GetCapacitySailings()
	nonCapacityRoute := GetNonCapacitySailings()

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
 * SailingsByDepartureAndDestinationTerminals
 *
 * Returns sailing data for given departure and destination terminal
 *
 * @param http.ResponseWriter w
 * @param *http.Request r
 * @param httprouter.Params ps
 *
 * @return void
 */
func SailingsByDepartureAndDestinationTerminals(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	departureTerminal := ps.ByName("departureTerminal")
	destinationTerminal := ps.ByName("destinationTerminal")
	capacityRoute := GetCapacitySailings()
	nonCapacityRoute := GetNonCapacitySailings()

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

func HealthCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonString, _ := json.Marshal("Server OK")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

/********************/
/* Helper Functions */
/********************/

func ConvertV1ResponseToV2Response(allData AllDataResponse) map[string]map[string]Route {
	schedule := make(map[string]map[string]Route)

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
				route := Route{
					SailingDuration: capRoute.SailingDuration,
					Sailings:        []Sailing{},
				}

				for _, capSailing := range capRoute.Sailings {
					if capSailing.SailingStatus == "future" {
						route.Sailings = append(route.Sailings, Sailing{
							DepartureTime: capSailing.DepartureTime,
							ArrivalTime:   capSailing.ArrivalTime,
							IsCancelled:   capSailing.SailingStatus == "Cancelled",
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
						schedule[fromTerminal] = make(map[string]Route)
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
				route := Route{
					SailingDuration: nonCapRoute.SailingDuration,
					Sailings:        []Sailing{},
				}

				for _, nonCapSailing := range nonCapRoute.Sailings {
					route.Sailings = append(route.Sailings, Sailing{
						DepartureTime: nonCapSailing.DepartureTime,
						ArrivalTime:   nonCapSailing.ArrivalTime,
						IsCancelled:   false,
						Fill:          0,
						CarFill:       0,
						OversizeFill:  0,
						VesselName:    nonCapSailing.VesselName,
						VesselStatus:  nonCapSailing.VesselStatus,
					})
				}

				if len(route.Sailings) > 0 {
					if _, ok := schedule[fromTerminal]; !ok {
						schedule[fromTerminal] = make(map[string]Route)
					}
					schedule[fromTerminal][toTerminal] = route
				}
			}
		}
	}

	return schedule
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
