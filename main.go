package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/robfig/cron"
)

var sailings Response

func UpdateSchedule() {
	sailings = ScrapeCapacityRoutes()

	fmt.Print("Updated sailing data at: ")
	fmt.Println(time.Now())
}

func GetAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonString, _ := json.Marshal(sailings)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func GetDepartureTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	departureTerminals := [6]string{
		"TSA",
		"SWB",
		"HSB",
		"DUK",
		"LNG",
		"NAN",
	}

	// Get url paramaters
	departureTerminal := ps.ByName("departureTerminal")

	// Find if departureTerminal is in departureTerminals
	for i := 0; i < len(departureTerminals); i++ {
		if strings.EqualFold(departureTerminal, departureTerminals[i]) {
			schedule := sailings.Schedule[departureTerminal]

			jsonString, _ := json.Marshal(schedule)

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonString)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func GetDestinationTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	departureTerminals := [6]string{
		"TSA",
		"SWB",
		"HSB",
		"DUK",
		"LNG",
		"NAN",
	}

	destinationTerminals := [6][]string{
		{"SWB", "SGI", "DUK"},
		{"TSA", "FUL", "SGI"},
		{"NAN", "LNG", "BOW"},
		{"TSA"},
		{"HSB"},
		{"HSB"},
	}

	// Get url paramaters
	departureTerminal := ps.ByName("departureTerminal")
	destinationTerminal := ps.ByName("destinationTerminal")

	// Find if departureTerminal is in departureTerminals
	for i := 0; i < len(departureTerminals); i++ {
		if strings.EqualFold(departureTerminal, departureTerminals[i]) {
			for j := 0; j < len(destinationTerminals[i]); j++ {
				if strings.EqualFold(destinationTerminal, destinationTerminals[i][j]) {
					schedule := sailings.Schedule[departureTerminal][destinationTerminal]

					jsonString, _ := json.Marshal(schedule)

					w.Header().Set("Content-Type", "application/json")
					w.Write(jsonString)
					return
				}
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func main() {
	// Create new schedule at startup
	UpdateSchedule()

	// Schedule update every hour
	c := cron.New()
	c.AddFunc("@every 1m", UpdateSchedule)
	c.Start()

	// router := httprouter.New()

	// // Root api call
	// router.GET("/api/", GetAll)
	// router.GET("/api/:departureTerminal/", GetDepartureTerminal)
	// router.GET("/api/:departureTerminal/:destinationTerminal/", GetDestinationTerminal)

	// // Home page
	// router.NotFound = http.FileServer(http.Dir("./static"))

	// var port = os.Getenv("PORT")

	// if port == "" {
	// 	log.Fatal("$PORT must be set")
	// }

	// http.ListenAndServe(":"+port, router)

	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/",
			Handler: func(c echo.Context) error {

				return c.String(200, "hello")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireGuestOnly(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api",
			Handler: func(c echo.Context) error {
				jsonString, _ := json.Marshal(sailings)
				return c.JSONBlob(200, jsonString)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireGuestOnly(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/healthcheck",
			Handler: func(c echo.Context) error {
				return c.String(200, "OK")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireGuestOnly(),
			},
		})

		return nil
	})

	if err := app.Start(); err != nil {
		fmt.Println("hello")
		log.Fatal(err)
	}
}
