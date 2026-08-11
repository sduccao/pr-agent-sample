package model

import "time"

// UserProfile represents nested metadata for a user.
type UserProfile struct {
	Bio         string    `json:"bio" db:"bio"`
	Role        string    `json:"role" db:"role"`
	Preferences string    `json:"preferences" db:"preferences"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a system user entity.
type User struct {
	ID        int64        `json:"id" db:"id"`
	Username  string       `json:"username" db:"username"`
	Email     string       `json:"email" db:"email"`
	Status    string       `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	Profile   *UserProfile `json:"profile,omitempty"`
}

// CreateUserRequest defines payload for user creation.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
