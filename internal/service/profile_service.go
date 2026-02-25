package service

import (
	"context"
	"errors"
	"fmt"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/repository"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
var websiteRegex = regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`)

type ProfileService struct {
	repo     *repository.ProfileRepository
	userRepo *repository.UserRepository
}

func NewProfileService(repo *repository.ProfileRepository, userRepo *repository.UserRepository) *ProfileService {
	return &ProfileService{
		repo:     repo,
		userRepo: userRepo,
	}
}

// helper to convert pgtype.Int8 to int8, handling NULL
func int8FromPgType(val pgtype.Int8) int8 {
	if val.Valid {
		return int8(val.Int64)
	}
	return 0
}

// helper to convert pgtype.Text to string, handling NULL
func stringFromPgType(val pgtype.Text) string {
	if val.Valid {
		return val.String
	}
	return ""
}

// helper to convert *int8 to pgtype.Int8
func pgTypeInt8(val *int8) pgtype.Int8 {
	if val == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: int64(*val), Valid: true}
}

// helper to convert *string to pgtype.Text
func pgTypeText(val *string) pgtype.Text {
	if val == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *val, Valid: true}
}

func (s *ProfileService) validateProfileInput(p model.Profile) error {
	type fieldError struct {
		Field string
		Msg   string
	}
	var errs []fieldError

	// Validate age (only if set)
	age := int8FromPgType(p.Age)
	if p.Age.Valid && (age < 0 || age > 120) {
		errs = append(errs, fieldError{"Age", "must be between 0 and 120"})
	}

	// Validate gender (only if set)
	gender := stringFromPgType(p.Gender)
	if p.Gender.Valid {
		gender = strings.ToLower(strings.TrimSpace(gender))
		if gender != "" && gender != "male" && gender != "female" && gender != "other" {
			errs = append(errs, fieldError{"Gender", "must be 'male', 'female', or 'other'"})
		}
	}

	bio := stringFromPgType(p.Bio)
	bioLen := utf8.RuneCountInString(bio)
	if p.Bio.Valid && bioLen > 500 {
		errs = append(errs, fieldError{"Bio", "maximum length is 500 characters"})
	}

	phoneNumber := stringFromPgType(p.PhoneNumber)
	if p.PhoneNumber.Valid && phoneNumber != "" && !phoneRegex.MatchString(phoneNumber) {
		errs = append(errs, fieldError{"PhoneNumber", "invalid format (expected 10-15 digits, optionally starting with +)"})
	}

	website := stringFromPgType(p.Website)
	if p.Website.Valid && website != "" && !websiteRegex.MatchString(website) {
		errs = append(errs, fieldError{"Website", "invalid format (expected http:// or https:// URL)"})
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

func (s *ProfileService) CreateProfile(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error) {
	if _, err := s.userRepo.FindByID(ctx, req.UserID); err != nil {
		return nil, fmt.Errorf("user with Id %d not found", req.UserID)
	}

	// Build Profile for validation (convert pointers to pgtype)
	validateProfile := model.Profile{
		UserID:      req.UserID,
		Age:         pgTypeInt8(req.Age),
		Gender:      pgTypeText(req.Gender),
		Bio:         pgTypeText(req.Bio),
		PhoneNumber: pgTypeText(req.PhoneNumber),
		Website:     pgTypeText(req.Website),
	}

	if err := s.validateProfileInput(validateProfile); err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error) {
	if req.ProfileId == 0 {
		return nil, fmt.Errorf("id is required for update")
	}

	stored, err := s.repo.FindByID(ctx, req.ProfileId)
	if err != nil {
		return nil, err
	}

	// Build merged profile for validation
	merged := stored
	if req.Age != nil {
		merged.Age = pgTypeInt8(req.Age)
	}
	if req.Gender != nil {
		merged.Gender = pgTypeText(req.Gender)
	}
	if req.Bio != nil {
		merged.Bio = pgTypeText(req.Bio)
	}
	if req.PhoneNumber != nil {
		merged.PhoneNumber = pgTypeText(req.PhoneNumber)
	}
	if req.Website != nil {
		merged.Website = pgTypeText(req.Website)
	}

	if err := s.validateProfileInput(merged); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &merged, nil
}

func (s *ProfileService) GetAllProfiles(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error) {
	return s.repo.FindAllWithUsers(ctx, req, lastCursor)
}

func (s *ProfileService) GetProfileByID(ctx context.Context, id int64) (model.Profile, error) {
	profile, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.Profile{}, err
	}

	return profile, nil
}

func (s *ProfileService) GetProfileByUserID(ctx context.Context, userID int64) (model.Profile, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return model.Profile{}, err
	}
	return profile, nil
}

func (s *ProfileService) DeleteProfile(ctx context.Context, profileId int64) error {
	return s.repo.Delete(ctx, profileId)
}
