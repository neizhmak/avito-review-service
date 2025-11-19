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

	userStorage := NewUserStorage(db)
	teamStorage := NewTeamStorage(db)
	ctx := context.Background()

	userID := "test-user-1"
	teamName := "test-team-users"

	// Clear DB
	_, _ = db.Exec("DELETE FROM users WHERE ID = $1", userID)
	_, _ = db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	err = teamStorage.Save(ctx, domain.Team{Name: teamName})
	if err != nil {
		t.Fatalf("unexpected error saving team: %v", err)
	}

	user := domain.User{
		ID:       userID,
		Username: "TestUser",
		IsActive: true,
		TeamName: teamName,
	}

	// Test create record
	err = userStorage.Save(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	// Test exist record
	var savedName string
	err = db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&savedName)
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if savedName != user.Username {
		t.Errorf("want %s, got %s", user.Username, savedName)
	}
}
