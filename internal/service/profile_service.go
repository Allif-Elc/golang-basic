package service

import (
	"context"
	"errors"
	"fmt"
	"golang-basic/internal/model"
	"golang-basic/internal/repository"
	"regexp"
	"strings"
	"unicode/utf8"
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

func (s *ProfileService) validateProfileInput(p model.Profile) error {
	type fieldError struct {
		Field string
		Msg   string
	}
	var errs []fieldError

	// Validate age
	if p.Age < 0 || p.Age > 120 {
		errs = append(errs, fieldError{"Age", "must be between 0 and 120"})
	}

	// Validate gender
	gender := strings.ToLower(strings.TrimSpace(p.Gender))
	if p.Gender != "" && gender != "male" && gender != "female" && gender != "other" {
		errs = append(errs, fieldError{"Gender", "must be 'male', 'female', or 'other'"})
	}

	bioLen := utf8.RuneCountInString(p.Bio)
	if bioLen > 500 {
		errs = append(errs, fieldError{"Bio", "maximum length is 500 characters"})
	}

	if p.PhoneNumber != "" && !phoneRegex.MatchString(p.PhoneNumber) {
		errs = append(errs, fieldError{"PhoneNumber", "invalid format (expected 10-15 digits, optionally starting with +)"})
	}

	if p.Website != "" && !websiteRegex.MatchString(p.Website) {
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

	if err := s.validateProfileInput(model.Profile{
		UserID:      req.UserID,
		Age:         req.Age,
		Gender:      req.Gender,
		Bio:         req.Bio,
		PhoneNumber: req.PhoneNumber,
		Website:     req.Website,
	}); err != nil {
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
	if req.Age != 0 {
		merged.Age = req.Age
	}
	if req.Gender != "" {
		merged.Gender = req.Gender
	}
	if req.Bio != "" {
		merged.Bio = req.Bio
	}
	if req.PhoneNumber != "" {
		merged.PhoneNumber = req.PhoneNumber
	}
	if req.Website != "" {
		merged.Website = req.Website
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

func (s *ProfileService) DeleteProfile(ctx context.Context, profileId int64) error {
	return s.repo.Delete(ctx, profileId)
}
