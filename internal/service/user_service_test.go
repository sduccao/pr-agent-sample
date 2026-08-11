package service_test

import (
	"context"
	"testing"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/benchmark/go-ai-review-benchmark/internal/repository"
	"github.com/benchmark/go-ai-review-benchmark/internal/service"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}
	return db
}

func TestUserService_RegisterAndRetrieve(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.InitSchema(ctx); err != nil {
		t.Fatalf("failed init schema: %v", err)
	}

	userService := service.NewUserService(userRepo)

	req := &model.CreateUserRequest{
		Username: "gopher",
		Email:    "gopher@golang.org",
		Role:     "ADMIN",
	}

	user, err := userService.RegisterUser(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error registering user: %v", err)
	}

	if user.ID == 0 {
		t.Errorf("expected non-zero ID, got %d", user.ID)
	}

	retrieved, err := userService.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching user: %v", err)
	}

	if retrieved.Username != "gopher" {
		t.Errorf("expected username gopher, got %s", retrieved.Username)
	}
}

func TestUserService_GetUserRole_NilSafety(t *testing.T) {
	userService := service.NewUserService(nil)

	if role := userService.GetUserRole(nil); role != "ANONYMOUS" {
		t.Errorf("expected ANONYMOUS for nil user, got %s", role)
	}

	userWithoutProfile := &model.User{ID: 1, Username: "test"}
	if role := userService.GetUserRole(userWithoutProfile); role != "DEFAULT_ROLE" {
		t.Errorf("expected DEFAULT_ROLE when profile is nil, got %s", role)
	}
}
