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
 * Builds a link to the DAILY schedule page for a given departure and destination.
 * Daily pages reflect route-specific service updates in effect for the current day.
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
 * MakeSeasonalScheduleLink
 *
 * Builds a link to the SEASONAL schedule page for a given departure and destination.
 *
 * @param string departure
 * @param string destination
 *
 * @return string
 */
func MakeSeasonalScheduleLink(departure, destination string) string {
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
			departure := departureTerminals[i]
			destination := destinationTerminals[i][j]

			dailyLink := MakeScheduleLink(departure, destination)
			html, err := fetchWithChromedp(ctx, dailyLink)
			if err == nil {
				document, parseErr := goquery.NewDocumentFromReader(strings.NewReader(html))
				if parseErr == nil {
					if ScrapeNonCapacityRoute(document, departure, destination, true) {
						continue
					}
				} else {
					log.Printf("ScrapeNonCapacityRoutes: failed to parse daily HTML for %s: %v", dailyLink, parseErr)
				}
			} else {
				log.Printf("ScrapeNonCapacityRoutes: daily fetch failed for %s: %v", dailyLink, err)
			}

			seasonalLink := MakeSeasonalScheduleLink(departure, destination)
			html, err = fetchWithChromedp(ctx, seasonalLink)
			if err != nil {
				log.Printf("ScrapeNonCapacityRoutes: seasonal fetch failed for %s: %v", seasonalLink, err)
				continue
			}

			document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				log.Printf("ScrapeNonCapacityRoutes: failed to parse seasonal HTML for %s: %v", seasonalLink, err)
				continue
			}

			ScrapeNonCapacityRoute(document, departure, destination, false)
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
 * @param bool isDaily
 *
 * @return bool - True when route data was parsed and persisted
 */
