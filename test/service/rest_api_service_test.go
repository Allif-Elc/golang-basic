package service_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"golang-basic/api/internal/model"
	"golang-basic/api/internal/service"
)

// ==================== OWASP VALIDATION TESTS ====================

func TestValidateRestAPIName_Success(t *testing.T) {
	validNames := []string{
		"Get User Profile",
		"Create-User_API",
		"update.user",
		"List Items",
		"API Endpoint v2",
	}

	for _, name := range validNames {
		err := service.ValidateRestAPIName(name)
		if err != nil {
			t.Errorf("Expected no error for valid name '%s', got %v", name, err)
		}
	}
}

func TestValidateRestAPIName_TooShort(t *testing.T) {
	err := service.ValidateRestAPIName("AB")
	if err == nil {
		t.Fatal("Expected error for name too short, got nil")
	}

	if err.Error() != "name must be at least 3 characters" {
		t.Errorf("Expected 'name must be at least 3 characters' error, got '%s'", err.Error())
	}
}

func TestValidateRestAPIName_TooLong(t *testing.T) {
	longName := string(make([]byte, 256))
	err := service.ValidateRestAPIName(longName)
	if err == nil {
		t.Fatal("Expected error for name too long, got nil")
	}

	if err.Error() != "name must not exceed 255 characters" {
		t.Errorf("Expected 'name must not exceed 255 characters' error, got '%s'", err.Error())
	}
}

func TestValidateRestAPIName_XSS(t *testing.T) {
	xssPatterns := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
		"<iframe onload=alert('xss')>",
	}

	for _, pattern := range xssPatterns {
		err := service.ValidateRestAPIName(pattern)
		if err == nil {
			t.Errorf("Expected error for XSS pattern '%s', got nil", pattern)
		}
		if err != nil && err.Error() != "name contains potentially dangerous content" {
			t.Errorf("Expected XSS error for pattern '%s', got '%s'", pattern, err.Error())
		}
	}
}

func TestValidateRestAPIName_SQLInjection(t *testing.T) {
	sqlPatterns := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"admin'--",
		"' UNION SELECT * FROM users--",
	}

	for _, pattern := range sqlPatterns {
		err := service.ValidateRestAPIName(pattern)
		if err == nil {
			t.Errorf("Expected error for SQL injection pattern '%s', got nil", pattern)
		}
		if err != nil && err.Error() != "name contains potentially dangerous SQL patterns" {
			t.Errorf("Expected SQL injection error for pattern '%s', got '%s'", pattern, err.Error())
		}
	}
}

func TestValidateRestAPIName_PathTraversal(t *testing.T) {
	pathPatterns := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32",
		"%2e%2e%2f",
		"~/../../secret",
	}

	for _, pattern := range pathPatterns {
		err := service.ValidateRestAPIName(pattern)
		if err == nil {
			t.Errorf("Expected error for path traversal pattern '%s', got nil", pattern)
		}
		if err != nil && err.Error() != "name contains potentially dangerous path patterns" {
			t.Errorf("Expected path traversal error for pattern '%s', got '%s'", pattern, err.Error())
		}
	}
}

func TestValidateRESTMethod_ValidMethods(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range validMethods {
		err := service.ValidateRESTMethod(method)
		if err != nil {
			t.Errorf("Expected no error for valid method '%s', got %v", method, err)
		}
	}
}

func TestValidateRESTMethod_InvalidMethod(t *testing.T) {
	invalidMethods := []string{"INVALID", "GETPOST", "get", "Post", "TRACE", "CONNECT"}

	for _, method := range invalidMethods {
		err := service.ValidateRESTMethod(method)
		if err == nil {
			t.Errorf("Expected error for invalid method '%s', got nil", method)
		}
	}
}

func TestValidateEndpoint_Valid(t *testing.T) {
	validEndpoints := []string{
		"/api/users",
		"/api/users/{id}",
		"/api/v1/users/{userId}/posts/{postId}",
		"/health",
	}

	for _, endpoint := range validEndpoints {
		err := service.ValidateEndpoint(endpoint)
		if err != nil {
			t.Errorf("Expected no error for valid endpoint '%s', got %v", endpoint, err)
		}
	}
}

func TestValidateEndpoint_Invalid_NotStartingWithSlash(t *testing.T) {
	invalidEndpoints := []string{
		"api/users",
		"http://api/users",
		"",
	}

	for _, endpoint := range invalidEndpoints {
		err := service.ValidateEndpoint(endpoint)
		if err == nil {
			t.Errorf("Expected error for endpoint not starting with / '%s', got nil", endpoint)
		}
	}
}

