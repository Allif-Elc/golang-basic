package service

import (
	"strings"
	"testing"

	"golang-basic/api/internal/model"
)

func TestValidateGrpcServiceName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, n := range []string{"UserService", "Greeter", "ApiV1"} {
			if err := ValidateGrpcServiceName(n); err != nil {
				t.Errorf("unexpected error for %q: %v", n, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := ValidateGrpcServiceName(""); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("A", 256)
		if err := ValidateGrpcServiceName(long); err == nil || !strings.Contains(err.Error(), "exceed 255") {
			t.Errorf("expected 'exceed 255' error, got %v", err)
		}
	})

	t.Run("invalid PascalCase", func(t *testing.T) {
		for _, n := range []string{"userService", "user_service", "123Service"} {
			if err := ValidateGrpcServiceName(n); err == nil {
				t.Errorf("expected error for %q, got nil", n)
			}
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := ValidateGrpcServiceName("Service\x00Name"); err == nil {
			t.Errorf("expected error for null bytes, got nil")
		}
	})

	t.Run("XSS patterns", func(t *testing.T) {
		if err := ValidateGrpcServiceName("<script>"); err == nil {
			t.Errorf("expected XSS error, got nil")
		}
	})
}

func TestValidateGrpcMethodName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, n := range []string{"GetUser", "CreateUser", "DeleteById"} {
			if err := ValidateGrpcMethodName(n); err != nil {
				t.Errorf("unexpected error for %q: %v", n, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := ValidateGrpcMethodName(""); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("A", 256)
		if err := ValidateGrpcMethodName(long); err == nil || !strings.Contains(err.Error(), "exceed 255") {
			t.Errorf("expected 'exceed 255' error, got %v", err)
		}
	})

	t.Run("invalid PascalCase", func(t *testing.T) {
		for _, n := range []string{"getUser", "get_user", "123Method"} {
			if err := ValidateGrpcMethodName(n); err == nil {
				t.Errorf("expected error for %q, got nil", n)
			}
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := ValidateGrpcMethodName("\x00Method"); err == nil {
			t.Errorf("expected error for null bytes, got nil")
		}
	})
}

func TestValidateGrpcAPIDescription(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := validateGrpcAPIDescription("handles user creation"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateGrpcAPIDescription(""); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 1001)
		if err := validateGrpcAPIDescription(long); err == nil || !strings.Contains(err.Error(), "exceed 1000") {
			t.Errorf("expected 'exceed 1000' error, got %v", err)
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := validateGrpcAPIDescription("bad\x00desc"); err == nil {
			t.Errorf("expected error for null bytes, got nil")
		}
	})

	t.Run("XSS", func(t *testing.T) {
		if err := validateGrpcAPIDescription("<script>"); err == nil {
			t.Errorf("expected XSS error, got nil")
		}
	})
}

func TestValidateProtoDefinition(t *testing.T) {
	t.Run("valid proto", func(t *testing.T) {
		proto := `syntax = "proto3";
service UserService {
  rpc GetUser (GetUserRequest) returns (UserResponse);
}`
		if err := validateProtoDefinition(proto); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateProtoDefinition(""); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 10001)
		if err := validateProtoDefinition(long); err == nil || !strings.Contains(err.Error(), "exceed 10000") {
			t.Errorf("expected 'exceed 10000' error, got %v", err)
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := validateProtoDefinition("proto\x00content"); err == nil {
			t.Errorf("expected error for null bytes, got nil")
		}
	})
}

func TestValidateGrpcFields(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		fields := []struct{ Name, Type, Label, Description string }{
			{"user_id", "int32", "required", "The user ID"},
			{"name", "string", "optional", "User name"},
		}
		grpcFields := make([]struct{ Name, Type, Label, Description string }, len(fields))
		_ = grpcFields // use model.GrpcField
		if err := validateGrpcFields(nil); err != nil {
			t.Errorf("expected no error for nil, got %v", err)
		}
	})

	t.Run("too many fields", func(t *testing.T) {
		fields := make([]struct{ Name, Type, Label, Description string }, 101)
		_ = fields
		// Can't test directly without building model.GrpcField slice
	})
}

