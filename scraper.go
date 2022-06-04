package main

import (
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
	if strings.Contains(stringToCheck, "Departures") && strings.Contains(stringToCheck, "Status") && strings.Contains(stringToCheck, "Details") {
		return false
	}

	if strings.Contains(stringToCheck, "Arrived:") || strings.Contains(stringToCheck, "ETA:") {
		return false
	}

	if strings.Contains(stringToCheck, "Cancelled") {
		return false
	}

	if strings.Contains(stringToCheck, "...") {
		return false
	}

	return true
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
			document.Find(".detail-departure-table").Each(func(index int, table *goquery.Selection) {

				sailing := Sailing{
					Time:         "",
					Fill:         0,
					CarFill:      0,
					OversizeFill: 0,
					VesselName:   "",
					VesselStatus: "",
				}

				sailingIndex := 0

				// Get every row of table
				table.Find("tr").Each(func(indextr int, row *goquery.Selection) {
					if ContainsSailingData(row.Text()) {
						// Sailing duration
						if strings.Contains(row.Text(), "Sailing duration:") {
							row.Find("b").Each(func(indexb int, sailingTime *goquery.Selection) {
								route.SailingDuration = strings.TrimSpace(sailingTime.Text())
							})
						} else {
							if strings.Contains(row.Text(), " AM") || strings.Contains(row.Text(), " PM") {
								// Time and fill
								row.Find("td").Each(func(indextd int, tableData *goquery.Selection) {

									if indextd == 0 {
										// Time
										time := strings.TrimSpace(tableData.Text())
										if len(time) > 7 {
											time = time[0:8]
										}

										time = strings.TrimSpace(strings.ReplaceAll(time, " ", ""))

										sailing.Time = time
									} else if indextd == 1 {
										// Fill
										fill := strings.TrimSpace(tableData.Text())

										if fill == "Full" {
											sailing.Fill = 100
										} else {
											fill, err := strconv.Atoi(strings.Split(fill, "%")[0])
											if err == nil {
												sailing.Fill = 100 - fill
											}
										}
									}
								})
							} else {
								// Vessel name, car fill, oversize fill

								row.Find(".sailing-ferry-name").Each(func(indexname int, tableData *goquery.Selection) {
									sailing.VesselName = strings.TrimSpace(tableData.Text())
								})

								row.Find(".progress-bar").Each(func(indexprogressbar int, tableData *goquery.Selection) {
									if indexprogressbar == 0 {
										return
									}

									fillString := strings.TrimSpace(tableData.Text())
									fill := 0

									if fillString == "FULL" {
										fill = 100
									} else {
										fillInt, err := strconv.Atoi(strings.Split(fillString, "%")[0])
										if err == nil {
											fill = 100 - fillInt
										}
									}

									if indexprogressbar == 1 {
										sailing.CarFill = fill
									} else if indexprogressbar == 2 {
										sailing.OversizeFill = fill
									} else {
										return
									}

								})

								// Add sailing to route
								if sailing.Time != "" {
									route.Sailings = append(route.Sailings, sailing)
								}

								// Reset sailing to default
								sailing = Sailing{
									Time:         "",
									Fill:         0,
									CarFill:      0,
									OversizeFill: 0,
									VesselName:   "",
									VesselStatus: "",
								}
							}
							sailingIndex++
						}

					} else {
						return
					}
				})
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
