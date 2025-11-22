package postgres

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
)

func TestUserStorage_Save(t *testing.T) {
	connStr := "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	// Check connections
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v. Make sure Docker is running!", err)
	}

	// Init
	userStorage := NewUserStorage(db)
	teamStorage := NewTeamStorage(db)
	ctx := context.Background()

	userID := "test-user-1"
	teamName := "test-team-users"

	// Clear DB
	_, _ = db.Exec("DELETE FROM users WHERE ID = $1", userID)
	_, _ = db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	if err = teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("unexpected error saving team: %v", err)
	}

	user := domain.User{
		ID:       userID,
		Username: "TestUser",
		IsActive: true,
		TeamName: teamName,
	}

	// Test create record
	if err = userStorage.Save(ctx, user); err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	// Verify saved record
	var savedName string
	if err = db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&savedName); err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if savedName != user.Username {
		t.Errorf("want %s, got %s", user.Username, savedName)
	}
}

func TestGetActiveUsersByTeam(t *testing.T) {
	connStr := "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	// Check connections
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v. Make sure Docker is running!", err)
	}

	// Init
	userStorage := NewUserStorage(db)
	teamStorage := NewTeamStorage(db)
	ctx := context.Background()

	userID1 := "test-user-1"
	userID2 := "test-user-2"
	userID3 := "test-user-3"
	teamName := "test-team-users"

	// Clear DB
	_, _ = db.Exec("DELETE FROM users WHERE ID = $1", userID1)
	_, _ = db.Exec("DELETE FROM users WHERE ID = $1", userID2)
	_, _ = db.Exec("DELETE FROM users WHERE ID = $1", userID3)
	_, _ = db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	if err = teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("unexpected error saving team: %v", err)
	}

	user1 := domain.User{
		ID:       userID1,
		Username: "TestUser1",
		IsActive: true,
		TeamName: teamName,
	}

	user2 := domain.User{
		ID:       userID2,
		Username: "TestUser2",
		IsActive: true,
		TeamName: teamName,
	}

	user3 := domain.User{
		ID:       userID3,
		Username: "TestUser3",
		IsActive: false,
		TeamName: teamName,
	}

	if err = userStorage.Save(ctx, user1); err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	if err = userStorage.Save(ctx, user2); err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	if err = userStorage.Save(ctx, user3); err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	// Test get active users by team
	if arr, _ := userStorage.GetActiveUsersByTeam(ctx, teamName); len(arr) != 2 {
		t.Fatalf("want 2, got %d", len(arr))
	}
}

func TestUserStorage_GetByID(t *testing.T) {
	connStr := "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Check connections
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v. Make sure Docker is running!", err)
	}

	// Init
	userStorage := NewUserStorage(db)
	teamStorage := NewTeamStorage(db)
	ctx := context.Background()

	userID := "test-get-id"
	teamName := "test-team-get-id"
	
	// Сlear DB
	db.Exec("DELETE FROM users WHERE id = $1", userID)
	db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}

	// Test non-existent user
	_, err = userStorage.GetByID(ctx, userID)
	if err == nil {
		t.Error("expected error for non-existent user, got nil")
	}

	expectedUser := domain.User{ID: userID, Username: "GetByIdUser", IsActive: true, TeamName: teamName}
	if err := userStorage.Save(ctx, expectedUser); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	// Test existing user
	u, err := userStorage.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Verify user fields
	if u.ID != expectedUser.ID || u.Username != expectedUser.Username {
		t.Errorf("want %v, got %v", expectedUser, u)
	}
}
