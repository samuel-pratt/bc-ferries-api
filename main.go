package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron"
)

var sailings Response

func UpdateSchedule() {
	sailings = ScrapeCapacityRoutes()

	fmt.Print("Updated sailing data at: ")
	fmt.Println(time.Now())
}

func HealthCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonString, _ := json.Marshal("Server OK")

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func GetAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Print("/api/ call at: ")
	fmt.Println(time.Now())

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
			fmt.Print("/api/" + departureTerminal + "/ call at: ")
			fmt.Println(time.Now())

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
					fmt.Print("/api/" + departureTerminal + "/" + destinationTerminal + "/ call at: ")
					fmt.Println(time.Now())

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

	router := httprouter.New()

	router.GET("/healthcheck/", HealthCheck)
	router.GET("/api/", GetAll)
	router.GET("/api/:departureTerminal/", GetDepartureTerminal)
	router.GET("/api/:departureTerminal/:destinationTerminal/", GetDestinationTerminal)

	// Home page
	router.NotFound = http.FileServer(http.Dir("./static"))

	var port = os.Getenv("PORT")

	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "4747"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}

	http.ListenAndServe(":"+port, router)
}
