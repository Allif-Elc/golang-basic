package main

import (
	"context"
	"flag"
	"fmt"
	"golang-basic/api/internal/config"
	"golang-basic/api/internal/migration"
	"golang-basic/api/internal/routes"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	migrateFlag     = flag.Bool("migrate", false, "Run pending database migrations and exit")
	migrateDownFlag = flag.Bool("migrate-down", false, "Rollback the last migration and exit")
	migrateStatus   = flag.Bool("migrate-status", false, "Show migration status and exit")
	migrateOneFlag  = flag.String("migrate-one", "", "Run a specific migration by name, version, or filename and exit")
)

func main() {
	flag.Parse()

	if err := config.InitDatabase(); err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
		if *migrateFlag || *migrateDownFlag || *migrateStatus || *migrateOneFlag != "" {
			log.Fatalf("Database connection required for migration operation: %v", err)
		}
		log.Println("Server will continue running without database connection")
	} else {
		defer config.CloseDatabase()
	}

	ctx := context.Background()
	migrator, err := migration.NewFromFiles(config.DB, "sql")
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	// Check migrate-one first (most specific)
	if *migrateOneFlag != "" {
		if err := runOneMigration(ctx, migrator, *migrateOneFlag); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		os.Exit(0)
	}

	if *migrateFlag {
		if err := runMigrations(ctx, migrator); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		os.Exit(0)
	}

	if *migrateDownFlag {
		if err := rollbackMigration(ctx, migrator); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		os.Exit(0)
	}

	if *migrateStatus {
		if err := showMigrationStatus(ctx, migrator); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		os.Exit(0)
	}

	if err := config.InitMinio(); err != nil {
		log.Printf("Warning: MinIO not available: %v", err)
	}

	r := routes.SetupRoutes()

	port := ":3003"
	fmt.Printf("Server is running on HTTPS port %s\n", port)
	log.Fatal(http.ListenAndServeTLS(port, "server.crt", "server.key", r))
}

func runMigrations(ctx context.Context, m *migration.Migrator) error {
	if err := m.Validate(ctx); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	log.Println("Running database migrations...")
	if err := m.Up(ctx); err != nil {
		return err
	}

	version, err := m.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if version != nil {
		log.Printf("Database is now at migration version: %d\n", *version)
	} else {
		log.Println("No migrations applied yet")
	}

	return nil
}

func rollbackMigration(ctx context.Context, m *migration.Migrator) error {
	log.Println("Rolling back last migration...")
	if err := m.Down(ctx); err != nil {
		return err
	}

	version, err := m.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if version != nil {
		log.Printf("Database is now at migration version: %d\n", *version)
	} else {
		log.Println("All migrations have been rolled back")
	}

	return nil
}

func showMigrationStatus(ctx context.Context, m *migration.Migrator) error {
	return m.Status(ctx)
}

func runOneMigration(ctx context.Context, m *migration.Migrator, identifier string) error {
	if err := m.Validate(ctx); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	log.Printf("Looking for migration: %s\n", identifier)
	if err := m.UpOne(ctx, identifier); err != nil {
		return err
	}

	version, err := m.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if version != nil {
		log.Printf("Database is now at migration version: %d\n", *version)
	}

	return nil
}
