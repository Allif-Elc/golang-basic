package service

import (
	"errors"
	"strings"
	"testing"
)

func TestIsValidationError(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		err := ValidationError{Field: "name", Message: "too short"}
		if !IsValidationError(err) {
			t.Error("expected true for ValidationError")
		}
	})

	t.Run("regular error", func(t *testing.T) {
		err := errors.New("regular error")
		if IsValidationError(err) {
			t.Error("expected false for regular error")
		}
	})

	t.Run("nil", func(t *testing.T) {
		if IsValidationError(nil) {
			t.Error("expected false for nil")
		}
	})
}

func TestAggregateErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		if err := AggregateErrors(nil); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		if err := AggregateErrors([]error{}); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("single error", func(t *testing.T) {
		err := AggregateErrors([]error{errors.New("name: too short")})
		if err == nil || err.Error() != "name: too short" {
			t.Errorf("expected 'name: too short', got %v", err)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		err := AggregateErrors([]error{
			errors.New("name: too short"),
			errors.New("type: invalid"),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "name: too short") {
			t.Errorf("expected to contain 'name: too short', got %s", err.Error())
		}
		if !strings.Contains(err.Error(), "type: invalid") {
			t.Errorf("expected to contain 'type: invalid', got %s", err.Error())
		}
	})
}

func TestPermissionService_validateAttributeType(t *testing.T) {
	svc := &PermissionService{}
	t.Run("valid types", func(t *testing.T) {
		for _, typ := range []string{"string", "number", "boolean", "enum"} {
			if err := svc.validateAttributeType(typ); err != nil {
				t.Errorf("expected no error for %q, got %v", typ, err)
			}
		}
	})
	t.Run("invalid types", func(t *testing.T) {
		for _, typ := range []string{"int", "float", "array", "object", ""} {
			if err := svc.validateAttributeType(typ); err == nil {
				t.Errorf("expected error for %q, got nil", typ)
			}
		}
	})
}

func TestPermissionService_validateEnumValues(t *testing.T) {
	svc := &PermissionService{}
	t.Run("enum with values", func(t *testing.T) {
		if err := svc.validateEnumValues("enum", []string{"a", "b"}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("enum without values", func(t *testing.T) {
		if err := svc.validateEnumValues("enum", nil); err == nil {
			t.Error("expected error for enum without values, got nil")
		}
		if err := svc.validateEnumValues("enum", []string{}); err == nil {
			t.Error("expected error for empty enum values, got nil")
		}
	})
	t.Run("non-enum type", func(t *testing.T) {
		if err := svc.validateEnumValues("string", nil); err != nil {
			t.Errorf("expected no error for non-enum, got %v", err)
		}
	})
}

func TestPermissionService_validateEffect(t *testing.T) {
	svc := &PermissionService{}
	t.Run("allow", func(t *testing.T) {
		if err := svc.validateEffect("allow"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("deny", func(t *testing.T) {
		if err := svc.validateEffect("deny"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		for _, e := range []string{"", "grant", "revoke"} {
			if err := svc.validateEffect(e); err == nil {
				t.Errorf("expected error for %q, got nil", e)
			}
		}
	})
}

func TestPermissionService_validateActions(t *testing.T) {
	svc := &PermissionService{}
	t.Run("valid", func(t *testing.T) {
		if err := svc.validateActions([]string{"read"}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := svc.validateActions([]string{"read", "write"}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if err := svc.validateActions(nil); err == nil {
			t.Error("expected error for nil, got nil")
		}
		if err := svc.validateActions([]string{}); err == nil {
			t.Error("expected error for empty, got nil")
		}
	})
}

func TestPermissionService_validateResourceType(t *testing.T) {
	svc := &PermissionService{}
	t.Run("empty", func(t *testing.T) {
		if err := svc.validateResourceType(""); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})
	t.Run("valid", func(t *testing.T) {
		for _, rt := range []string{"general", "api", "document", "data"} {
			if err := svc.validateResourceType(rt); err != nil {
				t.Errorf("expected no error for %q, got %v", rt, err)
			}
		}
	})
	t.Run("invalid", func(t *testing.T) {
		if err := svc.validateResourceType("invalid"); err == nil {
			t.Error("expected error for invalid, got nil")
		}
	})
}

func TestPermissionService_validateName(t *testing.T) {
	svc := &PermissionService{}
	t.Run("valid", func(t *testing.T) {
		if err := svc.validateName("read", "name"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := svc.validateName("custom-attr", "name"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("too short", func(t *testing.T) {
		if err := svc.validateName("", "name"); err == nil {
			t.Error("expected error for empty, got nil")
		}
	})
	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 101)
		if err := svc.validateName(long, "name"); err == nil {
			t.Error("expected error for too long, got nil")
		}
	})
	t.Run("invalid chars", func(t *testing.T) {
		if err := svc.validateName("hello world", "name"); err == nil {
			t.Error("expected error for space, got nil")
		}
		if err := svc.validateName("test@name", "name"); err == nil {
			t.Error("expected error for @, got nil")
		}
	})
	t.Run("SQL-like", func(t *testing.T) {
		if err := svc.validateName("DROP", "name"); err == nil {
			t.Error("expected error for 'DROP', got nil")
		}
		if err := svc.validateName("SELECT", "name"); err == nil {
			t.Error("expected error for 'SELECT', got nil")
		}
	})
}

func TestPermissionService_validateDescription(t *testing.T) {
	svc := &PermissionService{}
	t.Run("empty", func(t *testing.T) {
		if err := svc.validateDescription("", "desc"); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})
	t.Run("valid", func(t *testing.T) {
		if err := svc.validateDescription("Read access", "desc"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 1001)
		if err := svc.validateDescription(long, "desc"); err == nil {
			t.Error("expected error for too long, got nil")
		}
	})
	t.Run("invalid chars", func(t *testing.T) {
		if err := svc.validateDescription("<script>", "desc"); err == nil {
			t.Error("expected error for HTML tags, got nil")
		}
	})
}

func TestPermissionService_validateID(t *testing.T) {
	svc := &PermissionService{}
	t.Run("valid", func(t *testing.T) {
		if err := svc.validateID(1, "id"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("zero", func(t *testing.T) {
		if err := svc.validateID(0, "id"); err == nil {
			t.Error("expected error for 0, got nil")
		}
	})
	t.Run("negative", func(t *testing.T) {
		if err := svc.validateID(-1, "id"); err == nil {
			t.Error("expected error for negative, got nil")
		}
	})
}
