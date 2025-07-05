package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"github.com/samuel-pratt/bc-ferries-api/cmd/staticdata"
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
 * ScrapeCapacityRoutes
 *
 * Scrapes capacity routes
 *
 * @return void
 */
func ScrapeCapacityRoutes() {
	departureTerminals := staticdata.GetCapacityDepartureTerminals()
	destinationTerminals := staticdata.GetCapacityDestinationTerminals()

	for i := 0; i < len(departureTerminals); i++ {
		for j := 0; j < len(destinationTerminals[i]); j++ {
			link := MakeCurrentConditionsLink(departureTerminals[i], destinationTerminals[i][j])

			// Make HTTP GET request
			client := &http.Client{}
			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				log.Printf("ScrapeCapacityRoutes: failed to create request for %s: %v", link, err)
				continue
			}

			req.Header.Add("User-Agent", "Mozilla")
			response, err := client.Do(req)
			if err != nil {
				log.Printf("ScrapeCapacityRoutes: failed to fetch %s: %v", link, err)
				continue
			}

			defer response.Body.Close()

			document, err := goquery.NewDocumentFromReader(response.Body)
			if err != nil {
				log.Printf("ScrapeCapacityRoutes: failed to parse response from %s: %v", link, err)
				continue
			}

			ScrapeCapacityRoute(document, departureTerminals[i], destinationTerminals[i][j])
		}
	}
}

/*
 * ScrapeCapacityRoute
 *
 * Scrapes capacity data for a given route
 *
 * @param *goquery.Document document
 * @param string fromTerminalCode
 * @param string toTerminalCode
 *
 * @return void
 */
