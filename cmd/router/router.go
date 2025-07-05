package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

/*
 * SetupRouter
 *
 * Initializes the HTTP router and registers all API endpoints.
 * Also serves static files for not-found routes.
 *
 * @return *httprouter.Router - configured router instance
 */
func SetupRouter() *httprouter.Router {
	router := httprouter.New()

	// V2 Routes
	router.GET("/v2/", GetCapacityAndNonCapacitySailings)
	router.GET("/v2/capacity/", GetCapacitySailings)
	router.GET("/v2/noncapacity/", GetNonCapacitySailings)

	// V1 Routes
	router.GET("/api/", GetAllSailings)
	router.GET("/api/:departureTerminal/", GetSailingsByDepartureTerminal)
	router.GET("/api/:departureTerminal/:destinationTerminal/", GetSailingsByDepartureAndDestinationTerminals)

	router.GET("/healthcheck/", HealthCheck)

	router.NotFound = http.FileServer(http.Dir("./static"))

	return router
}
