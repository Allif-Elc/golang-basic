package service

import (
	"fmt"
	"testing"

	"golang-basic/api/internal/model"
)

func TestValidateRestAPIDescription(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := validateRestAPIDescription("A normal description"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateRestAPIDescription(""); err != nil {
			t.Errorf("expected no error for empty, got %v", err)
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := make([]byte, 1001)
		for i := range long {
			long[i] = 'a'
		}
		if err := validateRestAPIDescription(string(long)); err == nil {
			t.Error("expected error for too long")
		}
	})

	t.Run("null bytes", func(t *testing.T) {
		if err := validateRestAPIDescription("bad\x00desc"); err == nil {
			t.Error("expected error for null bytes")
		}
	})

	t.Run("XSS", func(t *testing.T) {
		if err := validateRestAPIDescription("<script>alert(1)</script>"); err == nil {
			t.Error("expected XSS error")
		}
	})

	t.Run("SQL injection", func(t *testing.T) {
		if err := validateRestAPIDescription("' OR '1'='1"); err == nil {
			t.Error("expected SQL error")
		}
	})
}

func TestValidateHeaderName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, n := range []string{"Authorization", "Content-Type", "X-Custom-Header", "Accept"} {
			if err := validateHeaderName(n); err != nil {
				t.Errorf("expected no error for %q, got %v", n, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateHeaderName(""); err == nil {
			t.Error("expected error for empty header name")
		}
	})

	t.Run("too long", func(t *testing.T) {
		long := make([]byte, 101)
		for i := range long {
			long[i] = 'a'
		}
		if err := validateHeaderName(string(long)); err == nil {
			t.Error("expected error for too long")
		}
	})

	t.Run("invalid chars", func(t *testing.T) {
		for _, n := range []string{"header name!", "header@test", "foo/bar"} {
			if err := validateHeaderName(n); err == nil {
				t.Errorf("expected error for invalid header %q", n)
			}
		}
	})
}

func TestValidateParameterName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, n := range []string{"id", "user_id", "pageSize", "sort_order"} {
			if err := validateParameterName(n); err != nil {
				t.Errorf("expected no error for %q, got %v", n, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := validateParameterName(""); err == nil {
			t.Error("expected error for empty")
		}
	})

	t.Run("invalid chars", func(t *testing.T) {
		for _, n := range []string{"param-name", "param.name", "param name"} {
			if err := validateParameterName(n); err == nil {
				t.Errorf("expected error for invalid %q", n)
			}
		}
	})
}

func TestValidateParameterType(t *testing.T) {
	for _, typ := range []string{"string", "integer", "boolean", "number"} {
		if err := validateParameterType(typ); err != nil {
			t.Errorf("expected no error for %q, got %v", typ, err)
		}
	}

	for _, typ := range []string{"invalid", "int", "bool", "array", "object"} {
		if err := validateParameterType(typ); err == nil {
			t.Errorf("expected error for invalid type %q", typ)
		}
	}
}

