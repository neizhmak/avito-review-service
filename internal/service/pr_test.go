package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/storage/postgres"
)

func TestPRService_Create(t *testing.T) {
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
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)

	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "service-test-team"
	authorID := "u-author"
	reviewer1ID := "u-rev-1"
	reviewer2ID := "u-rev-2"
	prID := "pr-service-1"

	db.Exec("DELETE FROM pr_reviewers WHERE pull_request_id = $1", prID)
	db.Exec("DELETE FROM pull_requests WHERE id = $1", prID)
	db.Exec("DELETE FROM users WHERE team_name = $1", teamName)
	db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	team := domain.Team{Name: teamName}
	if err := teamStorage.Save(ctx, team); err != nil {
		t.Fatalf("prep failed: %v", err)
	}

	users := []domain.User{
		{ID: authorID, Username: "Author", IsActive: true, TeamName: teamName},
		{ID: reviewer1ID, Username: "Rev1", IsActive: true, TeamName: teamName},
		{ID: reviewer2ID, Username: "Rev2", IsActive: true, TeamName: teamName},
	}

	for _, u := range users {
		if err := userStorage.Save(ctx, u); err != nil {
			t.Fatalf("prep failed saving user %s: %v", u.ID, err)
		}
	}

	reqPR := domain.PullRequest{
		ID:       prID,
		Title:    "Service Test PR",
		AuthorID: authorID,
		Status:   "OPEN",
	}

	createdPR, err := service.Create(ctx, reqPR)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify created PR
	if createdPR.ID != prID {
		t.Errorf("want pr id %s, got %s", prID, createdPR.ID)
	}

	if len(createdPR.Reviewers) != 2 {
		t.Errorf("want 2 reviewers, got %d: %v", len(createdPR.Reviewers), createdPR.Reviewers)
	}

	dbPR, err := prStorage.GetByID(ctx, prID)
	if err != nil {
		t.Fatalf("failed to get PR from DB: %v", err)
	}
	if dbPR.Status != "OPEN" {
		t.Errorf("want OPEN in DB, got %s", dbPR.Status)
	}
}

func TestPRService_Merge(t *testing.T) {
	connStr := "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("conn failed: %v", err)
	}
	defer db.Close()

	// Check connections
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v. Make sure Docker is running!", err)
	}

	// Init
	prStorage := postgres.NewPullRequestStorage(db)
	userStorage := postgres.NewUserStorage(db)
	teamStorage := postgres.NewTeamStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "merge-team"
	authorID := "merge-author"
	prID := "merge-pr-1"

	// Сlear DB
	db.Exec("DELETE FROM pull_requests WHERE id = $1", prID)
	db.Exec("DELETE FROM users WHERE id = $1", authorID)
	db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	teamStorage.Save(ctx, domain.Team{Name: teamName})
	userStorage.Save(ctx, domain.User{ID: authorID, Username: "Auth", IsActive: true, TeamName: teamName})

	originalPR := domain.PullRequest{ID: prID, Title: "Merge Me", AuthorID: authorID, Status: "OPEN"}
	prStorage.Save(ctx, db, originalPR)

	// Test merge
	mergedPR, err := service.Merge(ctx, prID)
	if err != nil {
		t.Fatalf("first merge failed: %v", err)
	}

	// Verify merged PR
	if mergedPR.Status != "MERGED" {
		t.Errorf("want MERGED, got %s", mergedPR.Status)
	}
	if mergedPR.MergedAt == nil {
		t.Error("want MergedAt not nil")
	}

	// Test idempotent merge
	mergedPR2, err := service.Merge(ctx, prID)
	if err != nil {
		t.Fatalf("second merge failed: %v", err)
	}

	// Verify idempotent merged PR
	if mergedPR2.Status != "MERGED" {
		t.Errorf("want MERGED, got %s", mergedPR2.Status)
	}
}
