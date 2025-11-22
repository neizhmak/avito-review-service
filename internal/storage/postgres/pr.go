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

// UpdateStatus updates the status of a pull request, setting merged_at if status is MERGED.
func (s *PullRequestStorage) UpdateStatus(ctx context.Context, executor QueryExecutor, id string, status string) error {
	var query string
	if status == "MERGED" {
		query = "UPDATE pull_requests SET status = $1, merged_at = NOW() WHERE id = $2"
	} else {
		query = "UPDATE pull_requests SET status = $1 WHERE id = $2"
	}

	_, err := executor.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update pr status: %w", err)
	}
	return nil
}

// GetReviewers retrieves the list of reviewer IDs for a given pull request.
func (s *PullRequestStorage) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	query := "SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1"
	rows, err := s.db.QueryContext(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewers: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// DeleteReviewer removes a reviewer from a pull request.
func (s *PullRequestStorage) DeleteReviewer(ctx context.Context, executor QueryExecutor, prID, reviewerID string) error {
	query := "DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2"
	res, err := executor.ExecContext(ctx, query, prID, reviewerID)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("reviewer not found on this PR")
	}
	return nil
}