func TestValidateDefaultValue(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		if err := validateDefaultValue(nil, "string"); err != nil {
			t.Errorf("expected no error for nil, got %v", err)
		}
	})

	t.Run("string", func(t *testing.T) {
		if err := validateDefaultValue("hello", "string"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := validateDefaultValue(123, "string"); err == nil {
			t.Error("expected error for non-string value")
		}
	})

	t.Run("integer", func(t *testing.T) {
		if err := validateDefaultValue(42, "integer"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := validateDefaultValue(42.0, "integer"); err != nil {
			t.Errorf("expected no error for float64 that is int, got %v", err)
		}
		if err := validateDefaultValue("hello", "integer"); err == nil {
			t.Error("expected error for string as integer")
		}
	})

	t.Run("boolean", func(t *testing.T) {
		if err := validateDefaultValue(true, "boolean"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := validateDefaultValue("true", "boolean"); err == nil {
			t.Error("expected error for string as boolean")
		}
	})

	t.Run("number", func(t *testing.T) {
		if err := validateDefaultValue(3.14, "number"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := validateDefaultValue(42, "number"); err != nil {
			t.Errorf("expected no error for int as number, got %v", err)
		}
		if err := validateDefaultValue("3.14", "number"); err == nil {
			t.Error("expected error for string as number")
		}
	})
}

func TestValidateJSONSchemaType(t *testing.T) {
	for _, typ := range []string{"object", "array", "string", "number", "integer", "boolean", "null"} {
		if err := validateJSONSchemaType(typ); err != nil {
			t.Errorf("expected no error for %q, got %v", typ, err)
		}
	}

	if err := validateJSONSchemaType("invalid"); err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestValidateJSONSchema(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if err := validateJSONSchema(nil); err != nil {
			t.Errorf("expected no error for nil, got %v", err)
		}
	})

	t.Run("valid object", func(t *testing.T) {
		schema := &model.JSONSchema{
			Type: "object",
			Properties: map[string]model.Property{
				"name": {Type: "string", Description: "User name"},
				"age":  {Type: "integer", Description: "User age"},
			},
		}
		if err := validateJSONSchema(schema); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid array", func(t *testing.T) {
		schema := &model.JSONSchema{
			Type:  "array",
			Items: &model.JSONSchema{Type: "string"},
		}
		if err := validateJSONSchema(schema); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		schema := &model.JSONSchema{Type: "invalid"}
		if err := validateJSONSchema(schema); err == nil {
			t.Error("expected error for invalid type")
		}
	})

	t.Run("too many properties", func(t *testing.T) {
		props := make(map[string]model.Property)
		for i := 0; i < 101; i++ {
			props[fmt.Sprintf("prop%d", i)] = model.Property{Type: "string"}
		}
		schema := &model.JSONSchema{Type: "object", Properties: props}
		if err := validateJSONSchema(schema); err == nil {
			t.Error("expected error for too many properties")
		}
	})
}

func TestValidateProperty(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		if err := validateProperty(&model.Property{Type: "string"}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid with enum", func(t *testing.T) {
		prop := &model.Property{
			Type: "string",
			Enum: []string{"a", "b", "c"},
		}
		if err := validateProperty(prop); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("too many enum", func(t *testing.T) {
		enum := make([]string, 51)
		for i := range enum {
			enum[i] = "v"
		}
		if err := validateProperty(&model.Property{Type: "string", Enum: enum}); err == nil {
			t.Error("expected error for too many enum values")
		}
	})

	t.Run("enum value too long", func(t *testing.T) {
		long := make([]byte, 101)
		for i := range long {
			long[i] = 'a'
		}
		if err := validateProperty(&model.Property{Type: "string", Enum: []string{string(long)}}); err == nil {
			t.Error("expected error for enum value too long")
		}
	})

	t.Run("description too long", func(t *testing.T) {
		long := make([]byte, 501)
		for i := range long {
			long[i] = 'a'
		}
		if err := validateProperty(&model.Property{Type: "string", Description: string(long)}); err == nil {
			t.Error("expected error for description too long")
		}
	})

	t.Run("description XSS", func(t *testing.T) {
		if err := validateProperty(&model.Property{Type: "string", Description: "<script>"}); err == nil {
			t.Error("expected error for XSS in description")
		}
	})

	t.Run("nested too many", func(t *testing.T) {
		nested := make(map[string]model.Property)
		for i := 0; i < 51; i++ {
			nested[fmt.Sprintf("prop%d", i)] = model.Property{Type: "string"}
		}
		if err := validateProperty(&model.Property{Type: "object", Properties: nested}); err == nil {
			t.Error("expected error for too many nested properties")
		}
	})
}

func TestValidateResponses(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		resp := map[int]model.ResponseExample{
			200: {Description: "Success", Body: map[string]interface{}{"id": 1}},
			404: {Description: "Not Found"},
		}
		if err := validateResponses(resp); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("too many", func(t *testing.T) {
		resp := make(map[int]model.ResponseExample)
		for i := 0; i < 21; i++ {
			resp[200+i] = model.ResponseExample{}
		}
		if err := validateResponses(resp); err == nil {
			t.Error("expected error for too many responses")
		}
	})

	t.Run("invalid status code", func(t *testing.T) {
		resp := map[int]model.ResponseExample{999: {Description: "Invalid"}}
		if err := validateResponses(resp); err == nil {
			t.Error("expected error for invalid status code")
		}
	})

	t.Run("description too long", func(t *testing.T) {
		long := make([]byte, 501)
		for i := range long {
			long[i] = 'a'
		}
		resp := map[int]model.ResponseExample{200: {Description: string(long)}}
		if err := validateResponses(resp); err == nil {
			t.Error("expected error for description too long")
		}
	})

	t.Run("body too large", func(t *testing.T) {
		body := make([]byte, 10001)
		body[0] = '{'
		body[10000] = '}'
		resp := map[int]model.ResponseExample{200: {Body: string(body)}}
		if err := validateResponses(resp); err == nil {
			t.Error("expected error for body too large")
		}
	})

	t.Run("non-JSON body", func(t *testing.T) {
		resp := map[int]model.ResponseExample{200: {Body: make(chan int)}}
		if err := validateResponses(resp); err == nil {
			t.Error("expected error for non-JSON body")
		}
	})
}
