package repository

import (
	"testing"
)

func TestParsePolicyRule_Valid(t *testing.T) {
	data := []byte(`{"role": "admin", "resource": "employee_records", "action": ["read", "write"]}`)
	rule, err := ParsePolicyRule(data)
	if err != nil {
		t.Fatalf("ParsePolicyRule failed: %v", err)
	}
	if rule.Role != "admin" {
		t.Errorf("Role = %q, want %q", rule.Role, "admin")
	}
	if rule.Resource != "employee_records" {
		t.Errorf("Resource = %q, want %q", rule.Resource, "employee_records")
	}
	if len(rule.Action) != 2 || rule.Action[0] != "read" {
		t.Errorf("Action = %v, want [read write]", rule.Action)
	}
}

func TestParsePolicyRule_AttributeBased(t *testing.T) {
	data := []byte(`{"attribute_name": "department", "attribute_value": "engineering", "resource": "repos", "action": ["*"]}`)
	rule, err := ParsePolicyRule(data)
	if err != nil {
		t.Fatalf("ParsePolicyRule failed: %v", err)
	}
	if rule.AttributeName != "department" {
		t.Errorf("AttributeName = %q, want %q", rule.AttributeName, "department")
	}
	if rule.AttributeValue != "engineering" {
		t.Errorf("AttributeValue = %q, want %q", rule.AttributeValue, "engineering")
	}
	if rule.Resource != "repos" {
		t.Errorf("Resource = %q, want %q", rule.Resource, "repos")
	}
	if len(rule.Action) != 1 || rule.Action[0] != "*" {
		t.Errorf("Action = %v, want [*]", rule.Action)
	}
}

func TestParsePolicyRule_InvalidJSON(t *testing.T) {
	_, err := ParsePolicyRule([]byte(`{invalid}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParsePolicyRule_Empty(t *testing.T) {
	_, err := ParsePolicyRule([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestParsePolicyRule_Partial(t *testing.T) {
	data := []byte(`{"resource": "dashboard"}`)
	rule, err := ParsePolicyRule(data)
	if err != nil {
		t.Fatalf("ParsePolicyRule failed: %v", err)
	}
	if rule.Resource != "dashboard" {
		t.Errorf("Resource = %q, want %q", rule.Resource, "dashboard")
	}
	if rule.Role != "" {
		t.Error("Role should be empty")
	}
}

func TestParsePolicyRule_WildcardAction(t *testing.T) {
	data := []byte(`{"role": "admin", "resource": "*", "action": ["*"]}`)
	rule, err := ParsePolicyRule(data)
	if err != nil {
		t.Fatalf("ParsePolicyRule failed: %v", err)
	}
	if rule.Resource != "*" {
		t.Errorf("Resource = %q, want %q", rule.Resource, "*")
	}
	if len(rule.Action) != 1 || rule.Action[0] != "*" {
		t.Errorf("Action = %v, want [*]", rule.Action)
	}
}

func TestParsePolicyRule_EmptyAction(t *testing.T) {
	data := []byte(`{"role": "viewer", "resource": "analytics", "action": []}`)
	_, err := ParsePolicyRule(data)
	if err != nil {
		t.Fatalf("ParsePolicyRule failed: %v", err)
	}
}
