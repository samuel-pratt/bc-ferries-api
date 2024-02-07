package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/log"
)

/*
 * MakeCurrentConditionsLink
 *
 * Makes a link to the current conditions page for a given departure and destination
 *
 * @param string departure
 * @param string destination
 *
 * @return string
 */
func MakeCurrentConditionsLink(departure, destination string) string {
	return "https://www.bcferries.com/current-conditions/" + departure + "-" + destination
}

/*
 * MakeScheduleLink
 *
 * Makes a link to the schedule page for a given departure and destination
 *
 * @param string departure
 * @param string destination
 *
 * @return string
 */
func MakeScheduleLink(departure, destination string) string {
	return "https://www.bcferries.com/routes-fares/schedules/daily/" + departure + "-" + destination
}

/*
 * ScrapeBCCapacityRoutes
 *
 * Scrapes BC Ferries capacity routes
 *
 * @return void
 */
func ScrapeCapacityRoutes() {
	departureTerminals := GetCapacityDepartureTerminals()

	destinationTerminals := GetCapacityDestinationTerminals()

	for i := 0; i < len(departureTerminals); i++ {
		for j := 0; j < len(destinationTerminals[i]); j++ {
			link := MakeCurrentConditionsLink(departureTerminals[i], destinationTerminals[i][j])

			// Make HTTP GET request
			client := &http.Client{}
			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				log.Error(err)
			}

			req.Header.Add("User-Agent", "Mozilla")
			response, err := client.Do(req)
			if err != nil {
				log.Error(err)
			}

			defer response.Body.Close()

			document, err := goquery.NewDocumentFromReader(response.Body)
			if err != nil {
				log.Error(err)
			}

			ScrapeCapacityRoute(document, departureTerminals[i], destinationTerminals[i][j])
		}
	}
}

/*
 * ScrapeCapacityRoute
 *
 * Scrapes BC Ferries capacity data for a given route
 *
 * @param *goquery.Document document
 * @param string fromTerminalCode
 * @param string toTerminalCode
 *
 * @return void
 */
