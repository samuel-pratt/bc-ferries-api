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

func updateSchedule() {
	response := scraper()
	jsonString, _ := json.Marshal(response)
	err := ioutil.WriteFile("sailings.json", []byte(jsonString), 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("Updated sailings.json at: ")
	fmt.Println(time.Now())
}

func getAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getDepartureTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		fmt.Println(err)
	}

	departureTerminals := [6]string{
		"Tsawwassen",
		"Swartz-Bay",
		"Horseshoe-Bay",
		"Nanaimo-(Duke-pt)",
		"Langdale",
		"Nanaimo-(Dep-Bay)",
	}

	// Get url paramaters
	departureTerminal := ps.ByName("departureTerminal")

	// Find if departureTerminal is in departureTerminals
	for i := 0; i < len(departureTerminals); i++ {
		if strings.EqualFold(departureTerminal, departureTerminals[i]) {
			schedule := gjson.Get(string(data), "schedule."+strings.ToLower(departureTerminal))

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(schedule.String()))
		}
	}
}

func getDestinationTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		fmt.Println(err)
	}

	departureTerminals := [6]string{
		"Tsawwassen",
		"Swartz-Bay",
		"Horseshoe-Bay",
		"Nanaimo-(Duke-pt.)",
		"Langdale",
		"Nanaimo-(Dep-Bay)",
	}

	destinationTerminals := [6][]string{
		{"Swartz-Bay", "Southern-Gulf-Islands", "Nanaimo-(Duke-pt)"},
		{"Tsawwassen", "Fulford-Habrbour-(Saltspring)", "Southern-Gulf-Islands"},
		{"Nanaimo-(Dep-Bay)", "Langdale", "Snug-Cove-(Bowen)"},
		{"Tsawwassen"},
		{"Horseshoe-Bay"},
		{"Horseshoe-Bay"},
	}

	// Get url paramaters
	departureTerminal := ps.ByName("departureTerminal")
	destinationTerminal := ps.ByName("destinationTerminal")

	// Find if departureTerminal is in departureTerminals
	for i := 0; i < len(departureTerminals); i++ {
		if strings.EqualFold(departureTerminal, departureTerminals[i]) {
			for j := 0; j < len(destinationTerminals[j]); j++ {
				schedule := gjson.Get(string(data), "schedule."+strings.ToLower(departureTerminal)+"."+strings.ToLower(destinationTerminal))

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(schedule.String()))
			}
		}
	}
}

func main() {
	// Schedule update every hour
	c := cron.New()
	c.AddFunc("@every 1m", updateSchedule)
	c.Start()

	router := httprouter.New()

	// Root api call
	router.GET("/api/", getAll)
	router.GET("/api/:departureTerminal/", getDepartureTerminal)
	router.GET("/api/:departureTerminal/:destinationTerminal/", getDestinationTerminal)

	// Home page
	router.NotFound = http.FileServer(http.Dir("./static"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000" // Default port if not specified
	}

	http.ListenAndServe(":"+port, router)
}
