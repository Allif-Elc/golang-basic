package service_test

import (
	"strings"
	"golang-basic/api/internal/service"
	"testing"
)

// TestHashVerifyPassword_RoundTrip tests that a hashed password can be verified
// This test exposes the bug in verifyPassword
func TestHashVerifyPassword_RoundTrip(t *testing.T) {
	password := "testPassword123"

	// Hash the password using the actual function
	hashed, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	t.Logf("Generated hash: %s", hashed)

	// Try to verify it using the actual function
	valid, err := service.VerifyPassword(password, hashed)
	if err != nil {
		t.Logf("VerifyPassword error: %v", err)
		t.Logf("This indicates a bug - verifyPassword cannot parse the hash it created!")
	}

	if !valid {
		t.Error("Password verification failed - the hash and verify functions are incompatible")
	}

	t.Logf("Valid: %v, Error: %v", valid, err)
}

// TestHashVerifyPassword_CorrectPassword tests correct password verification
func TestHashVerifyPassword_CorrectPassword(t *testing.T) {
	password := "MySecurePassword123!"

	hashed, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	valid, err := service.VerifyPassword(password, hashed)
	if err != nil {
		t.Errorf("VerifyPassword failed with error: %v (this indicates a parsing bug)", err)
	}
	if !valid {
		t.Error("Expected valid password to return true")
	}
}

// TestHashVerifyPassword_WrongPassword tests that wrong password is rejected
func TestHashVerifyPassword_WrongPassword(t *testing.T) {
	password := "MySecurePassword123!"
	wrongPassword := "WrongPassword456!"

	hashed, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	valid, err := service.VerifyPassword(wrongPassword, hashed)
	if err != nil {
		// If there's a parsing error, we can't even test wrong password
		t.Skipf("Skipping - verifyPassword has a parsing bug: %v", err)
	}
	if valid {
		t.Error("Expected wrong password to return false")
	}
}

// TestHashVerifyPassword_EmptyPassword tests empty password edge case
func TestHashVerifyPassword_EmptyPassword(t *testing.T) {
	password := ""

	hashed, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	valid, err := service.VerifyPassword(password, hashed)
	if err != nil {
		t.Errorf("VerifyPassword failed with error: %v", err)
	}
	if !valid {
		t.Error("Expected empty password to match its hash")
	}
}

// TestHashVerifyPassword_FormatTests tests the hash format
func TestHashVerifyPassword_FormatTests(t *testing.T) {
	password := "Test123!"

	hashed, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	t.Logf("Hash format: %s", hashed)
	t.Logf("Hash length: %d", len(hashed))

	// Expected format: $argon2id$v=<version>$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
	// Example: $argon2id$v=19$m=65536,t=3,p=2$<base64salt>$<base64hash>

	// Check if it starts with the correct prefix
	expectedPrefix := "$argon2id$v="
	if len(hashed) < len(expectedPrefix) || hashed[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Hash doesn't start with expected prefix %s, got: %s", expectedPrefix, hashed)
	}

	// Count the dollar signs - should be 5
	dollarCount := 0
	for _, c := range hashed {
		if c == '$' {
			dollarCount++
		}
	}
	if dollarCount != 5 {
		t.Errorf("Expected 5 dollar signs in hash, got %d. Hash: %s", dollarCount, hashed)
	}
}

