package cron

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/samuel-pratt/bc-ferries-api/cmd/scraper"
)

// Scheduler runs each scrape independently, with no overlapping runs of a job.
type Scheduler struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type scrapeJob struct {
	name      string
	interval  time.Duration
	location  *time.Location // also refresh at 00:01 local time, when set
	lastStart time.Time
	run       func()
}

func (j *scrapeJob) due(now time.Time) bool {
	// Strip monotonic readings: elapsed wall time must include host/VM sleep.
	now = now.UTC()
	last := j.lastStart.UTC()
	if j.lastStart.IsZero() || now.Before(last) || now.Sub(last) >= j.interval {
		return true
	}
	if j.location != nil {
		serviceDay := func(t time.Time) string {
			return t.In(j.location).Add(-time.Minute).Format("2006-01-02")
		}
		return serviceDay(now) != serviceDay(last)
	}
	return false
}

func runJob(ctx context.Context, ticks <-chan time.Time, now func() time.Time, job *scrapeJob) {
	for {
		if ctx.Err() != nil {
			return
		}
		at := now()
		if job.due(at) {
			job.lastStart = at.UTC()
			log.Printf("scrape %s started", job.name)
			job.run()
			log.Printf("scrape %s finished in %s", job.name, time.Since(at).Round(time.Millisecond))
		}
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ticks:
			if !ok {
				return
			}
			// Read the current clock, not a queued tick's pre-sleep timestamp.
		}
	}
}

// SetupCron refreshes immediately, then capacity every minute and schedules
// every four hours (also at 00:01 Vancouver time). An overdue job runs once
// from the current time, never waits out its lateness or replays missed runs.
func SetupCron() *Scheduler {
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{cancel: cancel}
	jobs := []*scrapeJob{
		{name: "capacity", interval: time.Minute, run: scraper.ScrapeCapacityRoutes},
		{name: "schedules", interval: 4 * time.Hour, location: loc, run: scraper.ScrapeNonCapacityRoutes},
	}
	for _, job := range jobs {
		s.wg.Add(1)
		go func(job *scrapeJob) {
			defer s.wg.Done()
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			runJob(ctx, ticker.C, time.Now, job)
		}(job)
	}
	return s
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}
