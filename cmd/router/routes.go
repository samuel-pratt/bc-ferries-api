package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
)

/**************/
/* V2 Structs */
/**************/

type AllDataResponse struct {
	CapacityRoutes    []models.CapacityRoute    `json:"capacityRoutes"`
	NonCapacityRoutes []models.NonCapacityRoute `json:"nonCapacityRoutes"`
}

type CapacityResponse struct {
	Routes []models.CapacityRoute `json:"routes"`
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
func GetCapacitySailings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routes := db.GetCapacitySailings()

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
func GetNonCapacitySailings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routes := db.GetNonCapacitySailings()

	response := models.NonCapacityResponse{
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
