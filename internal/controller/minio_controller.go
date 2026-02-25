package controller

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang-basic/api/internal/service"
)

type MinioController struct {
	minioService *service.MinioService
}

func NewMinioController(minioService *service.MinioService) *MinioController {
	return &MinioController{minioService: minioService}
}

func (c *MinioController) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	object := chi.URLParam(r, "object")
	if object == "" {
		http.Error(w, "object name is required", http.StatusBadRequest)
		return
	}

	url, err := c.minioService.GeneratePresignedUploadURL(object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func (c *MinioController) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	object := chi.URLParam(r, "object")
	if object == "" {
		http.Error(w, "object name is required", http.StatusBadRequest)
		return
	}

	url, err := c.minioService.GeneratePresignedDownloadURL(object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
