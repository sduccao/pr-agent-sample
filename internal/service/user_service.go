package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/benchmark/go-ai-review-benchmark/internal/repository"
)

// UserService handles business logic for managing users.
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService constructs a UserService.
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// RegisterUser validates input and creates a new user.
func (s *UserService) RegisterUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	if strings.TrimSpace(req.Username) == "" {
		return nil, errors.New("username cannot be empty")
	}
	if !strings.Contains(req.Email, "@") {
		return nil, errors.New("invalid email address format")
	}
	if req.Role == "" {
		req.Role = "USER"
	}

	user, err := s.userRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("user service RegisterUser failed: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves user profile safely checking nil references.
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service GetUserByID failed: %w", err)
	}
	return user, nil
}

// GetUserRole returns the role assigned to a user profile safely checking for nil.
func (s *UserService) GetUserRole(user *model.User) string {
	if user == nil {
		return "ANONYMOUS"
	}
	if user.Profile == nil {
		return "DEFAULT_ROLE"
	}
	return user.Profile.Role
}

// ListUsersByStatus fetches users with matching status.
func (s *UserService) ListUsersByStatus(ctx context.Context, status string) ([]model.User, error) {
	if status == "" {
		status = "ACTIVE"
	}
	return s.userRepo.GetByStatus(ctx, status)
}
