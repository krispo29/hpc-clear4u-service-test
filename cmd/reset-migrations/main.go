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

	// Remove the rollback migration from applied migrations
	fmt.Println("Removing rollback migration from applied migrations...")
	_, err = db.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = '003'")
	if err != nil {
		log.Fatalf("Failed to remove rollback migration: %v", err)
	}

	// Run migrations again
	fmt.Println("Running migrations again...")
	if err := database.RunMigrationsForMAWBSystem(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migrations reset and reapplied successfully!")
}
