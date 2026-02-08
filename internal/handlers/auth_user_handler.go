package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
	"github.com/swiftlead/backend-swiftlet/pkg/validator"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	token, user, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			response.Unauthorized(w, "Invalid email or password")
		} else {
			response.InternalError(w, "Failed to login")
		}
		return
	}

	response.Success(w, "Login successful", map[string]interface{}{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// Register handles POST /auth/register (admin only)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			response.Conflict(w, "User with this email already exists")
		} else {
			response.InternalError(w, "Failed to register user")
		}
		return
	}

	response.Created(w, "User registered successfully", user.ToResponse())
}

// PublicRegister handles POST /auth/register/public (open registration as farmer)
func (h *AuthHandler) PublicRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	user, err := h.userService.PublicRegister(r.Context(), &req)
	if err != nil {
		switch err {
		case repository.ErrUserAlreadyExists:
			response.Conflict(w, "User with this email already exists")
		case services.ErrWeakPassword:
			response.BadRequest(w, "Password must contain uppercase, lowercase, digit, and special character")
		default:
			response.InternalError(w, "Failed to register user")
		}
		return
	}

	response.Created(w, "Registration successful", user.ToResponse())
}

// ChangePassword handles POST /auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	err := h.userService.ChangePassword(r.Context(), claims.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			response.Unauthorized(w, "Old password is incorrect")
		case services.ErrSamePassword:
			response.BadRequest(w, "New password must be different from old password")
		case services.ErrWeakPassword:
			response.BadRequest(w, "Password must contain uppercase, lowercase, digit, and special character")
		default:
			response.InternalError(w, "Failed to change password")
		}
		return
	}

	response.Success(w, "Password changed successfully", nil)
}

// ForgotPassword handles POST /auth/forgot-password (admin resets user password)
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Generate a cryptographically random temporary password
	tempPassword, err := generateTempPassword(12)
	if err != nil {
		response.InternalError(w, "Failed to generate temporary password")
		return
	}
	err = h.userService.AdminResetPassword(r.Context(), req.Email, tempPassword)
	if err != nil {
		if err == repository.ErrUserNotFound {
			response.NotFound(w, "User not found")
		} else {
			response.InternalError(w, "Failed to reset password")
		}
		return
	}

	response.Success(w, "Password has been reset", map[string]string{
		"temporary_password": tempPassword,
		"message":            "Please change the password after login",
	})
}

// UserHandler handles user endpoints
type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMe handles GET /users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		response.InternalError(w, "Failed to get user")
		return
	}

	response.Success(w, "", user.ToResponse())
}

// UpdateMe handles PATCH /users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	user, err := h.userService.Update(r.Context(), claims.UserID, &req)
	if err != nil {
		response.InternalError(w, "Failed to update user")
		return
	}

	response.Success(w, "Profile updated", user.ToResponse())
}

// List handles GET /users (admin only)
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 20
	}

	users, total, err := h.userService.List(r.Context(), role, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list users")
		return
	}

	var userResponses []*models.UserResponse
	for _, u := range users {
		userResponses = append(userResponses, u.ToResponse())
	}

	response.SuccessWithMeta(w, userResponses, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// Create handles POST /users (admin only)
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			response.Conflict(w, "User with this email already exists")
		} else {
			response.InternalError(w, "Failed to create user")
		}
		return
	}

	response.Created(w, "User created", user.ToResponse())
}

// Helper to get pagination params
func getPagination(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 20
	}
	return
}

// generateTempPassword creates a cryptographically random password
// with at least one uppercase, lowercase, digit, and special character.
func generateTempPassword(length int) (string, error) {
	const (
		upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowerChars   = "abcdefghijklmnopqrstuvwxyz"
		digitChars   = "0123456789"
		specialChars = "!@#$%^&*"
		allChars     = upperChars + lowerChars + digitChars + specialChars
	)

	if length < 8 {
		length = 8
	}

	// Ensure at least one of each required type
	password := make([]byte, 0, length)
	for _, charset := range []string{upperChars, lowerChars, digitChars, specialChars} {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("crypto/rand failed: %w", err)
		}
		password = append(password, charset[idx.Int64()])
	}

	// Fill the rest with random chars from the full set
	for i := len(password); i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", fmt.Errorf("crypto/rand failed: %w", err)
		}
		password = append(password, allChars[idx.Int64()])
	}

	// Shuffle to avoid predictable positions
	for i := len(password) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("crypto/rand failed: %w", err)
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
}
