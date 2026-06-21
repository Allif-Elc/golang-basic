package service

import (
	"testing"
)

func TestMatchesResource(t *testing.T) {
	tests := []struct {
		policy   string
		request  string
		want     bool
	}{
		{"*", "anything", true},
		{"project", "project", true},
		{"project", "other", false},
		{"project_*", "project_view", true},
		{"project_*", "project_", true},
		{"project_*", "project", false},
		{"project_*", "other_view", false},
		{"a_*", "a_b_c", true},
		{"a_*", "ab_c", false},
	}
	for _, tt := range tests {
		t.Run(tt.policy+"/"+tt.request, func(t *testing.T) {
			if got := matchesResource(tt.policy, tt.request); got != tt.want {
				t.Errorf("matchesResource(%q, %q) = %v, want %v", tt.policy, tt.request, got, tt.want)
			}
		})
	}
}

func TestMatchesAction(t *testing.T) {
	tests := []struct {
		actions []string
		request string
		want    bool
	}{
		{[]string{"*"}, "anything", true},
		{[]string{"read"}, "read", true},
		{[]string{"read"}, "write", false},
		{[]string{"read", "write"}, "write", true},
		{[]string{"read", "write"}, "delete", false},
	}
	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			if got := matchesAction(tt.actions, tt.request); got != tt.want {
				t.Errorf("matchesAction(%v, %q) = %v, want %v", tt.actions, tt.request, got, tt.want)
			}
		})
	}
}

func TestParsePolicyRule(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		rule, err := parsePolicyRule([]byte(`{"resource":"*","action":["read"]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.Resource != "*" {
			t.Errorf("expected resource '*', got %q", rule.Resource)
		}
		if len(rule.Action) != 1 || rule.Action[0] != "read" {
			t.Errorf("expected action [read], got %v", rule.Action)
		}
	})

	t.Run("valid with attribute", func(t *testing.T) {
		rule, err := parsePolicyRule([]byte(`{"attribute_name":"role","attribute_value":"admin","resource":"*","action":["*"]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.AttributeName != "role" {
			t.Errorf("expected attribute_name 'role', got %q", rule.AttributeName)
		}
		if rule.AttributeValue != "admin" {
			t.Errorf("expected attribute_value 'admin', got %q", rule.AttributeValue)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parsePolicyRule([]byte(`{invalid`))
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("empty object", func(t *testing.T) {
		rule, err := parsePolicyRule([]byte(`{}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.Resource != "" {
			t.Errorf("expected empty resource, got %q", rule.Resource)
		}
	})
}
