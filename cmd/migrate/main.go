package main

import (
	"fmt"
	"log"

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

	// Run migrations
	fmt.Println("Running MAWB System Integration migrations...")
	if err := database.RunMigrationsForMAWBSystem(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("All migrations completed successfully!")
}
