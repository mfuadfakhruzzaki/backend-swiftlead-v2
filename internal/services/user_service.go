package services

import (
	"context"
	"errors"

	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSamePassword       = errors.New("new password must be different from old password")
	ErrWeakPassword       = errors.New("password does not meet strength requirements")
)

// UserService handles user business logic
type UserService struct {
	repo               repository.UserRepository
	jwtSecret          string
	jwtExpirationHours int
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, jwtSecret string, jwtExpirationHours int) *UserService {
	return &UserService{
		repo:               repo,
		jwtSecret:          jwtSecret,
		jwtExpirationHours: jwtExpirationHours,
	}
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpirationHours)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// Register creates a new user
func (s *UserService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

// Update updates a user's profile
func (s *UserService) Update(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// List retrieves users with pagination
func (s *UserService) List(ctx context.Context, role string, page, limit int) ([]*models.User, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(ctx, role, limit, offset)
}

// Delete removes a user
func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// PublicRegister creates a new user with farmer role (public registration)
func (s *UserService) PublicRegister(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, ErrWeakPassword
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         models.RoleFarmer, // Public registration always creates farmer
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !auth.CheckPassword(user.PasswordHash, oldPassword) {
		return ErrInvalidCredentials
	}

	// Check new password is different
	if auth.CheckPassword(user.PasswordHash, newPassword) {
		return ErrSamePassword
	}

	// Validate new password strength
	if err := auth.ValidatePasswordStrength(newPassword); err != nil {
		return ErrWeakPassword
	}

	// Hash and update
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, hashedPassword)
}

// AdminResetPassword resets a user's password (admin only)
func (s *UserService) AdminResetPassword(ctx context.Context, email, newPassword string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, user.ID, hashedPassword)
}
