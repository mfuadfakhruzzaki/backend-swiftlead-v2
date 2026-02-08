package services

import (
	"context"
	"testing"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	users map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.Email]; exists {
		return repository.ErrUserAlreadyExists
	}
	user.ID = "test-id"
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	if _, ok := m.users[user.Email]; ok {
		m.users[user.Email] = user
		return nil
	}
	return repository.ErrUserNotFound
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	// Not implemented for this test
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	for _, u := range m.users {
		if u.ID == userID {
			u.PasswordHash = passwordHash
			return nil
		}
	}
	return repository.ErrUserNotFound
}

func (m *MockUserRepository) List(ctx context.Context, role string, limit, offset int) ([]*models.User, int, error) {
	// Not implemented for this test
	return nil, 0, nil
}

func TestRegister(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, "secret", 1)

	req := &models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "farmer",
	}

	user, err := svc.Register(context.Background(), req)

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if user.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, user.Name)
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}

	// Verify it was saved
	saved, err := repo.GetByEmail(context.Background(), req.Email)
	if err != nil {
		t.Fatal("User not found in repo")
	}
	if saved.Email != req.Email {
		t.Error("Saved email mismatch")
	}
}

func TestRegister_Duplicate(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, "secret", 1)

	req := &models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "farmer",
	}

	svc.Register(context.Background(), req)

	// Try to register again
	_, err := svc.Register(context.Background(), req)
	if err != repository.ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}
