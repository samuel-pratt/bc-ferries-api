package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

func SetupRouter() *httprouter.Router {
	router := httprouter.New()

	// V2 Routes
	router.GET("/v2/", CapacityAndNonCapacitySailingsEndpoint)
	router.GET("/v2/capacity/", CapacitySailingsEndpoint)
	router.GET("/v2/noncapacity/", NonCapacitySailingsEndpoint)

	// V1 Routes
	router.GET("/api/", AllSailingsEndpoint)
	router.GET("/api/:departureTerminal/", SailingsByDepartureTerminal)
	router.GET("/api/:departureTerminal/:destinationTerminal/", SailingsByDepartureAndDestinationTerminals)

	router.GET("/healthcheck/", HealthCheck)

	router.NotFound = http.FileServer(http.Dir("./static"))

	return router
}

func SetupCron() {
	s := gocron.NewScheduler(time.UTC)

	s.Every(1).Minute().Do(func() {
		ScrapeCapacityRoutes()
	})

	s.Every(1).Hour().Do(func() {
		// ScrapeNonCapacityRoutes()
	})

	s.StartAsync()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: File not found")
	}

	if os.Getenv("PGHOST") == "" || os.Getenv("PGPORT") == "" || os.Getenv("PGUSER") == "" || os.Getenv("PGPASSWORD") == "" || os.Getenv("PGDATABASE") == "" {
		log.Fatal("Error loading .env file: Missing variables")
	}

	db := GetPostgresInstance()
	defer db.Close()

	SetupCron()

	var port = os.Getenv("PORT")

	if port == "" {
		port = "4747"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}

	router := SetupRouter()

	http.ListenAndServe(":"+port, router)
}
