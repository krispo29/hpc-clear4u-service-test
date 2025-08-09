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

	// Check if tables exist
	tables := []string{
		"cargo_manifest",
		"cargo_manifest_items",
		"draft_mawb",
		"draft_mawb_items",
		"draft_mawb_item_dims",
		"draft_mawb_charges",
		"schema_migrations",
	}

	fmt.Println("Verifying MAWB System Integration database schema...")
	fmt.Println("====================================================")

	for _, table := range tables {
		var exists bool
		_, err := db.QueryOneContext(ctx, &exists, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_name = ?
			)
		`, table)

		if err != nil {
			fmt.Printf("❌ Error checking table %s: %v\n", table, err)
			continue
		}

		if exists {
			fmt.Printf("✅ Table %s exists\n", table)
		} else {
			fmt.Printf("❌ Table %s does not exist\n", table)
		}
	}

	// Check applied migrations
	fmt.Println("\nApplied migrations:")
	fmt.Println("--------------------")

	type Migration struct {
		Version   string    `pg:"version"`
		Name      string    `pg:"name"`
		AppliedAt time.Time `pg:"applied_at"`
	}

	var migrations []Migration
	_, err = db.QueryContext(ctx, &migrations, "SELECT version, name, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		fmt.Printf("❌ Error fetching migrations: %v\n", err)
	} else {
		for _, migration := range migrations {
			fmt.Printf("✅ %s - %s (applied: %s)\n",
				migration.Version,
				migration.Name,
				migration.AppliedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Check foreign key constraints
	fmt.Println("\nForeign key constraints:")
	fmt.Println("-------------------------")

	var constraintCount int
	_, err = db.QueryOneContext(ctx, &constraintCount, `
		SELECT COUNT(*)
		FROM information_schema.table_constraints AS tc
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_name IN ('cargo_manifest', 'cargo_manifest_items', 'draft_mawb', 'draft_mawb_items', 'draft_mawb_item_dims', 'draft_mawb_charges')
	`)

	if err != nil {
		fmt.Printf("❌ Error fetching foreign key constraints: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d foreign key constraints\n", constraintCount)

		// List expected constraints
		expectedConstraints := []string{
			"cargo_manifest -> tbl_mawb_info (CASCADE DELETE)",
			"cargo_manifest_items -> cargo_manifest (CASCADE DELETE)",
			"draft_mawb -> tbl_mawb_info (CASCADE DELETE)",
			"draft_mawb_items -> draft_mawb (CASCADE DELETE)",
			"draft_mawb_item_dims -> draft_mawb_items (CASCADE DELETE)",
			"draft_mawb_charges -> draft_mawb (CASCADE DELETE)",
		}

		fmt.Println("\nExpected foreign key relationships:")
		for _, constraint := range expectedConstraints {
			fmt.Printf("  • %s\n", constraint)
		}
	}

	fmt.Println("\nSchema verification completed!")
}
