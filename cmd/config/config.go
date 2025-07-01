package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
	SSL      string
	URL      string
}

var (
	DB         DBConfig
	ServerPort string
)

/*
 * LoadEnv
 *
 * Loads environment variables from a `.env` file using godotenv.
 *
 * Populates the DB configuration and server port. Constructs the database URL
 * using the retrieved values. Logs a fatal error and exits if any required DB
 * variables are missing or if the `.env` file cannot be loaded.
 *
 * @return void
 */
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// DB config
	DB = DBConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Database: os.Getenv("DB_NAME"),
		SSL:      os.Getenv("DB_SSL"),
	}

	if DB.User == "" || DB.Password == "" || DB.Host == "" || DB.Port == "" || DB.Database == "" || DB.SSL == "" {
		log.Fatal("Missing required SQL environment variables")
	}

	DB.URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", DB.User, DB.Password, DB.Host, DB.Port, DB.Database, DB.SSL)

	// Port
	ServerPort = os.Getenv("PORT")
}