func ScrapeNonCapacityRoute(document *goquery.Document, fromTerminalCode, toTerminalCode string, isDaily bool) bool {
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: failed to load PT location: %v", err)
		return false
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
	sailingDuration := ""
	if isDaily {
		if dailySailings, dailyDuration, ok := parseDailyScheduleSailings(document); ok {
			route.Sailings = dailySailings
			sailingDuration = dailyDuration
		}
	}

	if len(route.Sailings) == 0 {
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
			return false
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
			return false
		}

		clean := func(s string) string {
			s = strings.ReplaceAll(s, "\u00a0", " ") // NBSP -> space
			return strings.TrimSpace(s)
		}

		// Parse a list of month/day mentions from a status string like
		// "Only on Sep 14, 28 & Oct 12" or "Except on Oct 13".
		// Returns a set keyed by "MM-DD" for quick lookup.
		parseMentionedDates := func(note string, year int) map[string]struct{} {
			res := make(map[string]struct{})
			if note == "" {
				return res
			}
			lower := strings.ToLower(note)

			monthMap := map[string]time.Month{
				"jan": time.January, "january": time.January,
				"feb": time.February, "february": time.February,
				"mar": time.March, "march": time.March,
				"apr": time.April, "april": time.April,
				"may": time.May,
				"jun": time.June, "june": time.June,
				"jul": time.July, "july": time.July,
				"aug": time.August, "august": time.August,
				"sep": time.September, "sept": time.September, "september": time.September,
				"oct": time.October, "october": time.October,
				"nov": time.November, "november": time.November,
				"dec": time.December, "december": time.December,
			}

			// 1) Find explicit Month Day pairs
			mdRe := regexp.MustCompile(`(?i)(jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:t(?:ember)?)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\s+(\d{1,2})`)
			matches := mdRe.FindAllStringSubmatch(lower, -1)

			for _, m := range matches {
				monKey := m[1]
				dayStr := m[2]
				if mon, ok := monthMap[monKey]; ok {
					if d, err := strconv.Atoi(dayStr); err == nil {
						key := fmt.Sprintf("%02d-%02d", int(mon), d)
						res[key] = struct{}{}
					}
				}
			}

			// 2) Handle shorthand days following a month (e.g., "Sep 14, 28 & Oct 12")
			//    For each segment that starts with a month, capture trailing , <day> pieces until next month appears
			segRe := regexp.MustCompile(`(?i)(jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:t(?:ember)?)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\s+\d{1,2}([^a-z]*)`)
			pos := 0
			for {
				loc := segRe.FindStringSubmatchIndex(lower[pos:])
				if loc == nil {
					break
				}
				// Extract month for this segment
				seg := lower[pos+loc[0] : pos+loc[1]]
				mon := mdRe.FindStringSubmatch(seg)
				if len(mon) >= 3 {
					monKey := mon[1]
					if monVal, ok := monthMap[monKey]; ok {
						// After the first "Month DD", scan the tail for , DD patterns
						tail := seg[len(mon[0]):]
						// Match bare days like ", 28" without unsupported lookaheads
						ddRe := regexp.MustCompile(`(?i)[,&\s]+(\d{1,2})\b`)
						ddMatches := ddRe.FindAllStringSubmatch(tail, -1)
						for _, dm := range ddMatches {
							if d, err := strconv.Atoi(dm[1]); err == nil {
								key := fmt.Sprintf("%02d-%02d", int(monVal), d)
								res[key] = struct{}{}
							}
						}
					}
				}
				pos += loc[1]
			}

			return res
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
			var redNotes []string
			depCell.Find("p").Each(func(_ int, p *goquery.Selection) {
				txt := clean(p.Text())
				if txt == "" {
					return
				}
				// Only keep informative notes, skip if it's just whitespace
				// Common classes include red-text italic-style or text-black
				if p.HasClass("red-text") || p.HasClass("text-black") {
					statuses = append(statuses, txt)
					if p.HasClass("red-text") {
						redNotes = append(redNotes, txt)
					}
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

			// Filter: drop dangerous goods only sailings outright
			depLower := strings.ToLower(depCell.Text())
			if strings.Contains(depLower, "dangerous goods only") || strings.Contains(depLower, "no passengers permitted") {
				return
			}

			// Apply exception rules: "Only on <dates>" and "Except on <dates>"
			// Build a combined note string from red notes
			combinedRed := strings.ToLower(strings.Join(redNotes, "; "))
			today := time.Now().In(loc)
			todayKey := fmt.Sprintf("%02d-%02d", int(today.Month()), today.Day())

			// If there is an "only on" note, include only if today is listed
			if strings.Contains(combinedRed, "only on") {
				dates := parseMentionedDates(combinedRed, today.Year())
				if _, ok := dates[todayKey]; !ok {
					return
				}
			}
			// If there is an "except on" note, exclude if today is listed
			if strings.Contains(combinedRed, "except on") {
				dates := parseMentionedDates(combinedRed, today.Year())
				if _, ok := dates[todayKey]; ok {
					return
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
		if firstRow := dayBody.Find("tr.schedule-table-row").First(); firstRow.Length() > 0 {
			if cell := firstRow.Find("td").Eq(3); cell.Length() > 0 {
				sailingDuration = clean(cell.Text())
			}
		}
	}

	if len(route.Sailings) == 0 {
		log.Printf("ScrapeNonCapacityRoute: no sailings parsed for %s", route.RouteCode)
		return false
	}

	// ---- Step 5: save
	sailingsJSON, err := json.Marshal(route.Sailings)
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: marshal error for %s: %v", route.RouteCode, err)
		return false
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
		return false
	}

	return true
}

func parseDailyScheduleSailings(document *goquery.Document) ([]models.NonCapacitySailing, string, bool) {
	clean := func(s string) string {
		s = strings.ReplaceAll(s, "\u00a0", " ") // NBSP -> space
		return strings.TrimSpace(s)
	}
	timeRe := regexp.MustCompile(`(?i)\b\d{1,2}:\d{2}\s*[ap]m\b`)
	durationRe := regexp.MustCompile(`\b\d{1,2}:\d{2}\b`)
	extractTime := func(s string) string {
		return timeRe.FindString(clean(s))
	}

	var sailings []models.NonCapacitySailing
	sailingDuration := ""

	document.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(sailings) > 0 {
			return
		}

		headerText := strings.ToUpper(clean(table.Find("thead").First().Text()))
		if headerText == "" {
			headerText = strings.ToUpper(clean(table.Find("tr").First().Text()))
		}
		if !strings.Contains(headerText, "DEPART") || !strings.Contains(headerText, "ARRIVE") {
			return
		}

		tableSailings := make([]models.NonCapacitySailing, 0)
		tableDuration := ""

		rows := table.Find("tbody tr")
		if rows.Length() == 0 {
			rows = table.Find("tr")
		}

		rows.Each(func(_ int, row *goquery.Selection) {
			if row.Find("th").Length() > 0 {
				return
			}

			tds := row.Find("td")
			if tds.Length() < 2 {
				return
			}

			rowTextLower := strings.ToLower(clean(row.Text()))
			if strings.Contains(rowTextLower, "dangerous goods only") || strings.Contains(rowTextLower, "no passengers permitted") {
				return
			}

			var timeTokens []string
			tds.Each(func(_ int, td *goquery.Selection) {
				if m := extractTime(td.Text()); m != "" {
					timeTokens = append(timeTokens, m)
				}
			})
			if len(timeTokens) == 0 {
				return
			}
			departureTime := timeTokens[0]
			arrivalTime := ""
			if len(timeTokens) > 1 {
				arrivalTime = timeTokens[1]
			}

			if tableDuration == "" {
				tds.Each(func(_ int, td *goquery.Selection) {
					if tableDuration != "" {
						return
					}
					tdText := clean(td.Text())
					if tdText == "" {
						return
					}
					tdTextLower := strings.ToLower(tdText)
					if strings.Contains(tdTextLower, "am") || strings.Contains(tdTextLower, "pm") {
						return
					}
					if m := durationRe.FindString(tdText); m != "" {
						tableDuration = m
					}
				})
			}

			tableSailings = append(tableSailings, models.NonCapacitySailing{
				DepartureTime: departureTime,
				ArrivalTime:   arrivalTime,
			})
		})

		if len(tableSailings) == 0 {
			return
		}

		sailings = tableSailings
		sailingDuration = tableDuration
	})

	return sailings, sailingDuration, len(sailings) > 0
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
