package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/storage"
)

type PullRequestStorage struct {
	db *sql.DB
}

func NewPullRequestStorage(db *sql.DB) *PullRequestStorage {
	return &PullRequestStorage{db: db}
}

// Save saves a new pr to the database.
func (s *PullRequestStorage) Save(ctx context.Context, executor storage.QueryExecutor, pr domain.PullRequest) error {
	query := "INSERT INTO pull_requests (id, title, author_id, status) VALUES ($1, $2, $3, $4)"

	_, err := executor.ExecContext(ctx, query, pr.ID, pr.Title, pr.AuthorID, pr.Status)
	if err != nil {
		return fmt.Errorf("failed to insert pr: %w", err)
	}

	return nil
}

// SaveReviewer assigns a reviewer to a pull request.
func (s *PullRequestStorage) SaveReviewer(ctx context.Context, executor storage.QueryExecutor, prID, reviewerID string) error {
	query := "INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)"
	_, err := executor.ExecContext(ctx, query, prID, reviewerID)
	return err
}

// GetByID retrieves a pull request by its ID.
func (s *PullRequestStorage) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	query := "SELECT id, title, author_id, status, created_at, merged_at FROM pull_requests WHERE id = $1"

	row := s.db.QueryRowContext(ctx, query, id)

	var pr domain.PullRequest
	var createdAt time.Time
	var mergedAt sql.NullTime
	err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: pr", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to scan pr: %w", err)
	}

	pr.CreatedAt = &createdAt
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	reviewers, err := s.GetReviewers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	pr.Reviewers = reviewers

	return &pr, nil
}

// UpdateStatus updates the status of a pull request, setting merged_at if status is MERGED.
func (s *PullRequestStorage) UpdateStatus(ctx context.Context, executor storage.QueryExecutor, id string, status domain.PRStatus) error {
	var query string
	if status == domain.PRStatusMerged {
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
	defer func() {
		_ = rows.Close()
	}()

	ids := make([]string, 0)
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
func (s *PullRequestStorage) DeleteReviewer(ctx context.Context, executor storage.QueryExecutor, prID, reviewerID string) error {
	query := "DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2"
	res, err := executor.ExecContext(ctx, query, prID, reviewerID)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("%w: reviewer not found on this PR", ErrNotFound)
	}
	return nil
}

// GetByReviewerID retrieves all pull requests assigned to a specific reviewer.
func (s *PullRequestStorage) GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	query := `
		SELECT pr.id, pr.title, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers rev ON pr.id = rev.pull_request_id
		WHERE rev.reviewer_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewer prs: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

// RemoveReviewersByTeam removes all reviewers from open pull requests for a given team.
func (s *PullRequestStorage) RemoveReviewersByTeam(ctx context.Context, executor storage.QueryExecutor, teamName string) error {
	query := `
		DELETE FROM pr_reviewers
		WHERE reviewer_id IN (
			SELECT id FROM users WHERE team_name = $1
		)
		AND pull_request_id IN (
			SELECT id FROM pull_requests WHERE status = $2
		)
	`
	_, err := executor.ExecContext(ctx, query, teamName, domain.PRStatusOpen)
	if err != nil {
		return fmt.Errorf("failed to remove team reviewers: %w", err)
	}
	return nil
}

// GetSystemStats retrieves overall system statistics.
func (s *PullRequestStorage) GetSystemStats(ctx context.Context) (*domain.SystemStats, error) {
	stats := &domain.SystemStats{}

	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pull_requests").Scan(&stats.TotalPRs)
	if err != nil {
		return nil, fmt.Errorf("failed to count prs: %w", err)
	}

	query := `
		SELECT reviewer_id, COUNT(*) as cnt
		FROM pr_reviewers
		GROUP BY reviewer_id
		ORDER BY cnt DESC
		LIMIT 5
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query top reviewers: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var r domain.ReviewerStats
		if err := rows.Scan(&r.ReviewerID, &r.Count); err != nil {
			return nil, err
		}
		stats.TopReviewers = append(stats.TopReviewers, r)
	}

	return stats, nil
}
