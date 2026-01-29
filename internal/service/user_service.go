package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

var nameRegex = regexp.MustCompile(`^[A-Za-z ]+$`)
var addressRegex = regexp.MustCompile(`^[A-Za-z ]+$`)
var emailRegex = regexp.MustCompile(`^[A-Za-z]+(?:\.[A-Za-z]+)*@[A-Za-z]+(?:\.[A-Za-z]+)+$`)

type UserService struct {
	repo *repository.UserRepository
}

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultArgonParams = &argonParams{
	memory:      64 * 1024,
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// HashPassword hashes a password using Argon2id (exported for testing)
func HashPassword(password string) (string, error) {
	salt := make([]byte, defaultArgonParams.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultArgonParams.iterations,
		defaultArgonParams.memory,
		defaultArgonParams.parallelism,
		defaultArgonParams.keyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultArgonParams.memory,
		defaultArgonParams.iterations,
		defaultArgonParams.parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword verifies a password against an Argon2id hash (exported for testing)
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Format: $argon2id$v=<version>$m=<memory>,t=<time>,p=<parallelism>$<base64-salt>$<base64-hash>
	// Example: $argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQ=$c29tZWhhc2g=

	// Split by '$' to properly parse each component
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}

	// Parts: ["" "argon2id" "v=19" "m=65536,t=3,p=2" "<base64-salt>" "<base64-hash>"]

	// Parse version from "v=19"
	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("invalid version format: %w", err)
	}

	// Parse params from "m=65536,t=3,p=2"
	var memory, iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("invalid params format: %w", err)
	}

	// Decode base64 salt
	salt := parts[4]
	decodedSalt, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode base64 hash
	hash := parts[5]
	decodedHash, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Compute hash of provided password
	comparisonHash := argon2.IDKey(
		[]byte(password),
		decodedSalt,
		iterations,
		memory,
		parallelism,
		uint32(len(decodedHash)),
	)

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}

func (s *UserService) validateUserInput(u model.User) error {
	type fieldError struct {
		Field string
		Msg   string
	}
	var errs []fieldError

	nameLen := utf8.RuneCountInString(u.Name)
	if nameLen == 0 || nameLen > 100 {
		errs = append(errs, fieldError{"Name", "must be 1..100 characters"})
	} else if !nameRegex.MatchString(u.Name) {
		errs = append(errs, fieldError{"Name", "only letters A-Z, a-z and spaces are allowed"})
	}

	if utf8.RuneCountInString(u.Email) == 0 || utf8.RuneCountInString(u.Email) > 100 {
		errs = append(errs, fieldError{"Email", "must be 1..100 characters"})
	} else if !emailRegex.MatchString(u.Email) {
		errs = append(errs, fieldError{"Email", "invalid format (expected something like test@email.com)"})
	}

	if len(errs) == 0 {
		return nil
	}

	var parts []string
	for _, e := range errs {
		parts = append(parts, fmt.Sprintf("%s: %s", e.Field, e.Msg))
	}
	return errors.New(strings.Join(parts, "; "))
}

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	user := model.User{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.validateUserInput(user); err != nil {
		return nil, err
	}

	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	req.Password = hashedPassword

	created, err := s.repo.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req model.UpdateUserRequest) error {
	user := model.User{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.validateUserInput(user); err != nil {
		return fmt.Errorf("failed to validate user input: %w", err)
	}

	updated := s.repo.Update(ctx, req)
	if updated != nil {
		return updated
	}
	return updated
}

func (s *UserService) GetAllUsers(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error) {
	return s.repo.FindAll(ctx, req, lastCursor)
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) UpdatePassword(ctx context.Context, userID int64, req model.UpdatePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	valid, err := VerifyPassword(req.CurrentPassword, user.PasswordHash)
	if err != nil || !valid {
		return errors.New("current password is incorrect")
	}

	if len(req.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	hashedPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, userID, hashedPassword)
}
