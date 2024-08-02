package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type postgresSingleton struct {
	db *sql.DB
}

var postgresInstance *postgresSingleton
var postgresOnce sync.Once

func GetPostgresInstance() *sql.DB {
	postgresOnce.Do(func() {
		postgresInstance = &postgresSingleton{}

		var (
			DB_HOST = os.Getenv("DB_HOST")
			DB_PORT = os.Getenv("DB_PORT")
			DB_USER = os.Getenv("DB_USER")
			DB_PASS = os.Getenv("DB_PASS")
			DB_NAME = os.Getenv("DB_NAME")
			DB_SSL  = os.Getenv("DB_SSL")
		)

		postgresqlDbInfo := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=%s",
			DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, DB_SSL)

		db, err := sql.Open("postgres", postgresqlDbInfo)
		if err != nil {
			log.Fatal(err)
		}

		postgresInstance.db = db
	})

	return postgresInstance.db
}

func GetCapacitySailings() []CapacityRoute {
	db := GetPostgresInstance()

	var routes []CapacityRoute

	sqlStatement := `SELECT * FROM capacity_routes`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var route CapacityRoute
		var sailings []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &route.SailingDuration, &sailings)
		if err != nil {
			log.Fatal(err)
		}

		content := []CapacitySailing{}
		json.Unmarshal([]byte(sailings), &content)

		route.Sailings = content

		routes = append(routes, route)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return routes
}

func GetNonCapacitySailings() []NonCapacityRoute {
	db := GetPostgresInstance()

	var routes []NonCapacityRoute

	sqlStatement := `SELECT * FROM non_capacity_routes`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var route NonCapacityRoute
		var sailings []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &route.SailingDuration, &sailings)
		if err != nil {
			log.Fatal(err)
		}

		content := []NonCapacitySailing{}
		json.Unmarshal([]byte(sailings), &content)

		route.Sailings = content

		routes = append(routes, route)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return routes
}