func ScrapeCapacityRoute(document *goquery.Document, fromTerminalCode string, toTerminalCode string) {
	route := models.CapacityRoute{
		RouteCode:        fromTerminalCode + toTerminalCode,
		ToTerminalCode:   toTerminalCode,
		FromTerminalCode: fromTerminalCode,
		Sailings:         []models.CapacitySailing{},
	}

	document.Find("table.detail-departure-table").Each(func(i int, table *goquery.Selection) {
		table.Find("tbody").Each(func(j int, tbody *goquery.Selection) {
			tbody.Find("tr.mobile-friendly-row").Each(func(k int, row *goquery.Selection) {
				// Init sailing
				sailing := models.CapacitySailing{}

				row.Find("td").Each(func(l int, td *goquery.Selection) {
					if strings.Contains(row.Text(), "Arrived") {
						sailing.SailingStatus = "past"

						if l == 0 {
							timeString := strings.Join(strings.Fields(strings.TrimSpace(td.Find("p").Text())), " ")

							re := regexp.MustCompile(`(?P<DepartureTime>\d{1,2}:\d{2} [ap]m) Departed (?P<ActualDepartureTime>\d{1,2}:\d{2} [ap]m) (?P<VesselName>.+)`)

							// Find the matches
							matches := re.FindStringSubmatch(strings.Join(strings.Fields(timeString), " "))

							if len(matches) == 0 {
								fmt.Println("No matches found, regex error")
							} else {
								// Extracting named groups
								actualDepartureTime := matches[2]
								vesselName := matches[3]

								sailing.DepartureTime = actualDepartureTime
								sailing.VesselName = vesselName
							}
						} else if l == 1 {
							arrivalString := td.Find("div.cc-message-updates").Text()

							re := regexp.MustCompile(`Arrived: (?P<ArrivalTime>\d{1,2}:\d{2} [ap]m)`)

							// Find the matches
							matches := re.FindStringSubmatch(strings.Join(strings.Fields(arrivalString), " "))

							if len(matches) == 0 {
								fmt.Println("No matches found, regex error")
							} else {
								// Extracting named group
								arrivalTime := matches[1]

								sailing.ArrivalTime = arrivalTime
							}
						}
					} else if strings.Contains(row.Text(), "ETA") || strings.Contains(row.Text(), "...") {
						sailing.SailingStatus = "current"

						if l == 0 {
							timeString := strings.Join(strings.Fields(strings.TrimSpace(td.Find("p").Text())), " ")

							re := regexp.MustCompile(`(?P<DepartureTime>\d{1,2}:\d{2} [ap]m) Departed (?P<ActualDepartureTime>\d{1,2}:\d{2} [ap]m) (?P<VesselName>.+)`)

							// Find the matches
							matches := re.FindStringSubmatch(strings.Join(strings.Fields(timeString), " "))

							if len(matches) == 0 {
								fmt.Println("No matches found, regex error")
							} else {
								// Extracting named groups
								actualDepartureTime := matches[2]
								vesselName := matches[3]

								sailing.DepartureTime = actualDepartureTime
								sailing.VesselName = vesselName
							}
						} else if l == 1 {
							etaString := td.Find("div.cc-message-updates").Text()

							re := regexp.MustCompile(`ETA : (?P<ETA>\d{1,2}:\d{2} [ap]m|Variable)`)

							// Find the matches
							matches := re.FindStringSubmatch(strings.Join(strings.Fields(etaString), " "))

							if len(matches) == 0 {
								sailing.ArrivalTime = "..."
							} else {
								// Extracting named group
								etaTime := matches[1]

								sailing.ArrivalTime = etaTime
							}
						}
					} else if strings.Contains(row.Text(), "Details") || strings.Contains(row.Text(), "FULL") || strings.Contains(row.Text(), "Full") || strings.Contains(row.Text(), "%") {
						sailing.SailingStatus = "future"

						if l == 0 {
							// schedule time, vessel
							timeString := strings.Join(strings.Fields(strings.TrimSpace(td.Text())), " ")

							re := regexp.MustCompile(`(?P<Time>\d{1,2}:\d{2} [ap]m)(?: \(Tomorrow\))? (?P<VesselName>.+)`)

							// Find the matches
							matches := re.FindStringSubmatch(strings.Join(strings.Fields(timeString), " "))

							if len(matches) == 0 {
								fmt.Println("No matches found, regex error")
							} else {
								// Extracting named groups
								time := matches[1]
								vesselName := matches[2]

								sailing.DepartureTime = time
								sailing.VesselName = vesselName
							}
						} else if l == 1 {
							// details link
							// if word "Details" is in row, request from link, otherwise take percentage
							fillDetailsString := td.Text()

							if strings.Contains(fillDetailsString, "Details") {
								td.Find("a.vehicle-info-link").Each(func(m int, s *goquery.Selection) {
									href, exists := s.Attr("href")
									link := strings.ReplaceAll("https://www.bcferries.com"+href, " ", "%20")

									if exists {
										client := &http.Client{}
										req, err := http.NewRequest("GET", link, nil)
										if err != nil {
											log.Printf("ScrapeCapacityRoute: failed to create details request for %s: %v", link, err)
											return
										}

										req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
										response, err := client.Do(req)
										if err != nil {
											log.Printf("ScrapeCapacityRoute: failed to fetch details from %s: %v", link, err)
											return
										}

										defer response.Body.Close()

										fillDocument, err := goquery.NewDocumentFromReader(response.Body)
										if err != nil {
											log.Printf("ScrapeCapacityRoute: failed to parse fill details from %s: %v", link, err)
											return
										}

										// fmt.Println(fillDocument.Text())
										fillDocument.Find("p.vehicle-icon-text").Each(func(o int, percentageText *goquery.Selection) {
											if o == 0 {
												fillPercentage := strings.TrimSpace(percentageText.Text())

												if strings.Contains(fillPercentage, "Full") || strings.Contains(fillPercentage, "full") || strings.Contains(fillPercentage, "FULL") {
													sailing.Fill = 100
													sailing.CarFill = 100
													sailing.OversizeFill = 100
												} else {
													fillPercentageInt, err := strconv.Atoi(strings.ReplaceAll(fillPercentage, "%", ""))
													if err != nil {
														// ... handle error
													}

													sailing.Fill = 100 - fillPercentageInt
												}
											} else if o == 1 {
												fillPercentage := strings.TrimSpace(percentageText.Text())

												if strings.Contains(fillPercentage, "Full") || strings.Contains(fillPercentage, "full") || strings.Contains(fillPercentage, "FULL") {
													sailing.CarFill = 100
												} else {
													fillPercentageInt, err := strconv.Atoi(strings.ReplaceAll(fillPercentage, "%", ""))
													if err != nil {
														// ... handle error
													}

													sailing.CarFill = 100 - fillPercentageInt
												}
											} else if o == 2 {
												fillPercentage := strings.TrimSpace(percentageText.Text())

												if strings.Contains(fillPercentage, "Full") || strings.Contains(fillPercentage, "full") || strings.Contains(fillPercentage, "FULL") {
													sailing.OversizeFill = 100
												} else {
													fillPercentageInt, err := strconv.Atoi(strings.ReplaceAll(fillPercentage, "%", ""))
													if err != nil {
														// ... handle error
													}

													sailing.OversizeFill = 100 - fillPercentageInt
												}
											}
										})

									}
								})
							} else {
								if strings.Contains(fillDetailsString, "Full") {
									sailing.Fill = 100
									sailing.CarFill = 100
									sailing.OversizeFill = 100
								} else {
									fillPercentage := strings.TrimSpace(td.Find("span.cc-vessel-percent-full").Text())

									fillPercentageInt, err := strconv.Atoi(strings.ReplaceAll(fillPercentage, "%", ""))
									if err != nil {
										// ... handle error
									}

									sailing.Fill = 100 - fillPercentageInt
								}
							}
						}
					}
				})

				// Add salining to route
				route.Sailings = append(route.Sailings, sailing)
			})
		})
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
		log.Printf("ScrapeCapacityRoute: failed to marshal sailings for route %s: %v", route.RouteCode, err)
		return
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
	_, err = db.Conn.Exec(sqlStatement, route.RouteCode, route.FromTerminalCode, route.ToTerminalCode, sailingDuration, sailingsJson)
	if err != nil {
		log.Printf("ScrapeCapacityRoute: failed to insert route %s: %v", route.RouteCode, err)
		return
	}
}

/*
 * ScrapeNonCapacityRoutes
 *
 * Scrapes non-capacity routes
 *
 * @return void
 */
func ScrapeNonCapacityRoutes() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	departureTerminals := staticdata.GetNonCapacityDepartureTerminals()
	destinationTerminals := staticdata.GetNonCapacityDestinationTerminals()

	for i := 0; i < len(departureTerminals); i++ {
		for j := 0; j < len(destinationTerminals[i]); j++ {
			link := MakeScheduleLink(departureTerminals[i], destinationTerminals[i][j])

			html, err := fetchWithChromedp(ctx, link)
			if err != nil {
				fmt.Printf("ScrapeBCNonCapacityRoutes: chromedp fetch failed for %s: %v\n", link, err)
				continue
			}

			document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Printf("ScrapeBCNonCapacityRoutes: failed to parse HTML for %s: %v\n", link, err)
				continue
			}

			ScrapeNonCapacityRoute(document, departureTerminals[i], destinationTerminals[i][j])
		}
	}
}

