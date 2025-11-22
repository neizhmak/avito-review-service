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

// SaveReviewer assigns a reviewer to a pull request.
func (s *PullRequestStorage) SaveReviewer(ctx context.Context, executor QueryExecutor, prID, reviewerID string) error {
	query := "INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)"
	_, err := executor.ExecContext(ctx, query, prID, reviewerID)
	return err
}

// GetByID retrieves a pull request by its ID.
func (s *PullRequestStorage) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	query := "SELECT id, title, author_id, status, created_at, merged_at FROM pull_requests WHERE id = $1"

	row := s.db.QueryRowContext(ctx, query, id)

	var pr domain.PullRequest
	err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pr not found")
		}
		return nil, fmt.Errorf("failed to scan pr: %w", err)
	}

	return &pr, nil
}
