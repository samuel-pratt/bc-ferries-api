package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/log"
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
			DB_HOST = os.Getenv("PGHOST")
			DB_PORT = os.Getenv("PGPORT")
			DB_USER = os.Getenv("PGUSER")
			DB_PASS = os.Getenv("PGPASSWORD")
			DB_NAME = os.Getenv("PGDATABASE")
		)

		postgresqlDbInfo := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=require",
			DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME)

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
		log.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var route CapacityRoute
		var sailings []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &sailings)
		if err != nil {
			log.Error(err)
		}

		content := []CapacitySailing{}
		json.Unmarshal([]byte(sailings), &content)

		route.Sailings = content

		routes = append(routes, route)
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
	}

	return routes
}

func GetNonCapacitySailings() []NonCapacityRoute {
	db := GetPostgresInstance()

	var routes []NonCapacityRoute

	sqlStatement := `SELECT * FROM non_capacity_routes`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var route NonCapacityRoute
		var sailings []uint8

		err := rows.Scan(&route.RouteCode, &route.FromTerminalCode, &route.ToTerminalCode, &sailings)
		if err != nil {
			log.Error(err)
		}

		content := []NonCapacitySailing{}
		json.Unmarshal([]byte(sailings), &content)

		route.Sailings = content

		routes = append(routes, route)
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
	}

	return routes
}
