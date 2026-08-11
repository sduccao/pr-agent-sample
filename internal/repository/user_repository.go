package repository

import (
	"context"
	"fmt"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/jmoiron/sqlx"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository constructs a UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// InitSchema creates necessary tables for user entity.
func (r *UserRepository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL,
		bio TEXT,
		role TEXT,
		preferences TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// GetByID retrieves a user by their primary key.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	var bio, role, preferences *string

	query := `SELECT id, username, email, status, bio, role, preferences, created_at FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Status, &bio, &role, &preferences, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user repository GetByID: %w", err)
	}

	if bio != nil || role != nil || preferences != nil {
		user.Profile = &model.UserProfile{}
		if bio != nil {
			user.Profile.Bio = *bio
		}
		if role != nil {
			user.Profile.Role = *role
		}
		if preferences != nil {
			user.Profile.Preferences = *preferences
		}
	}

	return &user, nil
}

// GetByStatus retrieves users by status using parameterised queries (safe against SQL injection).
func (r *UserRepository) GetByStatus(ctx context.Context, status string) ([]model.User, error) {
	query := `SELECT id, username, email, status, created_at FROM users WHERE status = ?`
	var users []model.User
	err := r.db.SelectContext(ctx, &users, query, status)
	if err != nil {
		return nil, fmt.Errorf("user repository GetByStatus: %w", err)
	}
	return users, nil
}

// Create inserts a new user record.
func (r *UserRepository) Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	query := `INSERT INTO users (username, email, status, role) VALUES (?, ?, 'ACTIVE', ?)`
	res, err := r.db.ExecContext(ctx, query, req.Username, req.Email, req.Role)
	if err != nil {
		return nil, fmt.Errorf("user repository Create: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("user repository LastInsertId: %w", err)
	}

	return r.GetByID(ctx, id)
}
