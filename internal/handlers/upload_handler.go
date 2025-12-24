package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/internal/storage"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// UploadHandler handles file upload endpoints
type UploadHandler struct {
	storage     *storage.MinIOClient
	userService *services.UserService
	rbwService  *services.RBWService
	cfg         *config.Config
}

func NewUploadHandler(storage *storage.MinIOClient, userService *services.UserService, rbwService *services.RBWService, cfg *config.Config) *UploadHandler {
	return &UploadHandler{
		storage:     storage,
		userService: userService,
		rbwService:  rbwService,
		cfg:         cfg,
	}
}

// UploadAvatar handles POST /uploads/avatar
func (h *UploadHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())

	if h.storage == nil {
		response.InternalError(w, "Storage service not available")
		return
	}

	// Parse multipart form
	maxSize := int64(h.cfg.MaxAvatarSizeMB) * 1024 * 1024
	if err := r.ParseMultipartForm(maxSize); err != nil {
		response.BadRequest(w, fmt.Sprintf("File too large, max %dMB", h.cfg.MaxAvatarSizeMB))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	if !isAllowedImageFormat(ext, h.cfg.AllowedImageFormats) {
		response.BadRequest(w, "Invalid file format")
		return
	}

	// Generate object name
	objectName := storage.GenerateObjectName("avatars", fmt.Sprintf("%s.%s", claims.UserID, ext))

	// Upload to MinIO
	url, err := h.storage.Upload(r.Context(), objectName, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		response.InternalError(w, "Failed to upload file")
		return
	}

	// Update user avatar
	_, err = h.userService.Update(r.Context(), claims.UserID, &models.UpdateUserRequest{AvatarURL: &url})
	if err != nil {
		response.InternalError(w, "Failed to update avatar")
		return
	}

	response.Success(w, "Avatar uploaded", map[string]string{"url": url})
}

// UploadRBWPhoto handles POST /uploads/rbw/{rbw_id}/photo
func (h *UploadHandler) UploadRBWPhoto(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")

	if h.storage == nil {
		response.InternalError(w, "Storage service not available")
		return
	}

	// Parse multipart form
	maxSize := int64(h.cfg.MaxUploadSizeMB) * 1024 * 1024
	if err := r.ParseMultipartForm(maxSize); err != nil {
		response.BadRequest(w, fmt.Sprintf("File too large, max %dMB", h.cfg.MaxUploadSizeMB))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	if !isAllowedImageFormat(ext, h.cfg.AllowedImageFormats) {
		response.BadRequest(w, "Invalid file format")
		return
	}

	// Generate object name
	objectName := storage.GenerateObjectName("rbw", fmt.Sprintf("%s.%s", rbwID, ext))

	// Upload to MinIO
	url, err := h.storage.Upload(r.Context(), objectName, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		response.InternalError(w, "Failed to upload file")
		return
	}

	// Update RBW photo
	_, err = h.rbwService.Update(r.Context(), rbwID, &models.UpdateRBWRequest{PhotoURL: &url})
	if err != nil {
		response.InternalError(w, "Failed to update RBW photo")
		return
	}

	response.Success(w, "Photo uploaded", map[string]string{"url": url})
}

func isAllowedImageFormat(ext string, allowed []string) bool {
	for _, a := range allowed {
		if ext == a {
			return true
		}
	}
	return false
}

// Suppress unused
var _ = io.Reader(nil)
