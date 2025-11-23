package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/storage"
)

// OpenTestDB opens PostgreSQL connection for tests using TEST_DB_CONNECTION_STRING or default DSN.
func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()

	connStr := os.Getenv("TEST_DB_CONNECTION_STRING")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	if err = db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v. Make sure Docker is running!", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

// CleanupTeamData removes test data for a team to keep tests isolated.
func CleanupTeamData(t *testing.T, db *sql.DB, teamName string) {
	t.Helper()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "DELETE FROM pr_reviewers WHERE pull_request_id IN (SELECT id FROM pull_requests WHERE author_id IN (SELECT id FROM users WHERE team_name = $1))", teamName); err != nil {
		t.Fatalf("failed to cleanup pr_reviewers: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM pull_requests WHERE author_id IN (SELECT id FROM users WHERE team_name = $1)", teamName); err != nil {
		t.Fatalf("failed to cleanup pull_requests: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM users WHERE team_name = $1", teamName); err != nil {
		t.Fatalf("failed to cleanup users: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM teams WHERE name = $1", teamName); err != nil {
		t.Fatalf("failed to cleanup teams: %v", err)
	}
}

// SeedTeam inserts a team and users using provided storages to avoid import cycles.
type TeamSaver interface {
	Save(ctx context.Context, team domain.Team) error
}

type UserSaver interface {
	Save(ctx context.Context, user domain.User) error
}

type PRSaver interface {
	Save(ctx context.Context, executor storage.QueryExecutor, pr domain.PullRequest) error
	SaveReviewer(ctx context.Context, executor storage.QueryExecutor, prID, reviewerID string) error
}

func SeedTeam(t *testing.T, teamStorage TeamSaver, userStorage UserSaver, teamName string, users []domain.User) {
	t.Helper()
	ctx := context.Background()

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team %s: %v", teamName, err)
	}
	for _, u := range users {
		u.TeamName = teamName
		if err := userStorage.Save(ctx, u); err != nil {
			t.Fatalf("failed to save user %s: %v", u.ID, err)
		}
	}
}

// SeedPR inserts a pull request with optional reviewers using provided repository.
func SeedPR(t *testing.T, prStorage PRSaver, executor storage.QueryExecutor, pr domain.PullRequest, reviewers ...string) {
	t.Helper()
	ctx := context.Background()

	if pr.Status == "" {
		pr.Status = domain.PRStatusOpen
	}

	if err := prStorage.Save(ctx, executor, pr); err != nil {
		t.Fatalf("failed to save pr %s: %v", pr.ID, err)
	}
	for _, r := range reviewers {
		if err := prStorage.SaveReviewer(ctx, executor, pr.ID, r); err != nil {
			t.Fatalf("failed to save reviewer %s: %v", r, err)
		}
	}
}
