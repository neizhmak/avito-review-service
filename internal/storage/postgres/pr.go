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

func NewPullRequestStorage(db *sql.DB) *PullRequestStorage {
	return &PullRequestStorage{db: db}
}

// Save saves a new pr to the database.
func (s *PullRequestStorage) Save(ctx context.Context, pr domain.PullRequest) error {
	query := "INSERT INTO pull_requests (id, title, author_id, status) VALUES ($1, $2, $3, $4)"

	_, err := s.db.ExecContext(ctx, query, pr.ID, pr.Title, pr.AuthorID, pr.Status)
	if err != nil {
		return fmt.Errorf("failed to insert pr:  %w", err)
	}

	return nil
}
