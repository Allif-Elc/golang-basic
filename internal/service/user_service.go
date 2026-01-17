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
)

var nameRegex = regexp.MustCompile(`^[A-Za-z ]+$`)
var addressRegex = regexp.MustCompile(`^[A-Za-z ]+$`)
var emailRegex = regexp.MustCompile(`^[A-Za-z]+(?:\.[A-Za-z]+)*@[A-Za-z]+(?:\.[A-Za-z]+)+$`)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
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
		return err
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
