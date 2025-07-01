package db

import (
	"encoding/json"
	"log"

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
	var routes []models.CapacityRoute

	sqlStatement := `SELECT * FROM capacity_routes`

	rows, err := Conn.Query(sqlStatement)
	if err != nil {
		log.Printf("GetCapacitySailings: query failed: %v", err)
		return routes
	}
	defer rows.Close()

	for rows.Next() {
		var route models.CapacityRoute
		var sailings []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &route.SailingDuration, &sailings)
		if err != nil {
			log.Printf("GetCapacitySailings: row scan failed: %v", err)
			continue
		}

		var content []models.CapacitySailing
		if err := json.Unmarshal(sailings, &content); err != nil {
			log.Printf("GetCapacitySailings: JSON unmarshal failed: %v", err)
			continue
		}

		route.Sailings = content
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetCapacitySailings: row iteration error: %v", err)
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
	var routes []models.NonCapacityRoute

	sqlStatement := `SELECT * FROM non_capacity_routes`

	rows, err := Conn.Query(sqlStatement)
	if err != nil {
		log.Printf("GetNonCapacitySailings: query failed: %v", err)
		return routes
	}
	defer rows.Close()

	for rows.Next() {
		var route models.NonCapacityRoute
		var sailings []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &route.SailingDuration, &sailings)
		if err != nil {
			log.Printf("GetNonCapacitySailings: row scan failed: %v", err)
			continue
		}

		var content []models.NonCapacitySailing
		if err := json.Unmarshal(sailings, &content); err != nil {
			log.Printf("GetNonCapacitySailings: JSON unmarshal failed: %v", err)
			continue
		}

		route.Sailings = content
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetNonCapacitySailings: row iteration error: %v", err)
	}

	return routes
}
