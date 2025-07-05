package main

import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/samuel-pratt/bc-ferries-api/cmd/config"
	"github.com/samuel-pratt/bc-ferries-api/cmd/cron"
	"github.com/samuel-pratt/bc-ferries-api/cmd/db"
	"github.com/samuel-pratt/bc-ferries-api/cmd/router"
)

func main() {
	// Set up environment variables, database connection
	config.LoadEnv()
	db.Init()
	defer db.Conn.Close()

	cron.SetupCron()

	if config.ServerPort == "" {
		config.ServerPort = "8080"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + config.ServerPort)
	}

	router := router.SetupRouter()
	http.ListenAndServe(":"+config.ServerPort, router)
}
