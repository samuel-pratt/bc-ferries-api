package db

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/samuel-pratt/bc-ferries-api/cmd/models"
)

/*
 * GetCapacitySailings
 *
 * Retrieves all capacity route records from the database, including parsed sailing data.
 *
 * Queries the `capacity_routes` table and unmarshals the `sailings` JSON column
 * into a slice of `models.CapacitySailing` for each route.
 *
 * @return []models.CapacityRoute - a slice of capacity routes with their sailings
 */
func GetCapacitySailings() []models.CapacityRoute {
	routes := []models.CapacityRoute{}

	sqlStatement := `SELECT r.route_code,r.from_terminal_code,r.to_terminal_code,r.sailing_duration,r.sailings,COALESCE(o.payload,'{}') FROM capacity_routes r LEFT JOIN route_operational_data o ON o.kind='capacity' AND o.route_code=r.route_code ORDER BY r.route_code`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := Conn.QueryContext(ctx, sqlStatement)
	if err != nil {
		log.Printf("GetCapacitySailings: query failed: %v", err)
		return routes
	}
	defer rows.Close()

	for rows.Next() {
		var route models.CapacityRoute
		var sailings []uint8
		var operationalData []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &route.SailingDuration, &sailings, &operationalData)
		if err != nil {
			log.Printf("GetCapacitySailings: row scan failed: %v", err)
			return nil
		}

		var content []models.CapacitySailing
		if err := json.Unmarshal(sailings, &content); err != nil {
			log.Printf("GetCapacitySailings: JSON unmarshal failed: %v", err)
			return nil
		}

		route.Sailings = content
		var operational models.RouteOperationalData
		if len(operationalData) > 0 && json.Unmarshal(operationalData, &operational) == nil {
			route.Notices = operational.Notices
			route.Terminal = operational.Terminal
			route.ObservedAt = operational.ObservedAt
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetCapacitySailings: row iteration error: %v", err)
		return nil
	}

	return routes
}

/*
 * GetNonCapacitySailings
 *
 * Retrieves all non-capacity route records from the database, including parsed sailing data.
 *
 * Queries the `non_capacity_routes` table and unmarshals the `sailings` JSON column
 * into a slice of `models.NonCapacitySailing` for each route.
 *
 * @return []models.NonCapacityRoute - a slice of non-capacity routes with their sailings
 */
func GetNonCapacitySailings() []models.NonCapacityRoute {
	routes := []models.NonCapacityRoute{}

	sqlStatement := `SELECT r.route_code,r.from_terminal_code,r.to_terminal_code,r.sailing_duration,r.sailings,COALESCE(o.payload,'{}') FROM non_capacity_routes r LEFT JOIN route_operational_data o ON o.kind='noncapacity' AND o.route_code=r.route_code ORDER BY r.route_code`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := Conn.QueryContext(ctx, sqlStatement)
	if err != nil {
		log.Printf("GetNonCapacitySailings: query failed: %v", err)
		return routes
	}
	defer rows.Close()

	for rows.Next() {
		var route models.NonCapacityRoute
		var sailings []uint8
		var operationalData []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &route.SailingDuration, &sailings, &operationalData)
		if err != nil {
			log.Printf("GetNonCapacitySailings: row scan failed: %v", err)
			return nil
		}

		var content []models.NonCapacitySailing
		if err := json.Unmarshal(sailings, &content); err != nil {
			log.Printf("GetNonCapacitySailings: JSON unmarshal failed: %v", err)
			return nil
		}

		route.Sailings = content
		var operational models.RouteOperationalData
		if len(operationalData) > 0 && json.Unmarshal(operationalData, &operational) == nil {
			route.Notices = operational.Notices
			route.Terminal = operational.Terminal
			route.ObservedAt = operational.ObservedAt
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetNonCapacitySailings: row iteration error: %v", err)
		return nil
	}

	return routes
}
