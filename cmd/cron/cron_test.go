package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCapacityDue(t *testing.T) {
	start := time.Date(2026, 9, 21, 18, 1, 0, 0, time.UTC)
	job := scrapeJob{interval: time.Minute}
	if !job.due(start) {
		t.Fatal("must scrape on startup")
	}
	job.lastStart = start
	for _, tc := range []struct {
		name   string
		offset time.Duration
		want   bool
	}{
		{"before interval", 59 * time.Second, false},
		{"one minute", time.Minute, true},
		{"wake after a day", 24 * time.Hour, true},
		{"clock moved back", -time.Hour, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := job.due(start.Add(tc.offset)); got != tc.want {
				t.Fatalf("due=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestScheduleMidnightAndDST(t *testing.T) {
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		t.Fatal(err)
	}
	for _, day := range []string{"2026-09-22", "2026-03-08", "2026-11-01"} {
		t.Run(day, func(t *testing.T) {
			midnight, err := time.ParseInLocation("2006-01-02", day, loc)
			if err != nil {
				t.Fatal(err)
			}
			job := scrapeJob{interval: 4 * time.Hour, location: loc, lastStart: midnight.Add(-time.Hour)}
			if job.due(midnight) {
				t.Fatal("midnight refresh should wait until 00:01")
			}
			if !job.due(midnight.Add(time.Minute)) {
				t.Fatal("must refresh the new service day at 00:01")
			}
			job.lastStart = midnight.Add(time.Minute)
			if job.due(job.lastStart.Add(time.Hour)) {
				t.Fatal("must not repeat the daily refresh")
			}
			if !job.due(job.lastStart.Add(4 * time.Hour)) {
				t.Fatal("four-hour interval must still run")
			}
		})
	}
}

func TestRunJobRecoversFromClockJumpWithoutOverlapOrCatchupBurst(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	base := time.Date(2026, 9, 22, 1, 0, 0, 0, time.UTC)
	var elapsed atomic.Int64
	now := func() time.Time { return base.Add(time.Duration(elapsed.Load())) }
	ticks := make(chan time.Time, 1)
	started := make(chan struct{}, 10)
	release := make(chan struct{}, 10)
	job := &scrapeJob{interval: time.Minute, run: func() { started <- struct{}{}; <-release }}
	done := make(chan struct{})
	go func() { defer close(done); runJob(ctx, ticks, now, job) }()
	awaitRun := func() {
		t.Helper()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("refresh did not start")
		}
	}
	awaitRun()
	// A tick during an active scrape must not start another copy.
	elapsed.Store(int64(time.Minute))
	ticks <- base
	select {
	case <-started:
		t.Fatal("overlapping scrape")
	case <-time.After(20 * time.Millisecond):
	}
	release <- struct{}{}
	awaitRun()
	// Simulate a queued, old timer tick after a 24-hour host/VM sleep.
	elapsed.Store(int64(24 * time.Hour))
	ticks <- base
	release <- struct{}{}
	awaitRun()
	release <- struct{}{}
	ticks <- base
	select {
	case <-started:
		t.Fatal("replayed missed intervals")
	case <-time.After(20 * time.Millisecond):
	}
	// Normal once-per-minute refresh must resume after the jump.
	elapsed.Store(int64(24*time.Hour + time.Minute))
	ticks <- base
	awaitRun()
	cancel()
	release <- struct{}{}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop")
	}
}
