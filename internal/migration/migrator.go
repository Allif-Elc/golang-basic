package migration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration represents a single database migration
type Migration struct {
	Version int64
	Name    string
	Up      string
	Down    string
}

// Migrator handles database migrations
type Migrator struct {
	db        *pgxpool.Pool
	migrations []Migration
}

// New creates a new migrator instance
func New(db *pgxpool.Pool, migrations []Migration) *Migrator {
	// Sort migrations by version to ensure correct execution order
	sortedMigrations := make([]Migration, len(migrations))
	copy(sortedMigrations, migrations)
	sort.Slice(sortedMigrations, func(i, j int) bool {
		return sortedMigrations[i].Version < sortedMigrations[j].Version
	})

	return &Migrator{
		db:        db,
		migrations: sortedMigrations,
	}
}

// NewFromFiles creates a new migrator instance by loading SQL files from a directory
// File naming convention: YYYYMMDD_XXX_description.sql
// Down migrations: YYYYMMDD_XXX_description.down.sql (optional)
func NewFromFiles(db *pgxpool.Pool, sqlDir string) (*Migrator, error) {
	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sql directory: %w", err)
	}

	// Regex to parse migration files: YYYYMMDD_XXX_description.sql
	// Matches: 20260225_001_abac_schema.sql
	migrationRegex := regexp.MustCompile(`^(\d{8})_(\d{3})_([a-z0-9_]+)\.sql$`)

	type fileMigration struct {
		version int64
		name    string
		upFile  string
		downFile string
	}

	var fileMigrations []fileMigration

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Skip .down.sql files, they're paired separately
		if filepath.Ext(filename) == ".sql" && strings.HasSuffix(filename, ".down.sql") {
			continue
		}

		// Skip non-sql files
		if filepath.Ext(filename) != ".sql" {
			continue
		}

		matches := migrationRegex.FindStringSubmatch(filename)
		if matches == nil {
			log.Printf("Warning: skipping file with invalid naming format: %s", filename)
			continue
		}

		dateStr := matches[1]       // YYYYMMDD
		seqStr := matches[2]         // XXX
		name := matches[3]           // description

		// Parse version: YYYYMMDDXXX (e.g., 20260225001)
		dateVal, err := strconv.ParseInt(dateStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid date in filename %s: %w", filename, err)
		}

		seqVal, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid sequence in filename %s: %w", filename, err)
		}

		version := dateVal*1000 + seqVal

		// Check for corresponding down migration file
		downFile := filepath.Join(sqlDir, strings.TrimSuffix(filename, ".sql")+".down.sql")
		downExists := true
		if _, err := os.Stat(downFile); os.IsNotExist(err) {
			downExists = false
		} else if err != nil {
			return nil, fmt.Errorf("failed to check down file %s: %w", downFile, err)
		}

		fm := fileMigration{
			version:  version,
			name:     name,
			upFile:   filepath.Join(sqlDir, filename),
			downFile: "",
		}

		if downExists {
			fm.downFile = downFile
		}

		fileMigrations = append(fileMigrations, fm)
	}

	// Sort by version
	sort.Slice(fileMigrations, func(i, j int) bool {
		return fileMigrations[i].version < fileMigrations[j].version
	})

	// Load migration SQL files
	var migrations []Migration
	for _, fm := range fileMigrations {
		upSQL, err := os.ReadFile(fm.upFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read up migration file %s: %w", fm.upFile, err)
		}

		var downSQL string
		if fm.downFile != "" {
			downBytes, err := os.ReadFile(fm.downFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read down migration file %s: %w", fm.downFile, err)
			}
			downSQL = string(downBytes)
		}

		migrations = append(migrations, Migration{
			Version: fm.version,
			Name:    fm.name,
			Up:      string(upSQL),
			Down:    downSQL,
		})
	}

	if len(migrations) == 0 {
		log.Println("Warning: no migration files found in", sqlDir)
	}

	return &Migrator{
		db:        db,
		migrations: migrations,
	}, nil
}

// InitSchema creates the schema_migrations table if it doesn't exist
func (m *Migrator) InitSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	_, err := m.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	return nil
}

// GetAppliedVersions returns a map of applied migration versions
func (m *Migrator) GetAppliedVersions(ctx context.Context) (map[int64]bool, error) {
	query := `SELECT version FROM schema_migrations ORDER BY version;`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating applied migrations: %w", err)
	}

	return applied, nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.InitSchema(ctx); err != nil {
		return err
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return err
	}

	pending := make([]Migration, 0)
	for _, mig := range m.migrations {
		if !applied[mig.Version] {
			pending = append(pending, mig)
		}
	}

	if len(pending) == 0 {
		log.Println("No pending migrations to run")
		return nil
	}

	log.Printf("Running %d pending migration(s)\n", len(pending))
	for _, mig := range pending {
		if err := m.runMigration(ctx, mig, mig.Up); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", mig.Version, mig.Name, err)
		}
		log.Printf("  ✓ Applied migration %d: %s\n", mig.Version, mig.Name)
	}

	return nil
}

