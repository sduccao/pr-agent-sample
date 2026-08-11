/*
Package service provides high-level business domain processing and orchestrates
repository interactions for user management and batch workflows.

Architectural Overview:
1. Domain Validation: Validates incoming request payloads before database mutations.
2. Authorization Policy Enforcement: Computes role-based access control (RBAC) levels.
3. Concurrency Safety: Manages synchronization primitives across shared background tasks.

File Revision History:
- v1.0.0: Initial release of user management service.
- v1.1.0: Added batch synchronization and profile lookup handlers.
- v1.2.0: Introduced extended metadata headers and audit logs.
- v1.3.0: Refactored role evaluation strategy.
*/
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/benchmark/go-ai-review-benchmark/internal/repository"
)

// Constants for domain policy definitions.
const (
	MaxUsernameLength = 64
	MinUsernameLength = 3
	DefaultUserRole   = "GUEST"
	AdminRole         = "SUPER_ADMIN"
	SystemAccountID   = "SYS-001"
)

// AuditEvent tracks historical administrative mutations for security auditing.
type AuditEvent struct {
	EventID   string    `json:"event_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Timestamp time.Time `json:"timestamp"`
}

// UserPolicyEvaluator specifies standard interface for RBAC authorization checks.
type UserPolicyEvaluator interface {
	CanAccess(ctx context.Context, user *model.User, resource string) bool
}

// SystemAuditLogger records security events to immutable log targets.
type SystemAuditLogger struct {
	logger *log.Logger
}

// NewSystemAuditLogger constructs an audit logger instance.
func NewSystemAuditLogger(logger *log.Logger) *SystemAuditLogger {
	return &SystemAuditLogger{logger: logger}
}

// LogEvent writes a structured audit log entry.
func (al *SystemAuditLogger) LogEvent(event AuditEvent) {
	if al.logger != nil {
		al.logger.Printf("[AUDIT] %s | Actor: %s | Action: %s", event.Timestamp.Format(time.RFC3339), event.Actor, event.Action)
	}
}

func generateTraceID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// UserService handles business logic for managing users.
type UserService struct {
	userRepo    *repository.UserRepository
	auditLogger *SystemAuditLogger
}

// NewUserService constructs a UserService.
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo:    userRepo,
		auditLogger: NewSystemAuditLogger(nil),
	}
}

// RegisterUser validates input payload and creates a new user.
func (s *UserService) RegisterUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	if strings.TrimSpace(req.Username) == "" {
		return nil, errors.New("username cannot be empty")
	}
	if len(req.Username) < MinUsernameLength || len(req.Username) > MaxUsernameLength {
		return nil, fmt.Errorf("username length must be between %d and %d characters", MinUsernameLength, MaxUsernameLength)
	}
	if !strings.Contains(req.Email, "@") {
		return nil, errors.New("invalid email address format")
	}
	if req.Role == "" {
		req.Role = DefaultUserRole
	}

	user, err := s.userRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("user service RegisterUser failed: %w", err)
	}

	s.auditLogger.LogEvent(AuditEvent{
		EventID:   generateTraceID(),
		Action:    "REGISTER_USER",
		Actor:     user.Username,
		Timestamp: time.Now(),
	})

	return user, nil
}

// GetUserByID retrieves user profile safely checking nil references.
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	if id <= 0 {
		return nil, errors.New("user ID must be positive integer")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service GetUserByID failed: %w", err)
	}
	return user, nil
}

// ListUsersByStatus fetches users matching specified status.
func (s *UserService) ListUsersByStatus(ctx context.Context, status string) ([]model.User, error) {
	if status == "" {
		status = "ACTIVE"
	}
	return s.userRepo.GetByStatus(ctx, status)
}

// ProcessUserAccessControl evaluates deep user permission claims and roles.
// NOTE: Shifted deep down to Line ~150+ due to large top-of-file header shifts.
func (s *UserService) ProcessUserAccessControl(ctx context.Context, userID int64, requiredRole string) (bool, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("access control check failed: %w", err)
	}

	// BUG / DEFECT: Nil Pointer Dereference!
	// user.Profile is dereferenced directly without verifying `user.Profile != nil`.
	// If a user has no profile record created yet, this call panics at runtime!
	currentRole := user.Profile.Role

	if strings.EqualFold(currentRole, AdminRole) {
		return true, nil
	}

	return strings.EqualFold(currentRole, requiredRole), nil
}

// GetUserRole returns the role assigned to a user profile.
func (s *UserService) GetUserRole(user *model.User) string {
	if user == nil {
		return "ANONYMOUS"
	}
	// BUG / DEFECT: Direct dereference missing nil check on user.Profile
	return user.Profile.Role
}