func ScrapeCapacityRoute(document *goquery.Document, fromTerminalCode string, toTerminalCode string) {
	db := GetPostgresInstance()

	route := CapacityRoute{
		RouteCode:        fromTerminalCode + toTerminalCode,
		ToTerminalCode:   toTerminalCode,
		FromTerminalCode: fromTerminalCode,
		Sailings:         []CapacitySailing{},
	}

	document.Find(".cc-status-td ").Each(func(index int, sailingData *goquery.Selection) {
		sailing := CapacitySailing{}

		sailingDataString := strings.TrimSpace(sailingData.Text())
		sailingDataString = strings.ReplaceAll(sailingDataString, "\t", "")

		sailingDataArray := strings.Split(sailingDataString, "\n")
		var reducedArray []string

		for i := range sailingDataArray {
			if strings.TrimSpace(sailingDataArray[i]) != "" {
				reducedArray = append(reducedArray, strings.TrimSpace(sailingDataArray[i]))
			}
		}
		compareString := strings.ToLower(strings.Join(reducedArray, " |"))

		if !strings.Contains(compareString, "dangerous goods only") {
			if len(reducedArray) == 3 && reducedArray[2] == "..." || strings.Contains(compareString, "Departed") {
				sailing.SailingStatus = "current"
				sailing.DepartureTime = reducedArray[0]
				sailing.VesselName = reducedArray[1]
				sailing.ArrivalTime = reducedArray[2]
			} else if strings.Contains(compareString, "arrived") {
				sailing.SailingStatus = "past"
				if len(reducedArray) >= 5 {
					sailing.DepartureTime = reducedArray[2]
					sailing.VesselName = reducedArray[3]
					sailing.ArrivalTime = strings.Split(reducedArray[4], " ")[1] + " " + strings.Split(reducedArray[4], " ")[2]
				}
			} else if strings.Contains(compareString, "...") {
				sailing.SailingStatus = "past"

				sailing.DepartureTime = reducedArray[2]
				sailing.VesselName = reducedArray[1]

				if len(reducedArray) >= 5 {

					sailing.ArrivalTime = reducedArray[4]
				}
			} else if strings.Contains(compareString, "eta") {
				sailing.SailingStatus = "current"

				if len(reducedArray) >= 7 {
					sailing.DepartureTime = reducedArray[2]
					sailing.VesselName = reducedArray[3]
					sailing.ArrivalTime = reducedArray[6]
				}
			} else if strings.Contains(compareString, "cancelled") {
				sailing.SailingStatus = "cancelled"

				sailing.DepartureTime = reducedArray[0]
				sailing.VesselName = reducedArray[1]
			} else if strings.Contains(compareString, "%") || strings.Contains(compareString, "full") {
				sailing.SailingStatus = "future"

				sailing.DepartureTime = reducedArray[0]
				sailing.VesselName = reducedArray[1]
				sailing.ArrivalTime = "none"

				if strings.Contains(strings.Join(reducedArray, ""), "Delayed") && len(reducedArray) == 5 && strings.Contains(strings.Join(reducedArray, ""), "Full") {
					sailing.Fill = 100
					sailing.CarFill = 100
					sailing.OversizeFill = 100
				} else if len(reducedArray) == 3 || len(reducedArray) == 4 {
					if strings.Contains(strings.Join(reducedArray, ""), "Full") || strings.Contains(strings.Join(reducedArray, ""), "FULL") {
						sailing.Fill = 100
						sailing.CarFill = 100
						sailing.OversizeFill = 100
					} else {
						fill, err := strconv.Atoi(strings.Split(reducedArray[2], "%")[0])
						if err == nil {
							sailing.Fill = 100 - fill
						}
					}
				} else if len(reducedArray) == 5 || len(reducedArray) == 6 {
					if strings.Contains(reducedArray[2], "FULL") || strings.Contains(reducedArray[2], "Full") {
						sailing.Fill = 100
					} else {
						fill, err := strconv.Atoi(strings.Split(reducedArray[2], "%")[0])
						if err == nil {
							sailing.Fill = 100 - fill
						}
					}

					if strings.Contains(reducedArray[3], "FULL") || strings.Contains(reducedArray[3], "Full") {
						sailing.CarFill = 100
					} else {
						fill, err := strconv.Atoi(strings.Split(reducedArray[3], "%")[0])
						if err == nil {
							sailing.CarFill = 100 - fill
						}
					}

					if strings.Contains(reducedArray[4], "FULL") || strings.Contains(reducedArray[4], "Full") {
						sailing.OversizeFill = 100
					} else {
						fill, err := strconv.Atoi(strings.Split(reducedArray[4], "%")[0])
						if err == nil {
							sailing.OversizeFill = 100 - fill
						}
					}
				}
			}

			if strings.Contains(sailing.DepartureTime, "Delayed") {
				sailing.DepartureTime = strings.TrimSpace(strings.Split(sailing.DepartureTime, "Delayed")[0])
			}

			route.Sailings = append(route.Sailings, sailing)
		}
	})

	sailingDuration := strings.ReplaceAll(document.Find("span:contains('Sailing Duration')").Text(), "\u00A0", " ")

	sailingDuration = strings.ReplaceAll(sailingDuration, "Sailing duration:", "")

	if len(strings.TrimSpace(sailingDuration)) == 0 {
		sailingDuration = ""
	} else {
		sailingDuration = strings.TrimSpace(sailingDuration)
	}

	sailingsJson, err := json.Marshal(route.Sailings)
	if err != nil {
		log.Error(err)
	}

	sqlStatement := `
		INSERT INTO capacity_routes (
			route_code,
			from_terminal_code,
			to_terminal_code,
			sailing_duration,
			sailings
		)
		VALUES
			($1, $2, $3, $4, $5) ON CONFLICT (route_code) DO
		UPDATE
		SET
			route_code = EXCLUDED.route_code,
			from_terminal_code = EXCLUDED.from_terminal_code,
			to_terminal_code = EXCLUDED.to_terminal_code,
			sailing_duration = EXCLUDED.sailing_duration,
			sailings = EXCLUDED.sailings
		WHERE
			capacity_routes.route_code = EXCLUDED.route_code`
	_, err = db.Exec(sqlStatement, route.RouteCode, route.FromTerminalCode, route.ToTerminalCode, sailingDuration, sailingsJson)
	if err != nil {
		log.Error(err)
	}
}

