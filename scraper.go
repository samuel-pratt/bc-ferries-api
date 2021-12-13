package main

// Import OS and fmt packages
import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Response struct {
	Schedule  map[string]map[string][]map[string]string `json:"schedule"`
	ScrapedAt time.Time                                 `json:"scrapedAt"`
}

func scraper() Response {
	// Links to individual schedule pages
	routeLinks := [12]string{
		"https://www.bcferries.com/current-conditions/vancouver-tsawwassen-victoria-swartz-bay/TSA-SWB",
		"https://www.bcferries.com/current-conditions/vancouver-tsawwassen-southern-gulf-islands/TSA-SGI",
		"https://www.bcferries.com/current-conditions/vancouver-tsawwassen-nanaimo-duke-point/TSA-DUK",

		"https://www.bcferries.com/current-conditions/victoria-swartz-bay-vancouver-tsawwassen/SWB-TSA",
		"https://www.bcferries.com/current-conditions/victoria-swartz-bay-salt-spring-island-fulford-harbour/SWB-FUL",
		"https://www.bcferries.com/current-conditions/victoria-swartz-bay-southern-gulf-islands/SWB-SGI",

		"https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-nanaimo-departure-bay/HSB-NAN",
		"https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-sunshine-coast-langdale/HSB-LNG",
		"https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-bowen-island-snug-cove/HSB-BOW",

		"https://www.bcferries.com/current-conditions/nanaimo-duke-point-vancouver-tsawwassen/DUK-TSA",

		"https://www.bcferries.com/current-conditions/sunshine-coast-langdale-vancouver-horseshoe-bay/LNG-HSB",

		"https://www.bcferries.com/current-conditions/nanaimo-departure-bay-vancouver-horseshoe-bay/NAN-HSB",
	}

	// Tracks the correlating indexes between routeLinks and departureTerminals
	routeIndex := [12]int{0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 4, 5}

	departureTerminals := [6]string{
		"tsawwassen",
		"swartz-bay",
		"horseshoe-bay",
		"nanaimo-(duke-pt)",
		"langdale",
		"nanaimo-(dep-bay)",
	}

	// Tracks the correlating indexes between route links and destinationTerminals
	destinationIndex := [12]int{0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 0, 0}

	destinationTerminals := [6][]string{
		{"swartz-bay", "southern-gulf-islands", "nanaimo-(duke-pt)"},
		{"tsawwassen", "fulford-habrbour-(saltspring)", "southern-gulf-islands"},
		{"nanaimo-(dep-bay)", "langdale", "snug-cove-(bowen)"},
		{"tsawwassen"},
		{"horseshoe-bay"},
		{"horseshoe-bay"},
	}

	var schedule = make(map[string]map[string][]map[string]string)

	for i := 0; i < len(routeLinks); i++ {
		// Make HTTP GET request
		response, err := http.Get(routeLinks[i])
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

				// Remove tomorrow text
				text = strings.ReplaceAll(text, "(Tomorrow)", "")

				// Remove trailing whitespace
				text = strings.TrimSpace(text)

				// Save times
				times = append(times, text)
			}
		})

		// Process array into schedule map
		for j := 0; j < len(times); j += 2 {
			sailing := map[string]string{}
			sailing["time"] = times[j]
			sailing["capacity"] = times[j+1]

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