// Down rolls back the most recent migration that has a down migration
func (m *Migrator) Down(ctx context.Context) error {
	if err := m.InitSchema(ctx); err != nil {
		return err
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return err
	}

	// Find the latest applied migration that has a down migration
	var latest *Migration
	for i := len(m.migrations) - 1; i >= 0; i-- {
		mig := m.migrations[i]
		if applied[mig.Version] && mig.Down != "" {
			latest = &mig
			break
		}
	}

	if latest == nil {
		log.Println("No migrations to roll back")
		return nil
	}

	log.Printf("Rolling back migration %d: %s\n", latest.Version, latest.Name)
	if err := m.runMigration(ctx, *latest, latest.Down); err != nil {
		return fmt.Errorf("rollback of migration %d (%s) failed: %w", latest.Version, latest.Name, err)
	}

	// Remove the migration record from schema_migrations
	deleteQuery := `DELETE FROM schema_migrations WHERE version = $1;`
	if _, err := m.db.Exec(ctx, deleteQuery, latest.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	log.Printf("  ✓ Rolled back migration %d: %s\n", latest.Version, latest.Name)
	return nil
}

// Status shows the current migration status
func (m *Migrator) Status(ctx context.Context) error {
	if err := m.InitSchema(ctx); err != nil {
		return err
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return err
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	for _, mig := range m.migrations {
		status := "  PENDING"
		if applied[mig.Version] {
			status = "  APPLIED"
		}
		fmt.Printf("%s  %d  %s\n", status, mig.Version, mig.Name)
	}
	fmt.Println("================")
	return nil
}

// runMigration executes a single migration within a transaction
func (m *Migrator) runMigration(ctx context.Context, mig Migration, sql string) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, sql); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration
	insertQuery := `
		INSERT INTO schema_migrations (version, name)
		VALUES ($1, $2)
		ON CONFLICT (version) DO NOTHING;
	`
	if _, err := tx.Exec(ctx, insertQuery, mig.Version, mig.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// GetVersion returns the current schema version (latest applied migration)
func (m *Migrator) GetVersion(ctx context.Context) (*int64, error) {
	if err := m.InitSchema(ctx); err != nil {
		return nil, err
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	if len(applied) == 0 {
		return nil, nil
	}

	// Find the latest applied migration
	var latest int64
	for _, mig := range m.migrations {
		if applied[mig.Version] {
			latest = mig.Version
		}
	}

	return &latest, nil
}

// findByVersion locates a migration by its version number
func (m *Migrator) findByVersion(version int64) (*Migration, error) {
	for _, mig := range m.migrations {
		if mig.Version == version {
			return &mig, nil
		}
	}
	return nil, fmt.Errorf("migration with version %d not found", version)
}

// FindMigration parses an identifier and returns the matching migration
// Parsing priority:
// 1. Full filename (YYYYMMDD_XXX_name.sql) → Extract version via regex
// 2. Pure number → Parse as version
// 3. Otherwise → Treat as migration name (exact match)
func (m *Migrator) FindMigration(identifier string) (*Migration, error) {
	if identifier == "" {
		return nil, fmt.Errorf("migration identifier cannot be empty")
	}

	// Try parsing as full filename first
	migrationRegex := regexp.MustCompile(`^(\d{8})_(\d{3})_([a-z0-9_]+)\.sql$`)
	if matches := migrationRegex.FindStringSubmatch(identifier); matches != nil {
		dateStr := matches[1]
		seqStr := matches[2]

		dateVal, err := strconv.ParseInt(dateStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid date in filename: %w", err)
		}

		seqVal, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid sequence in filename: %w", err)
		}

		version := dateVal*1000 + seqVal
		return m.findByVersion(version)
	}

	// Try parsing as pure version number
	if version, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return m.findByVersion(version)
	}

	// Try exact name match
	for _, mig := range m.migrations {
		if mig.Name == identifier {
			return &mig, nil
		}
	}

	return nil, fmt.Errorf("migration not found: %s", identifier)
}

// UpOne runs a single migration by identifier (name, version, or filename)
func (m *Migrator) UpOne(ctx context.Context, identifier string) error {
	if err := m.InitSchema(ctx); err != nil {
		return err
	}

	mig, err := m.FindMigration(identifier)
	if err != nil {
		return err
	}

	// Check if already applied
	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if applied[mig.Version] {
		return fmt.Errorf("migration %d (%s) is already applied", mig.Version, mig.Name)
	}

	log.Printf("Applying migration %d: %s\n", mig.Version, mig.Name)
	if err := m.runMigration(ctx, *mig, mig.Up); err != nil {
		return fmt.Errorf("migration %d (%s) failed: %w", mig.Version, mig.Name, err)
	}
	log.Printf("  ✓ Applied migration %d: %s\n", mig.Version, mig.Name)

	return nil
}

// Validate checks for common migration issues
func (m *Migrator) Validate(ctx context.Context) error {
	versions := make(map[int64]bool)
	for _, mig := range m.migrations {
		if versions[mig.Version] {
			return fmt.Errorf("duplicate migration version: %d", mig.Version)
		}
		versions[mig.Version] = true

		if mig.Name == "" {
			return fmt.Errorf("migration %d has empty name", mig.Version)
		}

		if mig.Up == "" {
			return fmt.Errorf("migration %d (%s) has empty Up SQL", mig.Version, mig.Name)
		}
	}
	return nil
}
