package config

import (
	"log"
	"net"
	"net/url"
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
	// A local .env file is convenient for development. In Docker and other
	// production environments, configuration comes from the process environment.
	_ = godotenv.Load()

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

	u := url.URL{Scheme: "postgres", User: url.UserPassword(DB.User, DB.Password), Host: net.JoinHostPort(DB.Host, DB.Port), Path: "/" + DB.Database}
	q := url.Values{"sslmode": {DB.SSL}, "connect_timeout": {"5"}, "statement_timeout": {"10000"}}
	u.RawQuery = q.Encode()
	DB.URL = u.String()

	// Port
	ServerPort = os.Getenv("PORT")
}
