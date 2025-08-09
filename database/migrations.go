package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
)

// Migration represents a database migration
type Migration struct {
	Version string
	Name    string
	SQL     string
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db *pg.DB
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *pg.DB) *MigrationRunner {
	return &MigrationRunner{db: db}
}

// RunMigrations executes all pending migrations
func (mr *MigrationRunner) RunMigrations(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Load migration files
	migrations, err := mr.loadMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to load migration files: %v", err)
	}

	// Get applied migrations
	appliedMigrations, err := mr.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Execute pending migrations
	for _, migration := range migrations {
		if !contains(appliedMigrations, migration.Version) {
			if err := mr.executeMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to execute migration %s: %v", migration.Version, err)
			}
			fmt.Printf("Applied migration: %s - %s\n", migration.Version, migration.Name)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (mr *MigrationRunner) createMigrationsTable(ctx context.Context) error {
	sql := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := mr.db.ExecContext(ctx, sql)
	return err
}

// loadMigrationFiles loads all migration files from the migrations directory
func (mr *MigrationRunner) loadMigrationFiles() ([]Migration, error) {
	migrationsDir := "database/migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			// Skip rollback migrations by default
			if strings.Contains(file.Name(), "rollback") {
				continue
			}

			// Parse version and name from filename (e.g., "001_create_cargo_manifest_tables.sql")
			parts := strings.SplitN(file.Name(), "_", 2)
			if len(parts) != 2 {
				continue
			}

			version := parts[0]
			name := strings.TrimSuffix(parts[1], ".sql")
			name = strings.ReplaceAll(name, "_", " ")

			// Read file content
			content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read migration file %s: %v", file.Name(), err)
			}

			migrations = append(migrations, Migration{
				Version: version,
				Name:    name,
				SQL:     string(content),
			})
		}
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a list of applied migration versions
func (mr *MigrationRunner) getAppliedMigrations(ctx context.Context) ([]string, error) {
	var versions []string
	_, err := mr.db.QueryContext(ctx, &versions, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// executeMigration executes a single migration
func (mr *MigrationRunner) executeMigration(ctx context.Context, migration Migration) error {
	// Start transaction
	tx, err := mr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	_, err = tx.ExecContext(ctx, migration.SQL)
	if err != nil {
		return err
	}

	// Record migration as applied
	_, err = tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
		migration.Version, migration.Name)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RunMigrationsForMAWBSystem runs migrations specifically for the MAWB system integration
func RunMigrationsForMAWBSystem(db *pg.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runner := NewMigrationRunner(db)
	return runner.RunMigrations(ctx)
}
