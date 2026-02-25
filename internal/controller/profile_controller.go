package controller

import (
	"context"
	"encoding/json"
	"golang-basic/api/internal/model"
	"golang-basic/api/internal/utility"
	"net/http"
	"strconv"
	"strings"

	"golang-basic/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type ProfileServiceInterface interface {
	CreateProfile(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error)
	UpdateProfile(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error)
	GetAllProfiles(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error)
	GetProfileByID(ctx context.Context, id int64) (model.Profile, error)
	GetProfileByUserID(ctx context.Context, userID int64) (model.Profile, error)
	DeleteProfile(ctx context.Context, profileId int64) error
}

type ProfileController struct {
	service ProfileServiceInterface
}

func NewProfileController(service ProfileServiceInterface) *ProfileController {
	return &ProfileController{service: service}
}

func (c *ProfileController) CreateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	profile, err := c.service.CreateProfile(ctx, req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusCreated, "Profile created successfully", profile)
}

func (c *ProfileController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	profile, err := c.service.UpdateProfile(ctx, req)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Profile updated successfully", profile)
}

func (c *ProfileController) GetAllProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse pagination parameters from query string
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)

	if page < 1 {
		page = 1
	}
	if size == 0 {
		size = 10
	}

	req := model.PageRequest{
		Page: page,
		Size: size,
	}

	ctx := r.Context()
	profiles, err := c.service.GetAllProfiles(ctx, req, cursor)
	if err != nil {
		utility.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Profiles retrieved successfully", profiles)
}

func (c *ProfileController) GetProfileByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Try to get ID from chi URL param first, fallback to path parsing
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		// Parse from URL path for unit testing (e.g., "/profiles/123")
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) > 0 {
			idStr = pathParts[len(pathParts)-1]
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	ctx := r.Context()
	profile, err := c.service.GetProfileByID(ctx, id)
	if err != nil {
		utility.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Profile retrieved successfully", profile)
}

func (c *ProfileController) GetProfileByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userID int64
	var err error

	// Try to get userId from query parameter first
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr != "" {
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			utility.SendError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
	} else {
		// Fallback to context (for backward compatibility - fix the key to use middleware.UserIDKey)
		userIDValue := r.Context().Value(middleware.UserIDKey)
		if userIDValue == nil {
			utility.SendError(w, http.StatusUnauthorized, "User not authenticated")
			return
		}
		var ok bool
		userID, ok = userIDValue.(int64)
		if !ok {
			utility.SendError(w, http.StatusUnauthorized, "Invalid user ID")
			return
		}
	}

	ctx := r.Context()
	profile, err := c.service.GetProfileByUserID(ctx, userID)
	if err != nil {
		utility.SendError(w, http.StatusNotFound, "Profile not found")
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Profile retrieved successfully", profile)
}

func (c *ProfileController) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Try to get ID from chi URL param first, fallback to path parsing
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		// Parse from URL path for unit testing (e.g., "/profiles/123")
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) > 0 {
			idStr = pathParts[len(pathParts)-1]
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utility.SendError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	ctx := r.Context()
	if err := c.service.DeleteProfile(ctx, id); err != nil {
		utility.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Profile deleted successfully", nil)
}
