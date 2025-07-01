package db

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/samuel-pratt/bc-ferries-api/cmd/config"
)

var Conn *sql.DB

/*
 * Init
 *
 * Initializes the global PostgreSQL database connection using the DSN from config.DB.URL.
 *
 * Opens a connection pool and assigns it to the Conn variable.
 * Panics if the connection cannot be established.
 *
 * @return void
 */
func Init() {
	var err error
	Conn, err = sql.Open("postgres", config.DB.URL)
	if err != nil {
		panic(err)
	}
}
