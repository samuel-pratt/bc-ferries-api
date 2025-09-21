package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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
 * Builds a link to the SEASONAL schedule page for a given departure and destination.
 * Seasonal pages contain the weekly tables used by the non-capacity scraper.
 *
 * @param string departure
 * @param string destination
 *
 * @return string
 */
func MakeScheduleLink(departure, destination string) string {
    return "https://www.bcferries.com/routes-fares/schedules/seasonal/" + departure + "-" + destination
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
                        rowTextLower := strings.ToLower(row.Text())

                        // Handle explicitly cancelled rows
                        if strings.Contains(rowTextLower, "cancelled") {
                            sailing.SailingStatus = "cancelled"

                            if l == 0 {
                                // Scheduled time and vessel
                                timeString := strings.Join(strings.Fields(strings.TrimSpace(td.Text())), " ")
                                re := regexp.MustCompile(`(?P<Time>\d{1,2}:\d{2} [ap]m)(?: \(Tomorrow\))? (?P<VesselName>.+)`)
                                matches := re.FindStringSubmatch(strings.Join(strings.Fields(timeString), " "))
                                if len(matches) >= 3 {
                                    sailing.DepartureTime = matches[1]
                                    sailing.VesselName = matches[2]
                                }
                            } else if l == 1 {
                                // Capture reason if present under the red text block
                                // Prefer the second <p> which often holds the reason
                                reason := strings.TrimSpace(td.Find("div.text-red p").Eq(1).Text())
                                if reason == "" {
                                    // Fallback to the whole red block text
                                    reason = strings.TrimSpace(td.Find("div.text-red").Text())
                                }
                                if reason != "" {
                                    sailing.VesselStatus = reason
                                }
                            }
                        } else if strings.Contains(row.Text(), "Arrived") {
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
                        } else if strings.Contains(row.Text(), "Details") || strings.Contains(row.Text(), "%") || strings.Contains(strings.ToLower(row.Text()), "full") {
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

                                                if strings.Contains(strings.ToLower(fillPercentage), "full") {
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

                                                if strings.Contains(strings.ToLower(fillPercentage), "full") {
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

                                                if strings.Contains(strings.ToLower(fillPercentage), "full") {
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
                                if strings.Contains(strings.ToLower(fillDetailsString), "full") {
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

    // Try to find sailing duration text in a case-insensitive way
    sailingDuration := ""
    document.Find("span").Each(func(_ int, s *goquery.Selection) {
        if sailingDuration != "" {
            return
        }
        txt := strings.ReplaceAll(s.Text(), "\u00A0", " ")
        if strings.Contains(strings.ToLower(txt), "sailing duration:") {
            sailingDuration = txt
        }
    })
    sailingDuration = strings.ReplaceAll(sailingDuration, "Sailing duration:", "")
    sailingDuration = strings.ReplaceAll(sailingDuration, "sailing duration:", "")
    sailingDuration = strings.TrimSpace(sailingDuration)

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
func ScrapeNonCapacityRoute(document *goquery.Document, fromTerminalCode, toTerminalCode string) {
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: failed to load PT location: %v", err)
		return
	}

	normalizeDay := func(s string) string {
		s = strings.TrimSpace(strings.ToUpper(s))
		// Treat trailing "S" as optional: MONDAY == MONDAYS
		if strings.HasSuffix(s, "S") {
			s = strings.TrimSuffix(s, "S")
		}
		return s
	}
	todayNorm := normalizeDay(time.Now().In(loc).Weekday().String()) // e.g. "MONDAY"

	route := models.NonCapacityRoute{
		RouteCode:        fromTerminalCode + toTerminalCode,
		FromTerminalCode: fromTerminalCode,
		ToTerminalCode:   toTerminalCode,
		Sailings:         []models.NonCapacitySailing{},
	}

    // ---- Step 1: find the seasonal schedule table that contains weekday theads
    var scheduleTable *goquery.Selection
    document.Find("table.table-seasonal-schedule").Each(func(_ int, t *goquery.Selection) {
        if scheduleTable != nil {
            return
        }
        // Heuristic: a real schedule table has thead rows with day labels
        if t.Find("thead tr[data-schedule-day], thead [data-schedule-day], thead h4, thead b").Length() > 0 {
            scheduleTable = t
        }
    })
    // Fallback to the historical assumption (2nd table) if heuristic fails
    if scheduleTable == nil {
        scheduleTable = document.Find("table.table-seasonal-schedule").Eq(1)
    }
    if scheduleTable == nil || scheduleTable.Length() == 0 {
        log.Printf("ScrapeNonCapacityRoute: seasonal schedule table not found")
        return
    }

	// ---- Step 2: find the <thead> whose day matches today (MONDAY vs MONDAYS, any case)
    var dayBody *goquery.Selection
    scheduleTable.Find("thead").Each(func(_ int, thead *goquery.Selection) {
		if dayBody != nil {
			return
		}

		// Prefer the attribute if present.
		dayAttr := thead.Find("tr").First().AttrOr("data-schedule-day", "")
		dayAttrNorm := normalizeDay(dayAttr)

		match := (dayAttrNorm != "" && dayAttrNorm == todayNorm)
		if !match {
			// Fallback: try visible text inside thead (e.g., MONDAY Depart)
			txt := thead.Find("h4, b, th").First().Text()
			txtNorm := normalizeDay(txt)
			// If the text contains the weekday token (e.g., "MONDAY DEPART"), accept it.
			match = (txtNorm == todayNorm) || strings.Contains(txtNorm, todayNorm)
		}

		if match {
			// ---- Step 3: go to the NEXT sibling under the table; skip to the first <tbody>
			tb := thead.Next()
			for tb.Length() > 0 && goquery.NodeName(tb) != "tbody" {
				tb = tb.Next()
			}
			if tb.Length() > 0 && goquery.NodeName(tb) == "tbody" {
				dayBody = tb
			}
		}
	})

	if dayBody == nil {
		log.Printf("ScrapeNonCapacityRoute: no tbody found for today (%s) in second table", todayNorm)
		return
	}

	clean := func(s string) string {
		s = strings.ReplaceAll(s, "\u00a0", " ") // NBSP -> space
		return strings.TrimSpace(s)
	}

    // ---- Step 4: parse rows in the found <tbody>
    dayBody.Find("tr.schedule-table-row").Each(func(_ int, row *goquery.Selection) {
        tds := row.Find("td")
        if tds.Length() < 3 {
            return
        }

        // Extract clean departure time (first time token) and any status notes
        depCell := tds.Eq(1)
        depRaw := clean(depCell.Text())

        // Capture red/black status notes if present (e.g., Only on..., Except on..., Foot passengers only, Dangerous goods only)
        var statuses []string
        depCell.Find("p").Each(func(_ int, p *goquery.Selection) {
            txt := clean(p.Text())
            if txt == "" {
                return
            }
            // Only keep informative notes, skip if it's just whitespace
            // Common classes include red-text italic-style or text-black
            if p.HasClass("red-text") || p.HasClass("text-black") {
                statuses = append(statuses, txt)
            }
        })

        // Extract the first time-like token from the departure cell
        depTime := depRaw
        if re := regexp.MustCompile(`(?i)\b\d{1,2}:\d{2}\s*[ap]m\b`); re != nil {
            if m := re.FindString(depRaw); m != "" {
                depTime = m
            }
        }

        // Extract clean arrival time (first time token)
        arrCell := tds.Eq(2)
        arrRaw := clean(arrCell.Text())
        arrTime := arrRaw
        if re := regexp.MustCompile(`(?i)\b\d{1,2}:\d{2}\s*[ap]m\b`); re != nil {
            if m := re.FindString(arrRaw); m != "" {
                arrTime = m
            }
        }

        s := models.NonCapacitySailing{
            DepartureTime: depTime,
            ArrivalTime:   arrTime,
        }
        if len(statuses) > 0 {
            s.VesselStatus = strings.Join(statuses, " | ")
        }

        if s.DepartureTime != "" || s.ArrivalTime != "" {
            route.Sailings = append(route.Sailings, s)
        }
    })

	// Optional: route-level duration (from the first row's 4th cell, if present)
	sailingDuration := ""
	if firstRow := dayBody.Find("tr.schedule-table-row").First(); firstRow.Length() > 0 {
		if cell := firstRow.Find("td").Eq(3); cell.Length() > 0 {
			sailingDuration = clean(cell.Text())
		}
	}

	// ---- Step 5: save
	sailingsJSON, err := json.Marshal(route.Sailings)
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: marshal error for %s: %v", route.RouteCode, err)
		return
	}

	sqlStatement := `
		INSERT INTO non_capacity_routes (
			route_code, 
			from_terminal_code, 
			to_terminal_code, 
			sailing_duration,
			sailings
		) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (route_code) DO UPDATE SET
			from_terminal_code = EXCLUDED.from_terminal_code, 
			to_terminal_code = EXCLUDED.to_terminal_code,
			sailing_duration = EXCLUDED.sailing_duration,
			sailings = EXCLUDED.sailings 
	`
	_, err = db.Conn.Exec(sqlStatement,
		route.RouteCode, route.FromTerminalCode, route.ToTerminalCode, sailingDuration, sailingsJSON,
	)
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: DB insert/update failed for %s: %v", route.RouteCode, err)
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
