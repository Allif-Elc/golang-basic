package controller

import (
	"golang-basic/api/internal/utility"
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
		utility.SendErrorResponse(w, utility.ValidationError("object name is required"))
		return
	}

	url, err := c.minioService.GeneratePresignedUploadURL(object)
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError("Failed to generate upload URL"))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Upload URL generated", map[string]string{"url": url})
}

func (c *MinioController) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	object := chi.URLParam(r, "object")
	if object == "" {
		utility.SendErrorResponse(w, utility.ValidationError("object name is required"))
		return
	}

	url, err := c.minioService.GeneratePresignedDownloadURL(object)
	if err != nil {
		utility.SendErrorResponse(w, utility.InternalError("Failed to generate download URL"))
		return
	}

	utility.SendSuccess(w, http.StatusOK, "Download URL generated", map[string]string{"url": url})
}
