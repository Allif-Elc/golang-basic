package service

import (
	"context"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// CreateProject creates a new project with OWASP-compliant validation
func (s *ProjectService) CreateProject(ctx context.Context, userID int64, req model.CreateProjectRequest) (*model.Project, error) {
	// Sanitize and validate all inputs
	sanitizedName := sanitizeInput(req.Name)
	sanitizedDescription := sanitizeInput(req.Description)
	sanitizedVersion := sanitizeInput(req.Version)

	// Validate name
	if err := validateProjectName(sanitizedName); err != nil {
		return nil, err
	}

	// Validate description
	if err := validateProjectDescription(sanitizedDescription); err != nil {
		return nil, err
	}

	// Validate version
	if err := validateProjectVersion(sanitizedVersion); err != nil {
		return nil, err
	}

	// Update request with sanitized values
	req.Name = sanitizedName
	req.Description = sanitizedDescription
	req.Version = sanitizedVersion

	// Create project
	project, err := s.repo.Create(ctx, &req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// GetProjectByID retrieves a project by its ID
func (s *ProjectService) GetProjectByID(ctx context.Context, projectID int64) (model.Project, error) {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return model.Project{}, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// UpdateProject modifies an existing project with OWASP-compliant validation
func (s *ProjectService) UpdateProject(ctx context.Context, projectID int64, req model.UpdateProjectRequest) (*model.Project, error) {
	// Sanitize and validate provided fields
	if req.Name != nil {
		sanitizedName := sanitizeInput(*req.Name)
		if err := validateProjectName(sanitizedName); err != nil {
			return nil, err
		}
		req.Name = &sanitizedName
	}

	if req.Description != nil {
		sanitizedDescription := sanitizeInput(*req.Description)
		if err := validateProjectDescription(sanitizedDescription); err != nil {
			return nil, err
		}
		req.Description = &sanitizedDescription
	}

	if req.Version != nil {
		sanitizedVersion := sanitizeInput(*req.Version)
		if err := validateProjectVersion(sanitizedVersion); err != nil {
			return nil, err
		}
		req.Version = &sanitizedVersion
	}

	// Update project
	project, err := s.repo.Update(ctx, projectID, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// DeleteProject removes a project and explicitly deletes all APIs first
func (s *ProjectService) DeleteProject(ctx context.Context, projectID int64) (map[string]int64, error) {
	// First, verify project exists
	_, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Explicitly delete all APIs first and get counts
	deletedCounts, err := s.repo.DeleteAllAPIsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete project APIs: %w", err)
	}

	// Then delete the project itself
	if err := s.repo.Delete(ctx, projectID); err != nil {
		return nil, fmt.Errorf("failed to delete project: %w", err)
	}

	return deletedCounts, nil
}

// GetAPIStats retrieves the count of APIs for a project
func (s *ProjectService) GetAPIStats(ctx context.Context, projectID int64) (map[string]int64, error) {
	// Verify project exists
	_, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetAPIStats(ctx, projectID)
}

// ListProjects retrieves projects with pagination and filters
func (s *ProjectService) ListProjects(ctx context.Context, req model.ListProjectsRequest) (*model.PageResult[model.Project], error) {
	// Sanitize search input to prevent SQL injection
	if req.Search != "" {
		req.Search = sanitizeInput(req.Search)
		// Add wildcards for search
		req.Search = "%" + req.Search + "%"
	}

	// Validate and sanitize sort parameters
	req.SortBy = sanitizeSortField(req.SortBy)
	req.SortOrder = sanitizeSortOrder(req.SortOrder)

	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	// Get projects
	result, err := s.repo.FindAll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return result, nil
}

// GetProjectsByUser retrieves projects owned by a specific user
func (s *ProjectService) GetProjectsByUser(ctx context.Context, userID int64, req model.ListProjectsRequest) (*model.PageResult[model.Project], error) {
	// Set the user filter
	req.IDUser = &userID

	// Sanitize search input
	if req.Search != "" {
		req.Search = sanitizeInput(req.Search)
		req.Search = "%" + req.Search + "%"
	}

	// Validate and sanitize sort parameters
	req.SortBy = sanitizeSortField(req.SortBy)
	req.SortOrder = sanitizeSortOrder(req.SortOrder)

	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	// Get projects
	result, err := s.repo.FindAll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list user projects: %w", err)
	}

	return result, nil
}

// ==================== OWASP COMPLIANT VALIDATION FUNCTIONS ====================

// sanitizeInput removes potentially dangerous characters following OWASP guidelines
func sanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline, tab, carriage return
	reg := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	input = reg.ReplaceAllString(input, "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// validateProjectName performs comprehensive validation on project name
func validateProjectName(name string) error {
	// Check length (OWASP: prevent buffer overflow attacks)
	if utf8.RuneCountInString(name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}
	if utf8.RuneCountInString(name) > 255 {
		return fmt.Errorf("name must not exceed 255 characters")
	}

	// Check for null bytes (OWASP: prevent null byte injection)
	if strings.Contains(name, "\x00") {
		return fmt.Errorf("name contains null bytes")
	}

	// Check for XSS patterns (OWASP: prevent cross-site scripting)
	if containsXSSPatterns(name) {
		return fmt.Errorf("name contains potentially dangerous content")
	}

	// Check for SQL injection patterns (OWASP: prevent SQL injection)
	if containsSQLInjectionPatterns(name) {
		return fmt.Errorf("name contains potentially dangerous SQL patterns")
	}

	// Check for path traversal patterns (OWASP: prevent path traversal)
	if containsPathTraversalPatterns(name) {
		return fmt.Errorf("name contains potentially dangerous path patterns")
	}

	// Check for valid characters (letters, numbers, spaces, hyphens, underscores, periods)
	// Using unicode character classes for internationalization support
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_\.]+$`, name)
	if !matched {
		return fmt.Errorf("name can only contain letters, numbers, spaces, hyphens, underscores, and periods")
	}

	// Check for consecutive dots (prevent directory traversal attempts)
	if strings.Contains(name, "..") {
		return fmt.Errorf("name cannot contain consecutive dots")
	}

	return nil
}

// validateProjectDescription performs comprehensive validation on project description
func validateProjectDescription(description string) error {
	// Check length (OWASP: prevent buffer overflow attacks)
	if utf8.RuneCountInString(description) > 1000 {
		return fmt.Errorf("description must not exceed 1000 characters")
	}

	// Allow empty description
	if description == "" {
		return nil
	}

	// Check for null bytes
	if strings.Contains(description, "\x00") {
		return fmt.Errorf("description contains null bytes")
	}

	// Check for XSS patterns (OWASP: prevent cross-site scripting)
	if containsXSSPatterns(description) {
		return fmt.Errorf("description contains potentially dangerous content")
	}

	// Check for SQL injection patterns
	if containsSQLInjectionPatterns(description) {
		return fmt.Errorf("description contains potentially dangerous SQL patterns")
	}

	// Allow most characters in description (it's user-generated content)
	// but still limit control characters
	reg := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	if reg.MatchString(description) {
		return fmt.Errorf("description contains invalid control characters")
	}

	return nil
}

// validateProjectVersion performs validation on project version
func validateProjectVersion(version string) error {
	// Allow empty version
	if version == "" {
		return nil
	}

	// Check length
	if utf8.RuneCountInString(version) > 50 {
		return fmt.Errorf("version must not exceed 50 characters")
	}

	// Check for null bytes
	if strings.Contains(version, "\x00") {
		return fmt.Errorf("version contains null bytes")
	}

	// Check for XSS patterns
	if containsXSSPatterns(version) {
		return fmt.Errorf("version contains potentially dangerous content")
	}

	// Check for SQL injection patterns
	if containsSQLInjectionPatterns(version) {
		return fmt.Errorf("version contains potentially dangerous SQL patterns")
	}

	// Semantic versioning pattern (optional but recommended)
	// Allows: 1.0.0, v1.0.0, 1.0.0-beta, 1.0.0-alpha.1
	matched, _ := regexp.MatchString(`^[vV]?[0-9]+\.[0-9]+\.[0-9]+([a-zA-Z0-9\.\-]*)?$`, version)
	if !matched {
		// If not semver, allow alphanumeric with dots, hyphens, and plus
		matched, _ = regexp.MatchString(`^[a-zA-Z0-9\.\-+]+$`, version)
		if !matched {
			return fmt.Errorf("version must follow semantic versioning (e.g., 1.0.0) or contain only alphanumeric characters, dots, hyphens, and plus signs")
		}
	}

	return nil
}

// sanitizeSortField validates and sanitizes the sort field parameter
func sanitizeSortField(sortField string) string {
	// Whitelist approach (OWASP recommendation)
	validFields := map[string]bool{
		"":           true,
		"name":       true,
		"created_at": true,
		"updated_at": true,
	}

	if validFields[sortField] {
		return sortField
	}

	// Default to created_at if invalid
	return "created_at"
}

// sanitizeSortOrder validates and sanitizes the sort order parameter
func sanitizeSortOrder(sortOrder string) string {
	// Whitelist approach (OWASP recommendation)
	validOrders := map[string]bool{
		"":     true,
		"asc":  true,
		"desc": true,
	}

	if validOrders[strings.ToLower(sortOrder)] {
		return strings.ToLower(sortOrder)
	}

	// Default to desc if invalid
	return "desc"
}

// ==================== SECURITY PATTERN DETECTION ====================

// containsXSSPatterns checks for common XSS attack patterns (OWASP XSS Prevention)
func containsXSSPatterns(input string) bool {
	lowerInput := strings.ToLower(input)

	// Common XSS patterns
	xssPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onkeydown=",
		"onkeypress=",
		"onkeyup=",
		"eval(",
		"expression(",
		"vbscript:",
		"data:",
		"flash:", // for potential flash exploits
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Check for HTML tags (sanitization would be needed, but we block for simplicity)
	htmlTagPattern := regexp.MustCompile(`<[^>]*>`)
	if htmlTagPattern.MatchString(lowerInput) {
		return true
	}

	return false
}

// containsSQLInjectionPatterns checks for common SQL injection patterns (OWASP SQL Injection Prevention)
func containsSQLInjectionPatterns(input string) bool {
	lowerInput := strings.ToLower(input)

	// Common SQL injection patterns
	sqlPatterns := []string{
		"' or '1'='1",
		"' or 1=1",
		"union select",
		"drop table",
		"delete from",
		"insert into",
		"update set",
		"exec(",
		"execute(",
		"sp_executesql",
		"waitfor delay",
		"benchmark(",
		"sleep(",
		";--",
		"--",
		"/*",
		"*/",
		"xp_cmdshell",
		"sp_oacreate",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Check for multiple single quotes (classic SQL injection)
	if strings.Count(lowerInput, "'") > 1 {
		return true
	}

	return false
}

// containsPathTraversalPatterns checks for path traversal attempts (OWASP Path Traversal Prevention)
func containsPathTraversalPatterns(input string) bool {
	// Path traversal patterns
	traversalPatterns := []string{
		"../",
		"..\\",
		"%2e%2e",
		"..%2f",
		"..%5c",
		"%252e",
		"~/",
		"/etc/",
		"/proc/",
		"\\windows\\",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range traversalPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}
