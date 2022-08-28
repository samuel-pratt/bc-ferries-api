package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Response struct {
	Schedule  map[string]map[string]Route `json:"schedule"`
	ScrapedAt time.Time                   `json:"scrapedAt"`
}

type Route struct {
	SailingDuration string    `json:"sailingDuration"`
	Sailings        []Sailing `json:"sailings"`
}

type Sailing struct {
	Time         string `json:"time"`
	Fill         int    `json:"fill"`
	CarFill      int    `json:"carFill"`
	OversizeFill int    `json:"oversizeFill"`
	VesselName   string `json:"vesselName"`
	VesselStatus string `json:"vesselStatus"`
}

func MakeCurrentConditionsLink(departure, destination string) string {
	return "https://www.bcferries.com/current-conditions/" + departure + "-" + destination
}

/*
 *	ContainsSailingData()
 *
 *	Helper function to determine if a string should be read or skipped.
 * 	Works by checking if string contains various sets of strings that denote no useful data here.
 */
func ContainsSailingData(stringToCheck string) bool {
	// if strings.Contains(stringToCheck, "Departures") && strings.Contains(stringToCheck, "Status") && strings.Contains(stringToCheck, "Details") {
	// 	return false
	// }

	// if strings.Contains(stringToCheck, "Arrived:") || strings.Contains(stringToCheck, "ETA:") {
	// 	return false
	// }

	// if strings.Contains(stringToCheck, "Cancelled") {
	// 	return false
	// }

	// if strings.Contains(stringToCheck, "...") {
	// 	return false
	// }

	// return true

	if strings.Contains(stringToCheck, "%") || strings.Contains(stringToCheck, "Full") {
		return true
	}

	return false
}

func ScrapeCapacityRoutes() Response {
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

	var schedule = make(map[string]map[string]Route)

	for i := 0; i < len(departureTerminals); i++ {
		schedule[departureTerminals[i]] = make(map[string]Route)

		for j := 0; j < len(destinationTerminals[i]); j++ {
			// Make HTTP GET request
			response, err := http.Get(MakeCurrentConditionsLink(departureTerminals[i], destinationTerminals[i][j]))
			if err != nil {
				log.Fatal(err)
			}
			defer response.Body.Close()

			// For local testing, make sure to change "response.Body" to "response" below
			// file, err := os.OpenFile("./sample-site.html", os.O_RDWR, 0644)
			// if err != nil {
			// 	log.Fatal("failed")
			// }

			// var response io.Reader = (file)

			// Create a goquery document from the HTTP response
			document, err := goquery.NewDocumentFromReader(response.Body)
			if err != nil {
				log.Fatal("Error loading HTTP response body. ", err)
			}

			route := Route{
				SailingDuration: "",
				Sailings:        []Sailing{},
			}

			// Get table of times and capacities
			document.Find(".mobile-friendly-row").Each(func(index int, sailingData *goquery.Selection) {

				sailing := Sailing{
					Time:         "",
					Fill:         0,
					CarFill:      0,
					OversizeFill: 0,
					VesselName:   "",
					VesselStatus: "",
				}

				if ContainsSailingData(sailingData.Text()) {
					// TIME AND VESSEL NAME
					timeAndBoatName := sailingData.Find(".mobile-paragraph").First().Text()
					timeAndBoatNameArray := strings.Split(timeAndBoatName, "\n")

					for i := 0; i < len(timeAndBoatNameArray); i++ {
						item := strings.TrimSpace(timeAndBoatNameArray[i])
						item = strings.ReplaceAll(item, "\n", "")

						if strings.Contains(item, "AM") || strings.Contains(item, "PM") {
							sailing.Time = item
						} else if !strings.Contains(item, "Tomorrow") && len(item) > 5 {
							sailing.VesselName = item
						}
					}

					// FILL
					fill := strings.TrimSpace(sailingData.Find(".cc-percentage").First().Text())
					fmt.Println(fill)
					if fill == "Full" || fill == "100%" || strings.Contains(fill, "100") {
						sailing.Fill = 100
					} else {
						fill, err := strconv.Atoi(strings.Split(fill, "%")[0])
						if err == nil {
							sailing.Fill = 100 - fill
						}
					}

					// FILL BREAKDOWN
					sailingData.Find(".pcnt").Each(func(detailedFillIndex int, detailedFill *goquery.Selection) {
						fill = strings.TrimSpace(detailedFill.Text())
						fillResult := 0

						if fill == "FULL" {
							fillResult = 100
						} else {
							fill, err := strconv.Atoi(strings.Split(fill, "%")[0])
							if err == nil {
								fillResult = 100 - fill
							}
						}

						if detailedFillIndex == 0 {
							sailing.CarFill = fillResult
						} else if detailedFillIndex == 1 {
							sailing.OversizeFill = fillResult
						}
					})

					if sailing.CarFill == 0 && sailing.OversizeFill == 0 && sailing.Fill == 100 {
						sailing.CarFill = 100
						sailing.OversizeFill = 100
					}

					route.Sailings = append(route.Sailings, sailing)
				}
			})

			schedule[departureTerminals[i]][destinationTerminals[i][j]] = route
		}
	}

	// Add timestamp to data
	currentTime := time.Now()

	// Add schedule and timestamp to response object
	response := Response{
		Schedule:  schedule,
		ScrapedAt: currentTime,
	}

	return response
}
