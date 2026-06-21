package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_SortsByVersion(t *testing.T) {
	migs := []Migration{
		{Version: 20260226003, Name: "third"},
		{Version: 20260226001, Name: "first"},
		{Version: 20260226002, Name: "second"},
	}
	m := New(nil, migs)

	if len(m.migrations) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(m.migrations))
	}
	if m.migrations[0].Name != "first" {
		t.Errorf("first should be 'first', got %q", m.migrations[0].Name)
	}
	if m.migrations[2].Name != "third" {
		t.Errorf("last should be 'third', got %q", m.migrations[2].Name)
	}
}

func TestNew_Empty(t *testing.T) {
	m := New(nil, []Migration{})
	if len(m.migrations) != 0 {
		t.Error("expected empty migrations")
	}
}

func TestNewFromFiles_ValidFiles(t *testing.T) {
	dir := t.TempDir()

	// Create valid migration files
	createFile(t, dir, "20260225_001_create_users.sql", "CREATE TABLE users (...)")
	createFile(t, dir, "20260225_001_create_users.down.sql", "DROP TABLE users")
	createFile(t, dir, "20260226_002_add_projects.sql", "CREATE TABLE projects (...)")
	createFile(t, dir, "20260226_002_add_projects.down.sql", "DROP TABLE projects")
	createFile(t, dir, "20260227_003_add_indexes.sql", "CREATE INDEX ...")

	m, err := NewFromFiles(nil, dir)
	if err != nil {
		t.Fatalf("NewFromFiles failed: %v", err)
	}

	if len(m.migrations) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(m.migrations))
	}

	// Check sorting
	if m.migrations[0].Version > m.migrations[1].Version {
		t.Error("migrations should be sorted by version")
	}

	// Check content
	if m.migrations[0].Name != "create_users" {
		t.Errorf("name = %q, want %q", m.migrations[0].Name, "create_users")
	}
	if m.migrations[0].Up != "CREATE TABLE users (...)" {
		t.Errorf("up content mismatch")
	}
	if m.migrations[0].Down != "DROP TABLE users" {
		t.Errorf("down content should be set")
	}
	if m.migrations[2].Down != "" {
		t.Error("migration without down file should have empty Down")
	}
}

func TestNewFromFiles_SkipsInvalidFiles(t *testing.T) {
	dir := t.TempDir()

	createFile(t, dir, "invalid.sql", "bad")
	createFile(t, dir, "readme.txt", "text")
	createFile(t, dir, "20260225_001_valid.sql", "valid")

	m, err := NewFromFiles(nil, dir)
	if err != nil {
		t.Fatalf("NewFromFiles failed: %v", err)
	}

	if len(m.migrations) != 1 {
		t.Fatalf("expected 1 valid migration, got %d", len(m.migrations))
	}
	if m.migrations[0].Name != "valid" {
		t.Errorf("name = %q, want %q", m.migrations[0].Name, "valid")
	}
}

func TestNewFromFiles_NoDownFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "20260225_001_only_up.sql", "UP SQL")

	m, err := NewFromFiles(nil, dir)
	if err != nil {
		t.Fatalf("NewFromFiles failed: %v", err)
	}

	if len(m.migrations) != 1 {
		t.Fatalf("expected 1 migration, got %d", len(m.migrations))
	}
	if m.migrations[0].Down != "" {
		t.Error("Down should be empty when no .down.sql file exists")
	}
}

func TestNewFromFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	m, err := NewFromFiles(nil, dir)
	if err != nil {
		t.Fatalf("NewFromFiles failed: %v", err)
	}

	if len(m.migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(m.migrations))
	}
}

func TestNewFromFiles_SubDir(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	os.Mkdir(subDir, 0755)
	createFile(t, subDir, "20260225_001_inside_sub.sql", "UP")
	createFile(t, dir, "20260225_002_root.sql", "UP")

	m, err := NewFromFiles(nil, dir)
	if err != nil {
		t.Fatalf("NewFromFiles failed: %v", err)
	}

	// Should only pick up root files, not subdirectory files
	if len(m.migrations) != 1 {
		t.Fatalf("expected 1 migration (subdir ignored), got %d", len(m.migrations))
	}
	if m.migrations[0].Name != "root" {
		t.Errorf("name = %q, want %q", m.migrations[0].Name, "root")
	}
}

func TestNewFromFiles_NonExistentDir(t *testing.T) {
	_, err := NewFromFiles(nil, "/nonexistent/dir/migrations")
	if err == nil {
		t.Fatal("expected error for nonexistent dir")
	}
}

func TestMigrationSortCopy(t *testing.T) {
	original := []Migration{
		{Version: 2},
		{Version: 1},
	}
	m := New(nil, original)

	// Verify original isn't modified
	if original[0].Version != 2 {
		t.Error("New should not modify the original slice")
	}
	// Verify sorted copy
	if m.migrations[0].Version != 1 {
		t.Error("migrations should be sorted ascending")
	}
}

func createFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file %s: %v", name, err)
	}
}