// TestHashVerifyPassword_InvalidFormat tests invalid hash formats
func TestHashVerifyPassword_InvalidFormat(t *testing.T) {
	password := "testpassword"

	tests := []struct {
		name        string
		invalidHash string
	}{
		{
			name:        "empty hash",
			invalidHash: "",
		},
		{
			name:        "missing dollar signs",
			invalidHash: "argon2idv19m65536t3p2salt",
		},
		{
			name:        "too few parts",
			invalidHash: "$argon2id$v=19$m=65536,t=3,p=2$salt",
		},
		{
			name:        "invalid version format",
			invalidHash: "$argon2id$invalid$m=65536,t=3,p=2$salt$hash",
		},
		{
			name:        "invalid params format",
			invalidHash: "$argon2id$v=19$invalidparams$salt$hash",
		},
		{
			name:        "invalid base64 salt",
			invalidHash: "$argon2id$v=19$m=65536,t=3,p=2$!@#$%^&*()$hash",
		},
		{
			name:        "invalid base64 hash",
			invalidHash: "$argon2id$v=19$m=65536,t=3,p=2$salt$!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := service.VerifyPassword(password, tt.invalidHash)
			if err == nil {
				t.Error("Expected error for invalid hash format, got nil")
			}
			if valid {
				t.Error("Expected false for invalid hash format")
			}
		})
	}
}

// TestHashVerifyPassword_CaseSensitivity tests that password verification is case-sensitive
func TestHashVerifyPassword_CaseSensitivity(t *testing.T) {
	password := "MyPassword123"
	wrongCase := "mypassword123"

	hashed, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	valid, err := service.VerifyPassword(wrongCase, hashed)
	if err != nil {
		t.Errorf("VerifyPassword failed with error: %v", err)
	}
	if valid {
		t.Error("Password verification should be case-sensitive")
	}
}

// TestHashVerifyPassword_SpecialCharacters tests passwords with special characters
func TestHashVerifyPassword_SpecialCharacters(t *testing.T) {
	passwords := []string{
		"P@ssw0rd!",
		"test#123$ABC",
		"sp3cial!@#$%",
		"µñíçødé©hårß",
		"password\twith\nnewlines",
	}

	for _, pwd := range passwords {
		t.Run("password_"+strings.ReplaceAll(strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return '_'
		}, pwd), " ", "_"), func(t *testing.T) {
			hashed, err := service.HashPassword(pwd)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}

			valid, err := service.VerifyPassword(pwd, hashed)
			if err != nil {
				t.Errorf("VerifyPassword failed: %v", err)
			}
			if !valid {
				t.Error("Password verification failed for special character password")
			}
		})
	}
}

// TestHashVerifyPassword_LongPassword tests very long passwords
func TestHashVerifyPassword_LongPassword(t *testing.T) {
	password := strings.Repeat("a", 1000)

	hashed, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	valid, err := service.VerifyPassword(password, hashed)
	if err != nil {
		t.Errorf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("Long password verification failed")
	}
}

// TestHashVerifyPassword_SamePasswordDifferentHash tests that same password produces different hash each time
func TestHashVerifyPassword_SamePasswordDifferentHash(t *testing.T) {
	password := "testPassword123"

	hash1, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	// Hashes should be different due to random salt
	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to random salt")
	}

	// But both should verify successfully
	valid1, err1 := service.VerifyPassword(password, hash1)
	valid2, err2 := service.VerifyPassword(password, hash2)

	if err1 != nil || !valid1 {
		t.Errorf("First hash verification failed: valid=%v, err=%v", valid1, err1)
	}
	if err2 != nil || !valid2 {
		t.Errorf("Second hash verification failed: valid=%v, err=%v", valid2, err2)
	}
}

// TestHashVerifyPassword_Unicode tests passwords with unicode characters
func TestHashVerifyPassword_Unicode(t *testing.T) {
	passwords := []string{
		"пароль123",     // Russian
		"密码123",        // Chinese
		"motdepasse123",  // French
		"κωδικός123",     // Greek
		"😀🔑password",   // Emoji
	}

	for i, pwd := range passwords {
		t.Run("unicode_"+string(rune('a'+i)), func(t *testing.T) {
			hashed, err := service.HashPassword(pwd)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}

			valid, err := service.VerifyPassword(pwd, hashed)
			if err != nil {
				t.Errorf("VerifyPassword failed: %v", err)
			}
			if !valid {
				t.Error("Unicode password verification failed")
			}
		})
	}
}
