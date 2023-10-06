package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Route struct {
	SailingDuration string    `json:"sailingDuration"`
	Sailings        []Sailing `json:"sailings"`
}

type Sailing struct {
	DepartureTime string `json:"time"`
	ArrivalTime   string `json:"arrivalTime"`
	IsCancelled   bool   `json:"isCancelled"`
	Fill          int    `json:"fill"`
	CarFill       int    `json:"carFill"`
	OversizeFill  int    `json:"oversizeFill"`
	VesselName    string `json:"vesselName"`
	VesselStatus  string `json:"vesselStatus"`
}

func MakeCurrentConditionsLink(departure, destination string) string {
	return "https://www.bcferries.com/current-conditions/" + departure + "-" + destination
}

func MakeScheduleLink(departure, destination string) string {
	return "https://www.bcferries.com/routes-fares/schedules/daily/" + departure + "-" + destination
}

func ContainsSailingData(stringToCheck string) bool {
	if strings.Contains(stringToCheck, "%") || strings.Contains(stringToCheck, "FULL") || strings.Contains(stringToCheck, "Cancelled") || strings.Contains(stringToCheck, "Full") {
		return true
	}

	return false
}

func ScrapeRoutes(localMode bool) map[string]map[string]Route {
	departureTerminals := GetDepartureTerminals()

	destinationTerminals := GetDestinationTerminals()

	var schedule = make(map[string]map[string]Route)

	for i := 0; i < len(departureTerminals); i++ {
		schedule[departureTerminals[i]] = make(map[string]Route)

		for j := 0; j < len(destinationTerminals[i]); j++ {
			var document *goquery.Document

			if localMode == true {
				file, err := os.OpenFile("./sample-site.html", os.O_RDWR, 0644)
				if err != nil {
					log.Fatal("Local file read failed")
				}

				var response io.Reader = (file)

				// Create a goquery document from the HTTP response
				document, err = goquery.NewDocumentFromReader(response)
				if err != nil {
					log.Fatal("Error loading HTTP response body. ", err)
				}
			} else {
				link := ""
				if departureTerminals[i] == "FUL" || departureTerminals[i] == "BOW" {
					link = MakeScheduleLink(departureTerminals[i], destinationTerminals[i][j])
				} else {
					link = MakeCurrentConditionsLink(departureTerminals[i], destinationTerminals[i][j])
				}

				// Make HTTP GET request
				client := &http.Client{}
				req, err := http.NewRequest("GET", link, nil)
				req.Header.Add("User-Agent", "Mozilla")
				response, err := client.Do(req)

				if err != nil {
					log.Fatal(err)
				}

				defer response.Body.Close()

				document, err = goquery.NewDocumentFromReader(response.Body)
				if err != nil {
					log.Fatal("Error loading HTTP response body. ", err)
				}
			}

			if departureTerminals[i] == "FUL" || departureTerminals[i] == "BOW" {
				route := ScrapeNonCapacityRoute(document)

				schedule[departureTerminals[i]][destinationTerminals[i][j]] = route
			} else {
				route := ScrapeCapacityRoute(document)

				schedule[departureTerminals[i]][destinationTerminals[i][j]] = route
			}
		}
	}

	return schedule
}

func ScrapeCapacityRoute(document *goquery.Document) Route {
	route := Route{
		SailingDuration: "",
		Sailings:        []Sailing{},
	}

	// Get table of times and capacities
	document.Find(".mobile-friendly-row").Each(func(index int, sailingData *goquery.Selection) {

		sailing := Sailing{}

		if ContainsSailingData(sailingData.Text()) {
			// TIME AND VESSEL NAME
			timeAndBoatName := sailingData.Find(".mobile-paragraph").First().Text()
			timeAndBoatNameArray := strings.Split(timeAndBoatName, "\n")

			isTomorrow := false
			if strings.Contains(timeAndBoatName, "Tomorrow") {
				isTomorrow = true
			}

			for i := 0; i < len(timeAndBoatNameArray); i++ {
				item := strings.TrimSpace(timeAndBoatNameArray[i])
				item = strings.ReplaceAll(item, "\n", "")

                                if strings.Contains(strings.ToLower(item), "am") || strings.Contains(strings.ToLower(item), "pm") {
					sailing.DepartureTime = item
				} else if !strings.Contains(item, "Tomorrow") && len(item) > 5 {
					sailing.VesselName = item
				}
			}

			// FILL
			if isTomorrow {
				sailingData.Find(".cc-message-updates").Each(func(index int, tomorrowFillData *goquery.Selection) {
					fill := strings.TrimSpace(tomorrowFillData.Text())
					if index == 0 {
						if fill == "FULL" || fill == "Full" {
							sailing.Fill = 100
							sailing.IsCancelled = false
						} else if strings.Contains(fill, "Cancelled") {
							sailing.Fill = 0
							sailing.IsCancelled = true
						} else {
							fill, err := strconv.Atoi(strings.Split(fill, "%")[0])
							if err == nil {
								sailing.Fill = 100 - fill
							}

							sailing.IsCancelled = false
						}
					} else if index == 1 {
						tomorrowFillData.Find(".pcnt").Each(func(index int, tomorrowDetailedFillData *goquery.Selection) {
							if index == 0 {
								fill := strings.TrimSpace(tomorrowFillData.Text())

								if fill == "FULL" || fill == "Full" {
									sailing.CarFill = 100
								} else {
									fill, err := strconv.Atoi(strings.Split(fill, "%")[0])
									if err == nil {
										sailing.CarFill = 100 - fill
									}
								}
							} else if index == 1 {
								fill := strings.TrimSpace(tomorrowFillData.Text())

								if fill == "FULL" || fill == "Full" {
									sailing.OversizeFill = 100
								} else {
									fill, err := strconv.Atoi(strings.Split(fill, "%")[0])
									if err == nil {
										sailing.OversizeFill = 100 - fill
									}
								}
							}
						})
					}
				})
			} else {
				fill := strings.TrimSpace(sailingData.Find(".cc-percentage").First().Text())
				if fill == "FULL" || fill == "Full" {
					sailing.Fill = 100
					sailing.IsCancelled = false
				} else if strings.Contains(fill, "Cancelled") {
					sailing.Fill = 0
					sailing.IsCancelled = true
				} else {
					fill, err := strconv.Atoi(strings.Split(fill, "%")[0])
					if err == nil {
						sailing.Fill = 100 - fill
					}

					sailing.IsCancelled = false
				}
			}

			// FILL BREAKDOWN
			sailingData.Find(".pcnt").Each(func(detailedFillIndex int, detailedFill *goquery.Selection) {
				fill := strings.TrimSpace(detailedFill.Text())
				fillResult := 0

				if fill == "FULL" || fill == "Full" {
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

	return route
}

func ScrapeNonCapacityRoute(document *goquery.Document) Route {
	route := Route{
		SailingDuration: "",
		Sailings:        []Sailing{},
	}

	document.Find(".table-seasonal-schedule").First().Find("tbody").First().Find(".schedule-table-row").Each(func(index int, sailingData *goquery.Selection) {
		sailing := Sailing{}

		sailingData.Find("td").Each(func(index int, sailingData *goquery.Selection) {
			if index == 1 {
				sailing.DepartureTime = strings.TrimSpace(sailingData.Text())
			} else if index == 2 {
				sailing.ArrivalTime = strings.TrimSpace(sailingData.Text())
			}
		})

		route.Sailings = append(route.Sailings, sailing)
	})

	return route
}
