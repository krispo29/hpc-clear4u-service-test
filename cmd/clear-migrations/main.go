package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"hpc-express-service/config"
	"hpc-express-service/database"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to PostgreSQL
	db, err := database.NewPostgreSQLConnection(
		cfg.PostgreSQLUser,
		cfg.PostgreSQLPassword,
		cfg.PostgreSQLName,
		cfg.PostgreSQLHost,
		cfg.PostgreSQLPort,
		cfg.PostgreSQLSSLMode,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clear all migration records
	fmt.Println("Clearing all migration records...")
	_, err = db.ExecContext(ctx, "DELETE FROM schema_migrations")
	if err != nil {
		log.Fatalf("Failed to clear migrations: %v", err)
	}

	fmt.Println("Migration records cleared successfully!")
}
