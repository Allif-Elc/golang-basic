package service

import (
	"strings"
	"testing"

	"golang-basic/api/internal/model"
)

func TestValidateGraphQLAPIName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, n := range []string{"GetUser", "_field", "QueryUsers", "mutation1"} {
			if err := ValidateGraphQLAPIName(n); err != nil {
				t.Errorf("unexpected error for %q: %v", n, err)
			}
		}
	})

	t.Run("too short", func(t *testing.T) {
		if err := ValidateGraphQLAPIName("ab"); err == nil || !strings.Contains(err.Error(), "at least 3") {
			t.Errorf("expected 'at least 3' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 256)
		if err := ValidateGraphQLAPIName(long); err == nil || !strings.Contains(err.Error(), "exceed 255") {
			t.Errorf("expected 'exceed 255' error, got %v", err)
		}
	})

	t.Run("invalid chars", func(t *testing.T) {
		for _, n := range []string{"hello world", "test-name", "123abc"} {
			if err := ValidateGraphQLAPIName(n); err == nil {
				t.Errorf("expected error for %q", n)
			}
		}
	})

	t.Run("XSS", func(t *testing.T) {
		if err := ValidateGraphQLAPIName("<script>"); err == nil || !strings.Contains(err.Error(), "dangerous") {
			t.Errorf("expected dangerous content error, got %v", err)
		}
	})

	t.Run("SQL injection", func(t *testing.T) {
		if err := ValidateGraphQLAPIName("' OR '1'='1"); err == nil || !strings.Contains(err.Error(), "SQL") {
			t.Errorf("expected SQL error, got %v", err)
		}
	})
}

func TestValidateGraphQLAPIDescription(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := validateGraphQLAPIDescription("A description"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateGraphQLAPIDescription(""); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 1001)
		if err := validateGraphQLAPIDescription(long); err == nil || !strings.Contains(err.Error(), "exceed 1000") {
			t.Errorf("expected 'exceed 1000' error, got %v", err)
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := validateGraphQLAPIDescription("bad\x00desc"); err == nil || !strings.Contains(err.Error(), "null") {
			t.Errorf("expected null byte error, got %v", err)
		}
	})

	t.Run("XSS", func(t *testing.T) {
		if err := validateGraphQLAPIDescription("<script>alert(1)</script>"); err == nil {
			t.Errorf("expected XSS error, got nil")
		}
	})
}

func TestValidateGraphQLOperationType(t *testing.T) {
	for _, v := range []string{"query", "mutation", "subscription"} {
		if err := ValidateGraphQLOperationType(v); err != nil {
			t.Errorf("unexpected error for %q: %v", v, err)
		}
	}

	for _, v := range []string{"QUERY", "Mutation", "invalid", "rest"} {
		if err := ValidateGraphQLOperationType(v); err == nil {
			t.Errorf("expected error for %q", v)
		}
	}
}

func TestValidateGraphQLReturnType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, rt := range []string{"String", "Int!", "CustomType"} {
			if err := validateGraphQLReturnType(rt); err != nil {
				t.Errorf("unexpected error for %q: %v", rt, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateGraphQLReturnType(""); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 501)
		if err := validateGraphQLReturnType(long); err == nil || !strings.Contains(err.Error(), "exceed 500") {
			t.Errorf("expected 'exceed 500' error, got %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if err := validateGraphQLReturnType("<script>"); err == nil {
			t.Errorf("expected error for invalid return type")
		}
	})
}

