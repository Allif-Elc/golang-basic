package controller

import (
	"context"
	"encoding/json"
	"golang-basic/internal/model"
	"golang-basic/internal/service"
	"golang-basic/internal/utility"
	"net/http"
	"strconv"
	"strings"
)

type ProfileServiceInterface interface {
	CreateProfile(ctx context.Context, req model.CreateProfileRequest) (*model.Profile, error)
	UpdateProfile(ctx context.Context, req model.UpdateProfileRequest) (*model.Profile, error)
	GetAllProfiles(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error)
	GetProfileByID(ctx context.Context, id int64) (model.Profile, error)
	DeleteProfile(ctx context.Context, profileId int64) error
}

type ProfileController struct {
	service ProfileServiceInterface
}

func NewProfileController(service *service.ProfileService) *ProfileController {
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
	if r.Method != http.MethodPut {
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

	path := strings.TrimPrefix(r.URL.Path, "/profiles/")
	id, err := strconv.ParseInt(path, 10, 64)
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

func (c *ProfileController) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/profiles/")
	id, err := strconv.ParseInt(path, 10, 64)
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

func (c *ProfileController) HandleProfileRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path == "/profiles" || r.URL.Path == "/profiles/" {
		if r.Method == http.MethodPost {
			c.CreateProfile(w, r)
		} else if r.Method == http.MethodGet {
			c.GetAllProfiles(w, r)
		} else if r.Method == http.MethodPut {
			c.UpdateProfile(w, r)
		} else {
			utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	} else if strings.HasPrefix(r.URL.Path, "/profiles/") {
		if r.Method == http.MethodGet {
			c.GetProfileByID(w, r)
		} else if r.Method == http.MethodDelete {
			c.DeleteProfile(w, r)
		} else {
			utility.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	} else {
		utility.SendError(w, http.StatusNotFound, "Route not found")
	}
}
