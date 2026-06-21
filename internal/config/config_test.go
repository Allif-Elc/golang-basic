package config

import (
	"os"
	"testing"
	"time"
)

func TestGetEnv_Default(t *testing.T) {
	os.Unsetenv("TEST_VAR")
	got := getEnv("TEST_VAR", "default_val")
	if got != "default_val" {
		t.Errorf("expected default, got %q", got)
	}
}

func TestGetEnv_FromEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "from_env")
	defer os.Unsetenv("TEST_VAR")
	got := getEnv("TEST_VAR", "default")
	if got != "from_env" {
		t.Errorf("expected 'from_env', got %q", got)
	}
}

func TestGetEnv_EmptyEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "")
	defer os.Unsetenv("TEST_VAR")
	got := getEnv("TEST_VAR", "fallback")
	if got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}

func TestDatabaseConfig_Defaults(t *testing.T) {
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")

	cfg := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "golang_basic"),
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != "5432" {
		t.Errorf("Port = %q, want %q", cfg.Port, "5432")
	}
	if cfg.User != "postgres" {
		t.Errorf("User = %q, want %q", cfg.User, "postgres")
	}
	if cfg.DBName != "golang_basic" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "golang_basic")
	}
}

func TestConstants(t *testing.T) {
	if QueryTimeout != 5*time.Second {
		t.Error("QueryTimeout should be 5 seconds")
	}
	if SlowQueryThreshold != 50*time.Millisecond {
		t.Error("SlowQueryThreshold should be 50ms")
	}
}

func TestCloseDatabase_Nil(t *testing.T) {
	// DB is nil at package init, CloseDatabase should not panic
	CloseDatabase()
}
