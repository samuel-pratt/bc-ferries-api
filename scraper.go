package main

// Import OS and fmt packages
import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Response struct {
	Schedule  map[string]map[string][]map[string]string `json:"schedule"`
	ScrapedAt time.Time                                 `json:"scrapedAt"`
}

func ScrapeCapacityRoutes() Response {
	// Links to individual schedule pages
	capacityRouteLinks := [12]string{
		"https://www.bcferries.com/current-conditions/TSA-SWB",
		"https://www.bcferries.com/current-conditions/TSA-SGI",
		"https://www.bcferries.com/current-conditions/TSA-DUK",

		"https://www.bcferries.com/current-conditions/SWB-TSA",
		"https://www.bcferries.com/current-conditions/SWB-FUL",
		"https://www.bcferries.com/current-conditions/SWB-SGI",

		"https://www.bcferries.com/current-conditions/HSB-NAN",
		"https://www.bcferries.com/current-conditions/HSB-LNG",
		"https://www.bcferries.com/current-conditions/HSB-BOW",

		"https://www.bcferries.com/current-conditions/DUK-TSA",

		"https://www.bcferries.com/current-conditions/LNG-HSB",

		"https://www.bcferries.com/current-conditions/NAN-HSB",
	}

	// Tracks the correlating indexes between routeLinks and departureTerminals
	routeIndex := [12]int{0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 4, 5}

	departureTerminals := [6]string{
		"TSA",
		"SWB",
		"HSB",
		"DUK",
		"LNG",
		"NAN",
	}

	// Tracks the correlating indexes between route links and destinationTerminals
	destinationIndex := [12]int{0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 0, 0}

	destinationTerminals := [6][]string{
		{"SWB", "SGU", "DUK"},
		{"TSA", "FUL", "SGI"},
		{"NAN", "LNG", "BOW"},
		{"TSA"},
		{"HSB"},
		{"HSB"},
	}

	var schedule = make(map[string]map[string][]map[string]string)

	for i := 0; i < len(capacityRouteLinks); i++ {
		// Make HTTP GET request
		response, err := http.Get(capacityRouteLinks[i])
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()

		// Create a goquery document from the HTTP response
		document, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Fatal("Error loading HTTP response body. ", err)
		}

		// Array of times and capacities
		var times []string

		// Find all <p> tags and save them to array
		document.Find("p").Each(func(index int, element *goquery.Selection) {
			// Time and capacity data has an empty string as it's class
			class, exists := element.Attr("class")
			if exists && class == "" {
				// Get text
				text := element.Text()

				// Remove trailing whitespace
				text = strings.TrimSpace(text)

				// Remove text after time
				if len(text) > 15 {
					text = text[:15]
				}

				text = strings.TrimSpace(text)

				if text == "" {
					text = "Cancelled"
				}

				// Save times
				times = append(times, text)
			}
		})

		// Process array into schedule map
		for j := 0; j < len(times); j += 2 {
			sailing := map[string]string{}
			sailing["time"] = times[j]

			capacity, err := strconv.Atoi(strings.Split(times[j+1], "%")[0])

			if err == nil {
				sailing["capacity"] = strconv.Itoa(100 - capacity)
			} else {
				sailing["capacity"] = times[j+1]
			}

			departureTerminal := departureTerminals[routeIndex[i]]
			destinationTerminal := destinationTerminals[routeIndex[i]][destinationIndex[i]]

			if schedule[departureTerminal] == nil {
				schedule[departureTerminal] = make(map[string][]map[string]string)
			}

			schedule[departureTerminal][destinationTerminal] = append(schedule[departureTerminal][destinationTerminal], sailing)
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

/*
func ScrapeNonCapacityRoutes(link string) {
	nonCapacityRouteLinks := []string{
		// Metro Vancouver
		"https://www.bcferries.com/routes-fares/schedules/daily/HSB-NAN",
		"https://www.bcferries.com/routes-fares/schedules/daily/BOW-HSB",
		"https://www.bcferries.com/routes-fares/schedules/daily/HSB-BOW",
		"https://www.bcferries.com/routes-fares/schedules/daily/TSA-DUK",
		"https://www.bcferries.com/routes-fares/schedules/daily/TSA-SWB",
		"https://www.bcferries.com/routes-fares/schedules/daily/HSB-LNG",
		"https://www.bcferries.com/routes-fares/schedules/daily/TSA-POB",
		"https://www.bcferries.com/routes-fares/schedules/daily/TSA-PLH",
		"https://www.bcferries.com/routes-fares/schedules/daily/TSA-PVB",
		"https://www.bcferries.com/routes-fares/schedules/daily/TSA-PSB",
		"https://www.bcferries.com/routes-fares/schedules/daily/TSA-PST",

		// Vancouver Island
		"https://www.bcferries.com/routes-fares/schedules/daily/SWB-TSA",
		"https://www.bcferries.com/routes-fares/schedules/daily/SWB-FUL",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPH-SHW",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPH-POF",
		"https://www.bcferries.com/routes-fares/schedules/daily/CFT-VES",
		"https://www.bcferries.com/routes-fares/schedules/daily/NAN-HSB",
		"https://www.bcferries.com/routes-fares/schedules/daily/MCN-ALR",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPH-BEC",
		"https://www.bcferries.com/routes-fares/schedules/daily/DUK-TSA",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPH-PPR",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPH-PBB",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPH-KLE",
		"https://www.bcferries.com/routes-fares/schedules/daily/BTW-MIL",
		"https://www.bcferries.com/routes-fares/schedules/daily/NAH-GAB",
		"https://www.bcferries.com/routes-fares/schedules/daily/CHM-THT",
		"https://www.bcferries.com/routes-fares/schedules/daily/CHM-PEN",
		"https://www.bcferries.com/routes-fares/schedules/daily/CAM-QDR",
		"https://www.bcferries.com/routes-fares/schedules/daily/CMX-PWR",
		"https://www.bcferries.com/routes-fares/schedules/daily/BKY-DNM",
		"https://www.bcferries.com/routes-fares/schedules/daily/MCN-SOI",
		"https://www.bcferries.com/routes-fares/schedules/daily/MIL-BTW",

		// Sunshine Coast
		"https://www.bcferries.com/routes-fares/schedules/daily/ERL-SLT",
		"https://www.bcferries.com/routes-fares/schedules/daily/SLT-ERL",
		"https://www.bcferries.com/routes-fares/schedules/daily/LNG-HSB",
		"https://www.bcferries.com/routes-fares/schedules/daily/TEX-PWR",
		"https://www.bcferries.com/routes-fares/schedules/daily/PWR-TEX",
		"https://www.bcferries.com/routes-fares/schedules/daily/PWR-CMX",

		// Southern Gulf Islands
		"https://www.bcferries.com/routes-fares/schedules/daily/GAB-NAH",
		"https://www.bcferries.com/routes-fares/schedules/daily/PSB-TSA",
		"https://www.bcferries.com/routes-fares/schedules/daily/PVB-TSA",
		"https://www.bcferries.com/routes-fares/schedules/daily/POB-TSA",
		"https://www.bcferries.com/routes-fares/schedules/daily/PEN-CHM",
		"https://www.bcferries.com/routes-fares/schedules/daily/PEN-THT",
		"https://www.bcferries.com/routes-fares/schedules/daily/FUL-SWB",
		"https://www.bcferries.com/routes-fares/schedules/daily/PLH-TSA",
		"https://www.bcferries.com/routes-fares/schedules/daily/VES-CFT",
		"https://www.bcferries.com/routes-fares/schedules/daily/PST-TSA",
		"https://www.bcferries.com/routes-fares/schedules/daily/THT-CHM",
		"https://www.bcferries.com/routes-fares/schedules/daily/THT-PEN",

		// Northern Guld Islands
		"https://www.bcferries.com/routes-fares/schedules/daily/HRB-COR",
		"https://www.bcferries.com/routes-fares/schedules/daily/ALR-MCN",
		"https://www.bcferries.com/routes-fares/schedules/daily/ALR-SOI",
		"https://www.bcferries.com/routes-fares/schedules/daily/SOI-MCN",
		"https://www.bcferries.com/routes-fares/schedules/daily/DNM-BKY",
		"https://www.bcferries.com/routes-fares/schedules/daily/COR-HRB",
		"https://www.bcferries.com/routes-fares/schedules/daily/QDR-CAM",
		"https://www.bcferries.com/routes-fares/schedules/daily/SOI-ALR",
		"https://www.bcferries.com/routes-fares/schedules/daily/DNE-HRN",
		"https://www.bcferries.com/routes-fares/schedules/daily/HRN-DNE",

		// Central Coast
		"https://www.bcferries.com/routes-fares/schedules/daily/PBB-SHW",
		"https://www.bcferries.com/routes-fares/schedules/daily/SHW-PPH",
		"https://www.bcferries.com/routes-fares/schedules/daily/PBB-BEC",
		"https://www.bcferries.com/routes-fares/schedules/daily/BEC-POF",
		"https://www.bcferries.com/routes-fares/schedules/daily/BEC-PBB",
		"https://www.bcferries.com/routes-fares/schedules/daily/POF-SHW",
		"https://www.bcferries.com/routes-fares/schedules/daily/PBB-PPR",
		"https://www.bcferries.com/routes-fares/schedules/daily/SHW-POF",
		"https://www.bcferries.com/routes-fares/schedules/daily/PBB-PPH",
		"https://www.bcferries.com/routes-fares/schedules/daily/POF-BEC",
		"https://www.bcferries.com/routes-fares/schedules/daily/SHW-PBB",
		"https://www.bcferries.com/routes-fares/schedules/daily/BEC-PPH",
		"https://www.bcferries.com/routes-fares/schedules/daily/KLE-PPR",
		"https://www.bcferries.com/routes-fares/schedules/daily/PBB-POF",
		"https://www.bcferries.com/routes-fares/schedules/daily/POF-PBB",
		"https://www.bcferries.com/routes-fares/schedules/daily/BEC-SHW",
		"https://www.bcferries.com/routes-fares/schedules/daily/SHW-BEC",
		"https://www.bcferries.com/routes-fares/schedules/daily/PBB-KLE",
		"https://www.bcferries.com/routes-fares/schedules/daily/PBB-KLE",
		"https://www.bcferries.com/routes-fares/schedules/daily/POF-PPH",

		// North Coast
		"https://www.bcferries.com/routes-fares/schedules/daily/PPR-KLE",
		"https://www.bcferries.com/routes-fares/schedules/daily/KLE-PPH",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPR-PSK",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPR-PPH",
		"https://www.bcferries.com/routes-fares/schedules/daily/PPR-PBB",
		"https://www.bcferries.com/routes-fares/schedules/daily/KLE-PBB",

		// Haida Gwaii
		"https://www.bcferries.com/routes-fares/schedules/daily/ALF-PSK",
		"https://www.bcferries.com/routes-fares/schedules/daily/PSK-PPR",
		"https://www.bcferries.com/routes-fares/schedules/daily/PSK-ALF",
	}
}
*/
