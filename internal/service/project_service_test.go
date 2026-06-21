package service

import (
	"strings"
	"testing"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"null bytes", "hello\x00world", "helloworld"},
		{"control chars", "hello\x01\x02world", "helloworld"},
		{"whitespace trim", "  hello  ", "hello"},
		{"normal", "hello world", "hello world"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeInput(tt.in); got != tt.want {
				t.Errorf("sanitizeInput(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, n := range []string{"My Project", "API-V1_Test", "Test123", "a.b"} {
			if err := validateProjectName(n); err != nil {
				t.Errorf("unexpected error for %q: %v", n, err)
			}
		}
	})

	t.Run("too short", func(t *testing.T) {
		if err := validateProjectName("ab"); err == nil || !strings.Contains(err.Error(), "at least 3") {
			t.Errorf("expected 'at least 3' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 256)
		if err := validateProjectName(long); err == nil || !strings.Contains(err.Error(), "exceed 255") {
			t.Errorf("expected 'exceed 255' error, got %v", err)
		}
	})

	t.Run("XSS", func(t *testing.T) {
		if err := validateProjectName("<script>"); err == nil || !strings.Contains(err.Error(), "dangerous") {
			t.Errorf("expected dangerous content error, got %v", err)
		}
	})

	t.Run("SQL injection", func(t *testing.T) {
		if err := validateProjectName("'; DROP TABLE;--"); err == nil || !strings.Contains(err.Error(), "SQL") {
			t.Errorf("expected SQL error, got %v", err)
		}
	})

	t.Run("path traversal", func(t *testing.T) {
		if err := validateProjectName("../../../etc"); err == nil || !strings.Contains(err.Error(), "path") {
			t.Errorf("expected path error, got %v", err)
		}
	})

	t.Run("invalid chars", func(t *testing.T) {
		for _, n := range []string{"hello@world", "test#1", "foo/bar"} {
			if err := validateProjectName(n); err == nil {
				t.Errorf("expected error for invalid chars %q", n)
			}
		}
	})

	t.Run("consecutive dots", func(t *testing.T) {
		if err := validateProjectName("test..name"); err == nil || !strings.Contains(err.Error(), "consecutive dots") {
			t.Errorf("expected 'consecutive dots' error, got %v", err)
		}
	})
}

func TestValidateProjectDescription(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := validateProjectDescription("A normal description"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateProjectDescription(""); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 1001)
		if err := validateProjectDescription(long); err == nil || !strings.Contains(err.Error(), "exceed 1000") {
			t.Errorf("expected 'exceed 1000' error, got %v", err)
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := validateProjectDescription("bad\x00desc"); err == nil || !strings.Contains(err.Error(), "null") {
			t.Errorf("expected null byte error, got %v", err)
		}
	})

	t.Run("XSS", func(t *testing.T) {
		if err := validateProjectDescription("<script>alert(1)</script>"); err == nil {
			t.Errorf("expected XSS error, got nil")
		}
	})

	t.Run("SQL injection", func(t *testing.T) {
		if err := validateProjectDescription("' OR '1'='1"); err == nil {
			t.Errorf("expected SQL error, got nil")
		}
	})
}

func TestValidateProjectVersion(t *testing.T) {
	t.Run("valid semver", func(t *testing.T) {
		for _, v := range []string{"1.0.0", "v2.3.4", "1.0.0-beta", "2.0.0-alpha.1", "v1.2.3+build"} {
			if err := validateProjectVersion(v); err != nil {
				t.Errorf("expected no error for %q, got %v", v, err)
			}
		}
	})

	t.Run("valid non-semver", func(t *testing.T) {
		if err := validateProjectVersion("1.2"); err != nil {
			t.Errorf("expected no error for simple version, got %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateProjectVersion(""); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 51)
		if err := validateProjectVersion(long); err == nil || !strings.Contains(err.Error(), "exceed 50") {
			t.Errorf("expected 'exceed 50' error, got %v", err)
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := validateProjectVersion("1.0\x000"); err == nil || !strings.Contains(err.Error(), "null") {
			t.Errorf("expected null byte error, got %v", err)
		}
	})

	t.Run("invalid chars", func(t *testing.T) {
		if err := validateProjectVersion("!@#$"); err == nil {
			t.Errorf("expected error for invalid version")
		}
	})
}

func TestSanitizeSortField(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"name", "name"},
		{"created_at", "created_at"},
		{"updated_at", "updated_at"},
		{"invalid", "created_at"},
		{"DROP TABLE", "created_at"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := sanitizeSortField(tt.in); got != tt.want {
				t.Errorf("sanitizeSortField(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSanitizeSortOrder(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"asc", "asc"},
		{"desc", "desc"},
		{"ASC", "asc"},
		{"DESC", "desc"},
		{"invalid", "desc"},
		{"' OR '1'='1", "desc"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := sanitizeSortOrder(tt.in); got != tt.want {
				t.Errorf("sanitizeSortOrder(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestContainsXSSPatterns(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"<script>alert(1)</script>", true},
		{"javascript:alert(1)", true},
		{"onerror=alert(1)", true},
		{"onload=alert(1)", true},
		{"onclick=alert(1)", true},
		{"eval(something)", true},
		{"<img src=x>", true},
		{"<iframe>", true},
		{"vbscript:msgbox", true},
		{"data:text/html", true},
		{"expression(x)", true},
		{"normal text", false},
		{"hello world", false},
	}
	for _, tt := range tests {
		t.Run(tt.in[:min(len(tt.in), 20)], func(t *testing.T) {
			if got := containsXSSPatterns(tt.in); got != tt.want {
				t.Errorf("containsXSSPatterns(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestContainsSQLInjectionPatterns(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"' OR '1'='1", true},
		{"drop table users", true},
		{"delete from users", true},
		{"insert into users", true},
		{"union select", true},
		{"exec(sp_help)", true},
		{"xp_cmdshell", true},
		{"waitfor delay", true},
		{"benchmark(1000000)", true},
		{"';--", true},
		{"/* comment */", true},
		{"it''s a test", true},  // 2+ single quotes
		{"normal text", false},
		{"it's", false}, // 1 single quote is OK
	}
	for _, tt := range tests {
		t.Run(tt.in[:min(len(tt.in), 20)], func(t *testing.T) {
			if got := containsSQLInjectionPatterns(tt.in); got != tt.want {
				t.Errorf("containsSQLInjectionPatterns(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestContainsPathTraversalPatterns(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"../../../etc/passwd", true},
		{"..\\..\\windows\\system32", true},
		{"%2e%2e%2f", true},
		{"~/secret", true},
		{"/etc/passwd", true},
		{"/proc/self/fd", true},
		{"\\windows\\system32", true},
		{"normal/path.txt", false},
		{"just a name", false},
	}
	for _, tt := range tests {
		t.Run(tt.in[:min(len(tt.in), 20)], func(t *testing.T) {
			if got := containsPathTraversalPatterns(tt.in); got != tt.want {
				t.Errorf("containsPathTraversalPatterns(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
