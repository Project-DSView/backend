package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbURL         = flag.String("db-url", "", "Database connection URL (required)")
	migrationsDir = flag.String("migrations-dir", "./migrations", "Directory containing migration files")
	up            = flag.Bool("up", false, "Run migrations up")
	down          = flag.Bool("down", false, "Run migrations down")
	version       = flag.Bool("version", false, "Show current migration version")
)

func main() {
	flag.Parse()

	if *dbURL == "" {
		// Try to get from environment
		*dbURL = os.Getenv("DATABASE_URL")
		if *dbURL == "" {
			// Construct from individual env vars
			dbHost := getEnv("DB_HOST", "localhost")
			dbPort := getEnv("DB_PORT", "5432")
			dbUser := getEnv("DB_USER", "postgres")
			dbPassword := getEnv("DB_PASSWORD", "db_password")
			dbName := getEnv("DB_NAME", "dsview_db")
			*dbURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				dbHost, dbPort, dbUser, dbPassword, dbName)
		}
	}

	if *dbURL == "" {
		log.Fatal("Database URL is required. Use -db-url flag or set DATABASE_URL environment variable")
	}

	db, err := gorm.Open(postgres.Open(*dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	if *version {
		showVersion(db)
		return
	}

	if *up {
		if err := runMigrationsUp(db, *migrationsDir); err != nil {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
		return
	}

	if *down {
		log.Println("Down migrations not implemented yet")
		return
	}

	// Default: show help
	flag.Usage()
}

func createMigrationsTable(db *gorm.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`
	return db.Exec(query).Error
}

func showVersion(db *gorm.DB) {
	var version string
	err := db.Raw("SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1").Scan(&version).Error
	if err == gorm.ErrRecordNotFound {
		fmt.Println("No migrations applied")
		return
	}
	if err != nil {
		log.Fatalf("Failed to get migration version: %v", err)
	}
	fmt.Printf("Current migration version: %s\n", version)
}

func runMigrationsUp(db *gorm.DB, migrationsDir string) error {
	// Get list of migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort migration files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Get already applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run each migration
	for _, filename := range migrationFiles {
		version := strings.TrimSuffix(filename, ".sql")

		// Skip if already applied
		if appliedMigrations[version] {
			log.Printf("Skipping migration %s (already applied)", version)
			continue
		}

		log.Printf("Running migration: %s", filename)

		filepath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration in transaction
		err = db.Transaction(func(tx *gorm.DB) error {
			// Execute migration SQL
			if err := tx.Exec(string(content)).Error; err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", filename, err)
			}

			// Record migration
			if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", version, err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		log.Printf("Successfully applied migration: %s", version)
	}

	log.Println("All migrations completed successfully")
	return nil
}

func getAppliedMigrations(db *gorm.DB) (map[string]bool, error) {
	var versions []string
	if err := db.Raw("SELECT version FROM schema_migrations").Scan(&versions).Error; err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, version := range versions {
		applied[version] = true
	}
	return applied, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
