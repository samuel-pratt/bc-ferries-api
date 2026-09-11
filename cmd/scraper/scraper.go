package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
			client := &http.Client{Timeout: 20 * time.Second}
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

			if response.StatusCode != http.StatusOK {
				response.Body.Close()
				continue
			}

			document, err := goquery.NewDocumentFromReader(response.Body)
			response.Body.Close()
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
	if document.Find("table.detail-departure-table").Length() == 0 {
		log.Printf("capacity page invalid for %s%s; retaining previous data", fromTerminalCode, toTerminalCode)
		return
	}
	observedAt := time.Now().UTC()
	route := models.CapacityRoute{
		RouteCode:        fromTerminalCode + toTerminalCode,
		ToTerminalCode:   toTerminalCode,
		FromTerminalCode: fromTerminalCode,
		Sailings:         []models.CapacitySailing{},
		ObservedAt:       observedAt,
	}
	route.Notices = parseServiceNotices(document, route.RouteCode, fromTerminalCode, observedAt)
	route.Terminal = parseTerminalInfo(document, fromTerminalCode, observedAt)

	document.Find("table.detail-departure-table").Each(func(i int, table *goquery.Selection) {
		table.Find("tbody").Each(func(j int, tbody *goquery.Selection) {
			tbody.Find("tr.mobile-friendly-row").Each(func(k int, row *goquery.Selection) {
				// Init sailing
				sailing := models.CapacitySailing{ObservedAt: observedAt}

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
				normalizeCapacityStatus(&sailing)
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

	if len(route.Sailings) == 0 {
		log.Printf("no capacity sailings for %s; retaining previous data", route.RouteCode)
		return
	}
	sailingsJson, err := json.Marshal(route.Sailings)
	if err != nil {
		log.Printf("ScrapeCapacityRoute: failed to marshal sailings for route %s: %v", route.RouteCode, err)
		return
	}
	operationalJSON, err := json.Marshal(models.RouteOperationalData{Notices: route.Notices, Terminal: route.Terminal, ObservedAt: observedAt})
	if err != nil {
		log.Printf("ScrapeCapacityRoute: failed to marshal operational data for route %s: %v", route.RouteCode, err)
		return
	}

	err = db.SaveRoute("capacity", route.RouteCode, route.FromTerminalCode, route.ToTerminalCode, sailingDuration, sailingsJson, operationalJSON)
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
	ctx, cancel := newBrowserContext()
	defer cancel()
	if err := chromedp.Run(ctx); err != nil {
		log.Printf("Chrome startup failed: %v", err)
		return
	}

	departureTerminals := staticdata.GetNonCapacityDepartureTerminals()
	destinationTerminals := staticdata.GetNonCapacityDestinationTerminals()

	for i := 0; i < len(departureTerminals); i++ {
		for j := 0; j < len(destinationTerminals[i]); j++ {
			departure := departureTerminals[i]
			destination := destinationTerminals[i][j]

			dailyLink := MakeScheduleLink(departure, destination)
			html, finalURL, err := fetchWithChromedp(ctx, dailyLink)
			if err == nil {
				document, parseErr := goquery.NewDocumentFromReader(strings.NewReader(html))
				if parseErr == nil {
					isDaily := isDailySchedulePage(document, finalURL)
					if ScrapeNonCapacityRoute(document, departure, destination, isDaily) {
						continue
					}
				} else {
					log.Printf("ScrapeNonCapacityRoutes: failed to parse daily HTML for %s: %v", dailyLink, parseErr)
				}
			} else {
				log.Printf("ScrapeNonCapacityRoutes: daily fetch failed for %s: %v", dailyLink, err)
			}

			seasonalLink := MakeSeasonalScheduleLink(departure, destination)
			html, _, err = fetchWithChromedp(ctx, seasonalLink)
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
	observedAt := time.Now().UTC()
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
		ObservedAt:       observedAt,
	}
	route.Notices = parseServiceNotices(document, route.RouteCode, fromTerminalCode, observedAt)
	route.Terminal = parseTerminalInfo(document, fromTerminalCode, observedAt)
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
			// BC Ferries uses both phrases for date-specific exclusions.
			if strings.Contains(combinedRed, "except on") || strings.Contains(combinedRed, "not available on") {
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

	// Defensive final check: malformed or repeated markup must not leak duplicate
	// sailings into the API response.
	uniqueSailings := make([]models.NonCapacitySailing, 0, len(route.Sailings))
	seenSailings := make(map[string]struct{}, len(route.Sailings))
	for _, sailing := range route.Sailings {
		if sailing.DepartureTime == "" || sailing.ArrivalTime == "" {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(sailing.DepartureTime)) + "|" +
			strings.ToLower(strings.TrimSpace(sailing.ArrivalTime))
		if _, seen := seenSailings[key]; seen {
			continue
		}
		seenSailings[key] = struct{}{}
		sailing.Status = "scheduled"
		sailing.ObservedAt = observedAt
		uniqueSailings = append(uniqueSailings, sailing)
	}
	route.Sailings = uniqueSailings
	if len(route.Sailings) == 0 {
		log.Printf("ScrapeNonCapacityRoute: no valid sailings parsed for %s", route.RouteCode)
		return false
	}

	// ---- Step 5: save
	sailingsJSON, err := json.Marshal(route.Sailings)
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: marshal error for %s: %v", route.RouteCode, err)
		return false
	}
	operationalJSON, err := json.Marshal(models.RouteOperationalData{Notices: route.Notices, Terminal: route.Terminal, ObservedAt: observedAt})
	if err != nil {
		log.Printf("ScrapeNonCapacityRoute: operational metadata marshal error for %s: %v", route.RouteCode, err)
		return false
	}

	err = db.SaveRoute("noncapacity", route.RouteCode, route.FromTerminalCode, route.ToTerminalCode, sailingDuration, sailingsJSON, operationalJSON)
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

	table := document.Find("#dailyScheduleTableOnward").First()
	if table.Length() == 0 {
		return sailings, sailingDuration, false
	}

	func() {

		tableSailings := make([]models.NonCapacitySailing, 0)
		tableDuration := ""

		rows := table.Find("tbody").First().ChildrenFiltered("tr.schedule-table-row")

		rows.Each(func(_ int, row *goquery.Selection) {
			if row.Find("th").Length() > 0 {
				return
			}

			tds := row.ChildrenFiltered("td")
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
			if len(timeTokens) < 2 {
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

		sailings = tableSailings
		sailingDuration = tableDuration
	}()

	return sailings, sailingDuration, len(sailings) > 0
}

func normalizeCapacityStatus(sailing *models.CapacitySailing) {
	sailing.StatusText = strings.TrimSpace(sailing.VesselStatus)
	switch sailing.SailingStatus {
	case "cancelled":
		sailing.Status = "cancelled"
	case "past":
		sailing.Status = "arrived"
	case "current":
		sailing.Status = "departed"
	default:
		sailing.Status = "scheduled"
	}

	combined := strings.TrimSpace(sailing.VesselName + " " + sailing.VesselStatus)
	delayRe := regexp.MustCompile(`(?i)delay(?:ed)?(?:\s+approx(?:imately)?\.?)?\s*(\d+)\s*(?:m|min|minutes?)`)
	if match := delayRe.FindStringSubmatch(combined); len(match) == 2 && sailing.Status == "scheduled" {
		if minutes, err := strconv.Atoi(match[1]); err == nil {
			sailing.DelayMinutes = &minutes
			sailing.Status = "delayed"
			if sailing.StatusText == "" {
				sailing.StatusText = match[0]
			}
		}
	}
}

func parseServiceNotices(document *goquery.Document, routeCode, terminalCode string, observedAt time.Time) []models.ServiceNotice {
	seen := make(map[string]struct{})
	notices := make([]models.ServiceNotice, 0)
	document.Find(".cc-service-notice a[href]").Each(func(_ int, link *goquery.Selection) {
		href, ok := link.Attr("href")
		if !ok || strings.TrimSpace(href) == "" {
			return
		}
		absoluteURL := href
		if strings.HasPrefix(href, "/") {
			absoluteURL = "https://www.bcferries.com" + href
		}
		id := ""
		if match := regexp.MustCompile(`serviceNoticeCode=([0-9]+)`).FindStringSubmatch(href); len(match) == 2 {
			id = match[1]
		}
		if id == "" {
			id = strings.ToLower(strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return -1
			}, strings.TrimSpace(link.Text())))
		}
		if _, exists := seen[id]; exists || id == "" {
			return
		}
		seen[id] = struct{}{}
		title := strings.Join(strings.Fields(link.Text()), " ")
		noticeType, severity := classifyNotice(title)
		notices = append(notices, models.ServiceNotice{
			ID: id, Title: title, URL: absoluteURL, Type: noticeType, Severity: severity,
			TerminalCodes: []string{terminalCode}, RouteCodes: []string{routeCode}, ObservedAt: observedAt,
		})
	})
	return notices
}

func classifyNotice(title string) (string, string) {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "cancel") || strings.Contains(lower, "closure"):
		return "cancellation", "critical"
	case strings.Contains(lower, "delay") || strings.Contains(lower, "revised schedule"):
		return "delay", "warning"
	case strings.Contains(lower, "terminal") || strings.Contains(lower, "parking"):
		return "terminal", "information"
	case strings.Contains(lower, "amenit") || strings.Contains(lower, "capacity"):
		return "service-change", "warning"
	default:
		return "information", "information"
	}
}

func parseTerminalInfo(document *goquery.Document, terminalCode string, observedAt time.Time) *models.TerminalInfo {
	info := &models.TerminalInfo{Code: terminalCode, ObservedAt: observedAt, Webcams: []models.Webcam{}}
	terminal := document.Find("#terminalDetails").First()
	info.Name = strings.Join(strings.Fields(terminal.Find("h4").First().Text()), " ")
	if href, ok := terminal.Find("a.terminal-detail-btn").First().Attr("href"); ok {
		if strings.HasPrefix(href, "/") {
			href = "https://www.bcferries.com" + href
		}
		info.DetailsURL = href
	}
	terminal.Find("a[href*='maps']").EachWithBreak(func(_ int, link *goquery.Selection) bool {
		info.DirectionsURL, _ = link.Attr("href")
		return false
	})
	document.Find(".webcam-wrapper img[src]").Each(func(index int, image *goquery.Selection) {
		src, ok := image.Attr("src")
		if !ok {
			return
		}
		container := image.Parent()
		label := strings.Join(strings.Fields(container.Find("p").First().Text()), " ")
		info.Webcams = append(info.Webcams, models.Webcam{
			ID:    fmt.Sprintf("%s-%d", strings.ToLower(terminalCode), index+1),
			Label: label, ImageURL: src, ObservedAt: observedAt,
		})
	})
	if info.Name == "" && len(info.Webcams) == 0 && info.DetailsURL == "" {
		return nil
	}
	return info
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
func fetchWithChromedp(ctx context.Context, url string) (string, string, error) {
	tab, closeTab := chromedp.NewContext(ctx)
	defer closeTab()
	ctx, cancel := context.WithTimeout(tab, 45*time.Second)
	defer cancel()
	var html string
	var finalURL string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Location(&finalURL),
		chromedp.OuterHTML("html", &html),
	)

	return html, finalURL, err
}

func newBrowserContext() (context.Context, context.CancelFunc) {
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	// Docker runs Chrome as an unprivileged user, with all capabilities dropped.
	// Its nested namespace sandbox is unavailable on many Docker hosts.
	if os.Getenv("CHROME_NO_SANDBOX") == "true" {
		opts = append(opts, chromedp.NoSandbox)
	}
	alloc, closeAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, closeBrowser := chromedp.NewContext(alloc)
	return ctx, func() { closeBrowser(); closeAlloc() }
}

// isDailySchedulePage detects redirects from a daily URL to a seasonal page.
// The final browser URL is authoritative; canonical is a fallback for fixtures
// and pages whose client-side navigation leaves an ambiguous location.
func isDailySchedulePage(document *goquery.Document, finalURL string) bool {
	finalURL = strings.ToLower(finalURL)
	if strings.Contains(finalURL, "/schedules/seasonal/") {
		return false
	}
	if strings.Contains(finalURL, "/schedules/daily/") {
		return true
	}

	canonical, _ := document.Find(`link[rel="canonical"]`).First().Attr("href")
	canonical = strings.ToLower(canonical)
	return strings.Contains(canonical, "/schedules/daily/")
}
