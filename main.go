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
var isSiteDown bool

func UpdateSchedule() {
	sailings = ScrapeCapacityRoutes()

	// No reason for checking these sailings specifically, just acts as a check for if the site is down
	// When BC Ferries is down all sailigns will be empty arrays but it seems excessive to check every single one
	if len(sailings.Schedule["TSA"]["SWB"].Sailings) == 0 && len(sailings.Schedule["SWB"]["TSA"].Sailings) == 0 && len(sailings.Schedule["HSB"]["NAN"].Sailings) == 0 {
		isSiteDown = true
	} else {
		isSiteDown = false
	}

	fmt.Print("Updated sailing data at: ")
	fmt.Println(time.Now())
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func HealthCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonString, _ := json.Marshal("Server OK")

	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func GetAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	enableCors(&w)

	if isSiteDown == true {
		w.Header().Set("Content-Type", "application/json")
		w.Write(nil)
		return
	}

	fmt.Print("/api/ call at: ")
	fmt.Println(time.Now())

	jsonString, _ := json.Marshal(sailings)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func GetDepartureTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	enableCors(&w)

	if isSiteDown == true {
		w.Header().Set("Content-Type", "application/json")
		w.Write(nil)
		return
	}

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
	enableCors(&w)

	if isSiteDown == true {
		w.Header().Set("Content-Type", "application/json")
		w.Write(nil)
		return
	}

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
