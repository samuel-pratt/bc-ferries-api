package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron"
	"github.com/tidwall/gjson"
)

func updateSchedule() {
	response := scraper()
	jsonString, _ := json.Marshal(response)
	err := ioutil.WriteFile("sailings.json", []byte(jsonString), 0644)
	if err != nil {
		sentry.CaptureException(err)
	}
	fmt.Print("Updated sailings.json at: ")
	fmt.Println(time.Now())
}

func getAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		sentry.CaptureException(err)
	}

	sentry.CaptureMessage("REQUEST: getAll()")

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getDepartureTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, err := ioutil.ReadFile("sailings.json")
	if err != nil {
		sentry.CaptureException(err)
	}

	sentry.CaptureMessage("REQUEST: getDepartureTerminal()")

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
		sentry.CaptureException(err)
	}

	sentry.CaptureMessage("REQUEST: getDestinationTerminal()")

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
	// dotenv setup
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup sentry
	err = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	// Schedule update every minute
	c := cron.New()
	c.AddFunc("@every 1m", updateSchedule)
	c.Start()

	// Router setup
	router := httprouter.New()

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
