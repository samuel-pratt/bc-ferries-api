package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
	"log"
	"strings"
	"time"
)

func SaveSnapshot(kind, key string, value any) error {
	b, e := json.Marshal(value)
	if e != nil {
		return e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, e = Conn.ExecContext(ctx, `INSERT INTO source_snapshots(kind,source_key,payload) VALUES($1,$2,$3) ON CONFLICT(kind,source_key) DO UPDATE SET payload=EXCLUDED.payload`, kind, key, b)
	return e
}
func Snapshots(kind string) map[string]json.RawMessage {
	result := map[string]json.RawMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, e := Conn.QueryContext(ctx, `SELECT source_key,payload FROM source_snapshots WHERE kind=$1 ORDER BY source_key`, kind)
	if e != nil {
		log.Printf("snapshot query failed: %v", e)
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var b []byte
		if e = rows.Scan(&key, &b); e != nil {
			return result
		}
		result[key] = b
	}
	return result
}

// An unsuccessful observation must never advance the last successful timestamp
// or overwrite cached sailings. Missing routes get an explicit unavailable record.
func MarkRouteUnavailable(kind, from, to, source string, at time.Time) error {
	table := "capacity_routes"
	if kind == "noncapacity" {
		table = "non_capacity_routes"
	} else if kind != "capacity" {
		return fmt.Errorf("invalid kind")
	}
	state, _ := json.Marshal(models.DataState{DataStatus: "unavailable", CheckedAt: &at, SourceURL: source})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, e := Conn.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(ctx, "INSERT INTO "+table+` (route_code,from_terminal_code,to_terminal_code,sailing_duration,sailings) VALUES($1,$2,$3,'','[]') ON CONFLICT DO NOTHING`, from+to, from, to); e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, `INSERT INTO route_operational_data(kind,route_code,payload) VALUES($1,$2,$3) ON CONFLICT(kind,route_code) DO UPDATE SET payload=route_operational_data.payload || EXCLUDED.payload`, kind, from+to, state); e != nil {
		return e
	}
	return tx.Commit()
}

func GetSpecialServices() []models.SpecialService {
	out := []models.SpecialService{}
	for _, b := range Snapshots("special") {
		var s models.SpecialService
		if json.Unmarshal(b, &s) == nil {
			out = append(out, s)
		}
	}
	return out
}

func attachConditions(routes []models.NonCapacityRoute) {
	snapshots := Snapshots("conditions")
	for i := range routes {
		r := &routes[i]
		var c models.ConditionsSnapshot
		b, ok := snapshots[r.RouteCode]
		if !ok || json.Unmarshal(b, &c) != nil {
			continue
		}
		r.CurrentConditions = &c
		seen := map[string]bool{}
		for _, n := range r.Notices {
			seen[n.ID] = true
		}
		for _, n := range c.Notices {
			if !seen[n.ID] {
				r.Notices = append(r.Notices, n)
				seen[n.ID] = true
			}
		}
		if c.Terminal != nil {
			r.Terminal = c.Terminal
		}
		// Schedule time/arrival remain scheduled values; actual movement is additive.
		if c.DataStatus != "available" || time.Since(c.ObservedAt) > 10*time.Minute {
			continue
		}
		for j := range r.Sailings {
			s := &r.Sailings[j]
			var matches []models.CapacitySailing
			for _, live := range c.Sailings {
				if s.ServiceDate != "" && s.ServiceDate == live.ServiceDate && strings.EqualFold(s.DepartureTime, live.ScheduledTime) {
					matches = append(matches, live)
				}
			}
			if len(matches) != 1 {
				continue
			}
			live := matches[0]
			s.Status = live.Status
			s.StatusText = live.StatusText
			s.VesselName = live.VesselName
			s.DelayMinutes = live.DelayMinutes
			s.ActualDepartureTime = live.ActualDepartureTime
			s.EstimatedArrivalTime = live.EstimatedArrivalTime
			s.ActualArrivalTime = live.ActualArrivalTime
			s.StatusObservedAt = &c.ObservedAt
		}
	}
}
