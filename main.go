package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron"
	"github.com/tidwall/gjson"
)

func UpdateSchedule() {
	response := ScrapeCapacityRoutes()
	jsonString, _ := json.Marshal(response)
	err := ioutil.WriteFile("sailings.json", []byte(jsonString), 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("Updated sailings.json at: ")
	fmt.Println(time.Now())
}

func GetAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func GetDepartureTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		fmt.Println(err)
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
			schedule := gjson.Get(string(data), "schedule."+strings.ToUpper(departureTerminal))

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(schedule.String()))
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func GetDestinationTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		fmt.Println(err)
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
		{"SWB", "SGU", "DUK"},
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
					schedule := gjson.Get(string(data), "schedule."+strings.ToUpper(departureTerminal)+"."+strings.ToUpper(destinationTerminal))

					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(schedule.String()))
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

	// Root api call
	router.GET("/api/", GetAll)
	router.GET("/api/:departureTerminal/", GetDepartureTerminal)
	router.GET("/api/:departureTerminal/:destinationTerminal/", GetDestinationTerminal)

	// Home page
	router.NotFound = http.FileServer(http.Dir("./static"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000" // Default port if not specified
	}

	http.ListenAndServe(":"+port, router)
}
