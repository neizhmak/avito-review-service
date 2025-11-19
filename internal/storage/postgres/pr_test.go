package postgres

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
)

func TestPullRequestStorage_Save(t *testing.T) {
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
	teamStorage := NewTeamStorage(db)
	userStorage := NewUserStorage(db)
	prStorage := NewPullRequestStorage(db)
	ctx := context.Background()

	teamName := "pr-team"
	authorID := "pr-author"
	prID := "pr-1"

	// Clear DB
	db.Exec("DELETE FROM pull_requests WHERE id = $1", prID)
	db.Exec("DELETE FROM users WHERE id = $1", authorID)
	db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}

	author := domain.User{ID: authorID, Username: "Dev", IsActive: true, TeamName: teamName}
	if err := userStorage.Save(ctx, author); err != nil {
		t.Fatalf("failed to save author: %v", err)
	}

	pr := domain.PullRequest{
		ID:       prID,
		Title:    "Fix database",
		AuthorID: authorID,
		Status:   "OPEN",
	}

	// Test create record
	if err = prStorage.Save(ctx, pr); err != nil {
		t.Fatalf("failed to save PR: %v", err)
	}

	// Test exist record
	var status string
	if err = db.QueryRow("SELECT status FROM pull_requests WHERE id = $1", prID).Scan(&status); err != nil {
		t.Fatalf("failed to find PR: %v", err)
	}
	if status != "OPEN" {
		t.Errorf("want OPEN, got %s", status)
	}
}