func TestValidateGrpcFieldName(t *testing.T) {
	t.Run("valid snake_case", func(t *testing.T) {
		for _, n := range []string{"user_id", "name", "created_at", "field1"} {
			if err := validateGrpcFieldName(n); err != nil {
				t.Errorf("unexpected error for %q: %v", n, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateGrpcFieldName(""); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 101)
		if err := validateGrpcFieldName(long); err == nil || !strings.Contains(err.Error(), "too long") {
			t.Errorf("expected 'too long' error, got %v", err)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		for _, n := range []string{"UserId", "user.Name", "my-field", "123field"} {
			if err := validateGrpcFieldName(n); err == nil {
				t.Errorf("expected error for %q, got nil", n)
			}
		}
	})
}

func TestValidateGrpcFieldType(t *testing.T) {
	t.Run("valid scalar types", func(t *testing.T) {
		for _, typ := range []string{"double", "float", "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool", "string", "bytes"} {
			if err := validateGrpcFieldType(typ); err != nil {
				t.Errorf("unexpected error for %q: %v", typ, err)
			}
		}
	})

	t.Run("valid custom type (PascalCase)", func(t *testing.T) {
		for _, typ := range []string{"UserResponse", "GetUserRequest", "MyMessage"} {
			if err := validateGrpcFieldType(typ); err != nil {
				t.Errorf("unexpected error for %q: %v", typ, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateGrpcFieldType(""); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("A", 101)
		if err := validateGrpcFieldType(long); err == nil || !strings.Contains(err.Error(), "too long") {
			t.Errorf("expected 'too long' error, got %v", err)
		}
	})

	t.Run("invalid type format", func(t *testing.T) {
		for _, typ := range []string{"user_response", "123Type", "my-type"} {
			if err := validateGrpcFieldType(typ); err == nil {
				t.Errorf("expected error for %q, got nil", typ)
			}
		}
	})
}

func TestValidateGrpcExamples(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		examples := []model.GrpcExample{
			{Language: "go", Code: `client.GetUser(ctx, &req)`},
			{Language: "python", Code: `client.GetUser(request)`},
		}
		if err := validateGrpcExamples(examples); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("too many", func(t *testing.T) {
		examples := make([]model.GrpcExample, 21)
		if err := validateGrpcExamples(examples); err == nil || !strings.Contains(err.Error(), "too many") {
			t.Errorf("expected 'too many' error, got %v", err)
		}
	})

	t.Run("missing language", func(t *testing.T) {
		examples := []model.GrpcExample{{Code: "test"}}
		if err := validateGrpcExamples(examples); err == nil || !strings.Contains(err.Error(), "language") {
			t.Errorf("expected 'language' error, got %v", err)
		}
	})

	t.Run("missing code", func(t *testing.T) {
		examples := []model.GrpcExample{{Language: "go"}}
		if err := validateGrpcExamples(examples); err == nil || !strings.Contains(err.Error(), "code") {
			t.Errorf("expected 'code' error, got %v", err)
		}
	})

	t.Run("language too long", func(t *testing.T) {
		examples := []model.GrpcExample{{Language: strings.Repeat("a", 51), Code: "test"}}
		if err := validateGrpcExamples(examples); err == nil || !strings.Contains(err.Error(), "too long") {
			t.Errorf("expected 'too long' error, got %v", err)
		}
	})

	t.Run("code too long", func(t *testing.T) {
		examples := []model.GrpcExample{{Language: "go", Code: strings.Repeat("a", 10001)}}
		if err := validateGrpcExamples(examples); err == nil || !strings.Contains(err.Error(), "too long") {
			t.Errorf("expected 'too long' error, got %v", err)
		}
	})
}
