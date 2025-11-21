package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/neizhmak/avito-review-service/internal/domain"
)

type PullRequestStorage struct {
	db *sql.DB
}

type QueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func NewPullRequestStorage(db *sql.DB) *PullRequestStorage {
	return &PullRequestStorage{db: db}
}

// Save saves a new pr to the database.
func (s *PullRequestStorage) Save(ctx context.Context, executor QueryExecutor, pr domain.PullRequest) error {
	query := "INSERT INTO pull_requests (id, title, author_id, status) VALUES ($1, $2, $3, $4)"

	_, err := executor.ExecContext(ctx, query, pr.ID, pr.Title, pr.AuthorID, pr.Status)
	if err != nil {
		return fmt.Errorf("failed to insert pr: %w", err)
	}

	return nil
}

func (s *PullRequestStorage) SaveReviewer(ctx context.Context, executor QueryExecutor, prID, reviewerID string) error {
	query := "INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)"
	_, err := executor.ExecContext(ctx, query, prID, reviewerID)
	return err
}