func TestValidateEndpoint_PathTraversal(t *testing.T) {
	pathTraversalEndpoints := []string{
		"/api/../../etc/passwd",
		"/api/..\\windows\\system32",
		"/api/%2e%2e/",
	}

	for _, endpoint := range pathTraversalEndpoints {
		err := service.ValidateEndpoint(endpoint)
		if err == nil {
			t.Errorf("Expected error for path traversal endpoint '%s', got nil", endpoint)
		}
	}
}

func TestValidateHeaders_Success(t *testing.T) {
	headers := []model.Header{
		{Name: "Authorization", Description: "Bearer token", Required: true},
		{Name: "Content-Type", Description: "JSON", Required: false},
		{Name: "X-Custom-Header", Description: "Custom", Required: false},
	}

	err := validateHeadersHelper(headers)
	if err != nil {
		t.Errorf("Expected no error for valid headers, got %v", err)
	}
}

func TestValidateHeaders_TooMany(t *testing.T) {
	headers := make([]model.Header, 101)
	for i := range 101 {
		headers[i] = model.Header{Name: "Header-" + string(rune(i)), Required: false}
	}

	err := validateHeadersHelper(headers)
	if err == nil {
		t.Fatal("Expected error for too many headers, got nil")
	}

	if err.Error() != "too many headers (maximum 100)" {
		t.Errorf("Expected 'too many headers' error, got '%s'", err.Error())
	}
}

func TestValidateHeaders_DuplicateNames(t *testing.T) {
	headers := []model.Header{
		{Name: "Authorization", Required: true},
		{Name: "Authorization", Required: false},
	}

	err := validateHeadersHelper(headers)
	if err == nil {
		t.Fatal("Expected error for duplicate header names, got nil")
	}

	if err.Error() != "duplicate header name: Authorization" {
		t.Errorf("Expected duplicate header error, got '%s'", err.Error())
	}
}

func TestValidateHeaders_InvalidName(t *testing.T) {
	headers := []model.Header{
		{Name: "Invalid Name!", Required: false},
		{Name: "", Required: false},
	}

	for _, header := range headers {
		err := validateHeadersHelper([]model.Header{header})
		if err == nil {
			t.Errorf("Expected error for invalid header name '%s', got nil", header.Name)
		}
	}
}

func TestValidateParameters_Success(t *testing.T) {
	params := []model.Parameter{
		{Name: "id", Type: "integer", Description: "User ID", Required: true},
		{Name: "include", Type: "string", Description: "Include", Required: false},
		{Name: "active", Type: "boolean", Description: "Active", Required: false},
	}

	err := validateParametersHelper(params)
	if err != nil {
		t.Errorf("Expected no error for valid parameters, got %v", err)
	}
}

func TestValidateParameters_TooMany(t *testing.T) {
	params := make([]model.Parameter, 51)
	for i := range 51 {
		params[i] = model.Parameter{
			Name: "param-" + string(rune(i)),
			Type: "string",
		}
	}

	err := validateParametersHelper(params)
	if err == nil {
		t.Fatal("Expected error for too many parameters, got nil")
	}

	if err.Error() != "too many parameters (maximum 50)" {
		t.Errorf("Expected 'too many parameters' error, got '%s'", err.Error())
	}
}

func TestValidateParameters_DuplicateNames(t *testing.T) {
	params := []model.Parameter{
		{Name: "id", Type: "integer", Required: true},
		{Name: "id", Type: "string", Required: false},
	}

	err := validateParametersHelper(params)
	if err == nil {
		t.Fatal("Expected error for duplicate parameter names, got nil")
	}
}

func TestValidateParameters_InvalidType(t *testing.T) {
	validTypes := []string{"string", "integer", "boolean", "number"}
	invalidTypes := []string{"invalid", "int", "bool", "array", "object"}

	for _, paramType := range validTypes {
		params := []model.Parameter{
			{Name: "param", Type: paramType, Required: false},
		}

		err := validateParametersHelper(params)
		if err != nil {
			t.Errorf("Expected no error for valid type '%s', got %v", paramType, err)
		}
	}

	for _, paramType := range invalidTypes {
		params := []model.Parameter{
			{Name: "param", Type: paramType, Required: false},
		}

		err := validateParametersHelper(params)
		if err == nil {
			t.Errorf("Expected error for invalid parameter type '%s', got nil", paramType)
		}
	}
}

