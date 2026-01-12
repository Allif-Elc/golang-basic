package controller

import (
	"context"
	"encoding/json"
	"golang-basic/internal/model"
	"golang-basic/internal/utility"
	"net/http"
	"strconv"
	"strings"
)

type UserServiceInterface interface {
	CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	UpdateUser(ctx context.Context, req model.UpdateUserRequest) error
	GetAllUsers(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error)
	GetUserByID(ctx context.Context, id int64) (model.User, error)
}

type UserController struct {
	service UserServiceInterface
}

func NewUserController(service UserServiceInterface) *UserController {
	return &UserController{service: service}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := c.service.CreateUser(r.Context(), req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "User created successfully", user)
}

func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := c.service.UpdateUser(r.Context(), req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User updated successfully", nil)
}

func (c *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse pagination parameters from query string
	page := 1
	size := 10
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	// Parse cursor parameter
	lastCursor := int64(0)
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		if c, err := strconv.ParseInt(cursorStr, 10, 64); err == nil {
			lastCursor = c
		}
	}

	pageReq := model.PageRequest{
		Page: page,
		Size: size,
	}

	result, err := c.service.GetAllUsers(r.Context(), pageReq, lastCursor)
	if err != nil {
		utility.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Users retrieved successfully", result.Data)
}

func (c *UserController) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := c.service.GetUserByID(r.Context(), id)
	if err != nil {
		utility.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "User retrieved successfully", user)
}

func (c *UserController) HandleUserRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path == "/users" || r.URL.Path == "/users/" {
		if r.Method == http.MethodPost {
			c.CreateUser(w, r)
		} else if r.Method == http.MethodGet {
			c.GetAllUsers(w, r)
		} else if r.Method == http.MethodPut {
			c.UpdateUser(w, r)
		} else {
			utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	} else if strings.HasPrefix(r.URL.Path, "/users/") {
		c.GetUserByID(w, r)
	} else {
		utility.SendError(w, http.StatusNotFound, "Route not found")
	}
}