/*
 * ScrapeBCNonCapacityRoutes
 *
 * Scrapes BC Ferries non-capacity routes
 *
 * @return void
 */
func ScrapeNonCapacityRoutes() {
	departureTerminals := GetNonCapacityDepartureTerminals()

	destinationTerminals := GetNonCapacityDestinationTerminals()

	for i := 0; i < len(departureTerminals); i++ {
		for j := 0; j < len(destinationTerminals[i]); j++ {
			link := MakeScheduleLink(departureTerminals[i], destinationTerminals[i][j])

			// Make HTTP GET request
			client := &http.Client{}
			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				log.Error(err)
			}

			req.Header.Add("User-Agent", "Mozilla")
			response, err := client.Do(req)
			if err != nil {
				log.Error(err)
			}

			defer response.Body.Close()

			document, err := goquery.NewDocumentFromReader(response.Body)
			if err != nil {
				log.Error(err)
			}

			ScrapeNonCapacityRoute(document, departureTerminals[i], destinationTerminals[i][j])
		}
	}
}

/*
 * ScrapeNonCapacityRoute
 *
 * Scrapes BC Ferries schedule data for a given route
 *
 * @param *goquery.Document document
 * @param string fromTerminalCode
 * @param string toTerminalCode
 *
 * @return void
 */
func ScrapeNonCapacityRoute(document *goquery.Document, fromTerminalCode string, toTerminalCode string) {
	db := GetPostgresInstance()

	route := NonCapacityRoute{
		RouteCode:        fromTerminalCode + toTerminalCode,
		ToTerminalCode:   toTerminalCode,
		FromTerminalCode: fromTerminalCode,
		Sailings:         []NonCapacitySailing{},
	}

	document.Find(".table-seasonal-schedule").First().Find("tbody").First().Find(".schedule-table-row").Each(func(index int, sailingData *goquery.Selection) {
		sailing := NonCapacitySailing{}

		sailingData.Find("td").Each(func(index int, sailingData *goquery.Selection) {
			if index == 1 {
				sailing.DepartureTime = strings.TrimSpace(sailingData.Text())
			} else if index == 2 {
				sailing.ArrivalTime = strings.TrimSpace(sailingData.Text())
			}
		})

		route.Sailings = append(route.Sailings, sailing)
	})

	sailingsJson, err := json.Marshal(route.Sailings)
	if err != nil {
		log.Error(err)
	}

	sailingDuration := ""

	document.Find("table#dailyScheduleTableOnward").Find("tbody").Find("tr").First().Find("td").Each(func(index int, td *goquery.Selection) {
		if index == 3 {
			sailingDuration = strings.TrimSpace(td.Text())
		}
	})

	sqlStatement := `
		INSERT INTO non_capacity_routes (
			route_code, 
			from_terminal_code, 
			to_terminal_code, 
			sailing_duration,
			sailings
		) 
		VALUES 
			($1, $2, $3, $4, $5) ON CONFLICT (route_code) DO 
		UPDATE 
		SET 
			route_code = EXCLUDED.route_code, 
			from_terminal_code = EXCLUDED.from_terminal_code, 
			to_terminal_code = EXCLUDED.to_terminal_code,
			sailing_duration = EXCLUDED.sailing_duration,
			sailings = EXCLUDED.sailings 
		WHERE 
			non_capacity_routes.route_code = EXCLUDED.route_code`
	_, err = db.Exec(sqlStatement, route.RouteCode, route.FromTerminalCode, route.ToTerminalCode, sailingDuration, sailingsJson)
	if err != nil {
		log.Error(err)
	}
}