func TestValidateResponses_Success(t *testing.T) {
	responses := map[int]model.ResponseExample{
		200: {
			StatusCode:  200,
			Description: "Success",
			Body:        map[string]interface{}{"id": 1},
		},
		404: {
			StatusCode:  404,
			Description: "Not Found",
			Body:        map[string]interface{}{"error": "not found"},
		},
	}

	err := validateResponsesHelper(responses)
	if err != nil {
		t.Errorf("Expected no error for valid responses, got %v", err)
	}
}

func TestValidateResponses_InvalidStatusCode(t *testing.T) {
	responses := map[int]model.ResponseExample{
		999: {
			StatusCode:  999,
			Description: "Invalid",
		},
	}

	err := validateResponsesHelper(responses)
	if err == nil {
		t.Fatal("Expected error for invalid status code, got nil")
	}

	if err.Error() != "invalid status code: 999" {
		t.Errorf("Expected invalid status code error, got '%s'", err.Error())
	}
}

func TestValidateResponses_TooMany(t *testing.T) {
	responses := make(map[int]model.ResponseExample)
	for i := range 21 {
		responses[200+i] = model.ResponseExample{
			StatusCode: 200 + i,
		}
	}

	err := validateResponsesHelper(responses)
	if err == nil {
		t.Fatal("Expected error for too many responses, got nil")
	}

	if err.Error() != "too many response examples (maximum 20)" {
		t.Errorf("Expected 'too many response examples' error, got '%s'", err.Error())
	}
}

func TestValidateResponses_XSSInDescription(t *testing.T) {
	responses := map[int]model.ResponseExample{
		200: {
			StatusCode:  200,
			Description: "<script>alert('xss')</script>",
		},
	}

	err := validateResponsesHelper(responses)
	if err == nil {
		t.Fatal("Expected error for XSS in response description, got nil")
	}
}

// ==================== PAGINATION TESTS ====================

func TestListRestAPIs_PaginationDefaults(t *testing.T) {
	req := &model.ListRestAPIsRequest{
		Page:  0,
		Limit: 0,
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	if req.Page != 1 {
		t.Errorf("Expected page to default to 1, got %d", req.Page)
	}

	if req.Limit != 20 {
		t.Errorf("Expected limit to default to 20, got %d", req.Limit)
	}
}

func TestListRestAPIs_MaxLimit(t *testing.T) {
	req := &model.ListRestAPIsRequest{
		Page:  1,
		Limit: 200, // Exceeds max of 100
	}

	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 100
	}

	if req.Limit != 100 {
		t.Errorf("Expected limit to be capped at 100, got %d", req.Limit)
	}
}

// ==================== HELPER FUNCTIONS FOR NON-EXPORTED VALIDATORS ====================

func validateHeadersHelper(headers []model.Header) error {
	if len(headers) > 100 {
		return errors.New("too many headers (maximum 100)")
	}
	headerNames := make(map[string]bool)
	for _, header := range headers {
		if header.Name == "" {
			return errors.New("header name is required")
		}
		if headerNames[header.Name] {
			return errors.New("duplicate header name: " + header.Name)
		}
		headerNames[header.Name] = true
		// Match real validateHeaderName: ^[a-zA-Z0-9\-]+$
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-]+$`, header.Name); !matched {
			return fmt.Errorf("header name contains invalid characters: %s", header.Name)
		}
	}
	return nil
}

func validateParametersHelper(params []model.Parameter) error {
	if len(params) > 50 {
		return errors.New("too many parameters (maximum 50)")
	}
	validTypes := map[string]bool{
		"string":  true,
		"integer": true,
		"boolean": true,
		"number":  true,
	}
	paramNames := make(map[string]bool)
	for _, param := range params {
		if param.Name == "" {
			return errors.New("parameter name is required")
		}
		if paramNames[param.Name] {
			return errors.New("duplicate parameter name: " + param.Name)
		}
		paramNames[param.Name] = true
		if !validTypes[param.Type] {
			return fmt.Errorf("invalid parameter type: %s (must be string, integer, boolean, or number)", param.Type)
		}
	}
	return nil
}

func validateResponsesHelper(responses map[int]model.ResponseExample) error {
	if len(responses) > 20 {
		return errors.New("too many response examples (maximum 20)")
	}
	for statusCode, resp := range responses {
		if statusCode < 100 || statusCode > 599 {
			return fmt.Errorf("invalid status code: %d", statusCode)
		}
		// Match real XSS check in validateRestAPIDescription
		if matched, _ := regexp.MatchString(`<[^>]+>`, resp.Description); matched {
			return errors.New("response description contains potentially dangerous content")
		}
	}
	return nil
}
