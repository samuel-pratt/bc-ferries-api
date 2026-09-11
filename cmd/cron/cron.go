package cron

import (
	"time"

	"github.com/go-co-op/gocron"
	"github.com/samuel-pratt/bc-ferries-api/cmd/scraper"
)

/*
 * SetupCron
 *
 * Initializes and starts scheduled background scraping tasks using gocron.
 *
 * - Scrapes capacity route data every 1 minute.
 * - Scrapes non-capacity route data every 4 hours.
 *
 * The scheduler runs asynchronously in the background.
 *
 * @return void
 */
func SetupCron() *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	s.Every(1).Minute().SingletonMode().Do(func() {
		scraper.ScrapeCapacityRoutes()
	})

	s.Every(4).Hour().SingletonMode().Do(func() {
		scraper.ScrapeNonCapacityRoutes()
	})

	s.StartAsync()
	return s
}