func TestValidateGraphQLArguments(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		args := []model.GraphQLArgument{
			{Name: "id", Type: "ID!", Description: "User ID"},
			{Name: "name", Type: "String", Description: "Name"},
		}
		if err := validateGraphQLArguments(args); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("too many", func(t *testing.T) {
		args := make([]model.GraphQLArgument, 51)
		for i := range 51 {
			args[i] = model.GraphQLArgument{Name: "arg", Type: "String"}
		}
		if err := validateGraphQLArguments(args); err == nil || !strings.Contains(err.Error(), "maximum 50") {
			t.Errorf("expected 'maximum 50' error, got %v", err)
		}
	})

	t.Run("duplicate names", func(t *testing.T) {
		args := []model.GraphQLArgument{
			{Name: "id", Type: "ID!"},
			{Name: "id", Type: "String"},
		}
		if err := validateGraphQLArguments(args); err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("expected 'duplicate' error, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		args := []model.GraphQLArgument{{Name: "", Type: "String"}}
		if err := validateGraphQLArguments(args); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("required with default", func(t *testing.T) {
		defaultVal := "hello"
		args := []model.GraphQLArgument{
			{Name: "name", Type: "String", Required: true, DefaultValue: defaultVal},
		}
		if err := validateGraphQLArguments(args); err == nil || !strings.Contains(err.Error(), "cannot have default") {
			t.Errorf("expected 'cannot have default' error, got %v", err)
		}
	})
}

func TestValidateGraphQLArgumentName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, n := range []string{"id", "userName", "_field", "arg1"} {
			if err := validateGraphQLArgumentName(n); err != nil {
				t.Errorf("unexpected error for %q: %v", n, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateGraphQLArgumentName(""); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 101)
		if err := validateGraphQLArgumentName(long); err == nil || !strings.Contains(err.Error(), "maximum 100") {
			t.Errorf("expected 'maximum 100' error, got %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, n := range []string{"123abc", "hello world", "test-name"} {
			if err := validateGraphQLArgumentName(n); err == nil {
				t.Errorf("expected error for %q", n)
			}
		}
	})
}

func TestValidateGraphQLType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, gt := range []string{"String", "ID!", "CustomType"} {
			if err := validateGraphQLType(gt); err != nil {
				t.Errorf("unexpected error for %q: %v", gt, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateGraphQLType(""); err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("expected 'required' error, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 101)
		if err := validateGraphQLType(long); err == nil || !strings.Contains(err.Error(), "maximum 100") {
			t.Errorf("expected 'maximum 100' error, got %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, gt := range []string{"<String>", "123int", "hello world"} {
			if err := validateGraphQLType(gt); err == nil {
				t.Errorf("expected error for %q", gt)
			}
		}
	})
}

func TestValidateGraphQLExamples(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		examples := []model.GraphQLExample{
			{Name: "GetUser", Query: "query { user(id: 1) { name } }"},
		}
		if err := validateGraphQLExamples(examples); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("too many", func(t *testing.T) {
		examples := make([]model.GraphQLExample, 21)
		for i := range 21 {
			examples[i] = model.GraphQLExample{Name: "Ex", Query: "q"}
		}
		if err := validateGraphQLExamples(examples); err == nil || !strings.Contains(err.Error(), "maximum 20") {
			t.Errorf("expected 'maximum 20' error, got %v", err)
		}
	})

	t.Run("name required", func(t *testing.T) {
		examples := []model.GraphQLExample{{Name: "", Query: "q"}}
		if err := validateGraphQLExamples(examples); err == nil || !strings.Contains(err.Error(), "name is required") {
			t.Errorf("expected 'name is required' error, got %v", err)
		}
	})

	t.Run("name too long", func(t *testing.T) {
		examples := []model.GraphQLExample{{Name: strings.Repeat("a", 101), Query: "q"}}
		if err := validateGraphQLExamples(examples); err == nil || !strings.Contains(err.Error(), "name too long") {
			t.Errorf("expected 'name too long' error, got %v", err)
		}
	})

	t.Run("query required", func(t *testing.T) {
		examples := []model.GraphQLExample{{Name: "Test", Query: ""}}
		if err := validateGraphQLExamples(examples); err == nil || !strings.Contains(err.Error(), "query is required") {
			t.Errorf("expected 'query is required' error, got %v", err)
		}
	})

	t.Run("query too long", func(t *testing.T) {
		examples := []model.GraphQLExample{{Name: "Test", Query: strings.Repeat("q", 10001)}}
		if err := validateGraphQLExamples(examples); err == nil || !strings.Contains(err.Error(), "too long") {
			t.Errorf("expected 'too long' error, got %v", err)
		}
	})

	t.Run("variables must be object", func(t *testing.T) {
		examples := []model.GraphQLExample{{
			Name: "Test", Query: "q", Variables: "not an object",
		}}
		if err := validateGraphQLExamples(examples); err == nil || !strings.Contains(err.Error(), "not a string") {
			t.Errorf("expected 'not a string' error, got %v", err)
		}
	})

	t.Run("response must be object", func(t *testing.T) {
		examples := []model.GraphQLExample{{
			Name: "Test", Query: "q", Response: "not an object",
		}}
		if err := validateGraphQLExamples(examples); err == nil || !strings.Contains(err.Error(), "not a string") {
			t.Errorf("expected 'not a string' error, got %v", err)
		}
	})
}
