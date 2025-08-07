package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	Mode               string
	PostgreSQLHost     string
	PostgreSQLUser     string
	PostgreSQLPassword string
	PostgreSQLName     string
	PostgreSQLPort     string
	PostgreSQLSSLMode  string
	GCSProjectID       string
	GCSBucketName      string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	} else {
		log.Println("We are getting the env values")
	}

	config := Config{
		Port:               os.Getenv("PORT"),
		Mode:               os.Getenv("MODE"),
		PostgreSQLHost:     os.Getenv("POSTGRESQL_HOST"),
		PostgreSQLUser:     os.Getenv("POSTGRESQL_USER"),
		PostgreSQLPassword: os.Getenv("POSTGRESQL_PASSWORD"),
		PostgreSQLName:     os.Getenv("POSTGRESQL_NAME"),
		PostgreSQLPort:     os.Getenv("POSTGRESQL_PORT"),
		PostgreSQLSSLMode:  os.Getenv("POSTGRESQL_SSLMODE"),
		GCSProjectID:       os.Getenv("GCS_PROJECT_ID"),
		GCSBucketName:      os.Getenv("GCS_BUCKET_NAME"),
	}

	return &config
}