/*
 * ScrapeNonCapacityRoute
 *
 * Scrapes schedule data for a given route
 *
 * @param *goquery.Document document
 * @param string fromTerminalCode
 * @param string toTerminalCode
 *
 * @return void
 */
func ScrapeNonCapacityRoute(document *goquery.Document, fromTerminalCode string, toTerminalCode string) {
	route := models.NonCapacityRoute{
		RouteCode:        fromTerminalCode + toTerminalCode,
		ToTerminalCode:   toTerminalCode,
		FromTerminalCode: fromTerminalCode,
		Sailings:         []models.NonCapacitySailing{},
	}

	document.Find(".table-seasonal-schedule").First().Find("tbody").First().Find(".schedule-table-row").Each(func(index int, sailingData *goquery.Selection) {
		sailing := models.NonCapacitySailing{}

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
		log.Printf("ScrapeNonCapacityRoute: failed to marshal sailings for route %s: %v", route.RouteCode, err)
		return
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
	_, err = db.Conn.Exec(sqlStatement, route.RouteCode, route.FromTerminalCode, route.ToTerminalCode, sailingDuration, sailingsJson)
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: failed to insert route %s: %v", route.RouteCode, err)
		return
	}
}

/********************/
/* Helper Functions */
/********************/

/*
 * fetchWithChromedp
 *
 * Uses a headless Chrome browser to fetch and render the full HTML content of a given URL.
 * This is used to bypass JavaScript-based protections like Queue-it by executing the page
 * in a real browser environment.
 *
 * @param string url - The URL to navigate to
 *
 * @return string - The full outer HTML of the rendered page
 * @return error - Any error encountered during navigation or retrieval
 */
func fetchWithChromedp(ctx context.Context, url string) (string, error) {
	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)

	return html, err
}
