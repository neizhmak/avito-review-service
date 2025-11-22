package service

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/storage/postgres"
)

type PRService struct {
	prStorage   *postgres.PullRequestStorage
	userStorage *postgres.UserStorage
	teamStorage *postgres.TeamStorage
	db          *sql.DB
}

func NewPRService(
	prStorage *postgres.PullRequestStorage,
	userStorage *postgres.UserStorage,
	teamStorage *postgres.TeamStorage,
	db *sql.DB,
) *PRService {
	return &PRService{
		prStorage:   prStorage,
		userStorage: userStorage,
		teamStorage: teamStorage,
		db:          db,
	}
}

// Create creates a new pull request and assigns reviewers.
func (s *PRService) Create(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.prStorage.Save(ctx, tx, pr); err != nil {
		return nil, fmt.Errorf("failed to save pr: %w", err)
	}

	author, err := s.userStorage.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	candidates, err := s.userStorage.GetActiveUsersByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	reviewers := selectRandomReviewers(candidates, pr.AuthorID, 2)

	for _, r := range reviewers {
		if err := s.prStorage.SaveReviewer(ctx, tx, pr.ID, r.ID); err != nil {
			return nil, fmt.Errorf("failed to save reviewer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	for _, r := range reviewers {
		pr.Reviewers = append(pr.Reviewers, r.ID)
	}

	return &pr, nil
}

// selectRandomReviewers selects a specified number of random reviewers from the candidates, excluding the author.
func selectRandomReviewers(candidates []domain.User, authorID string, count int) []domain.User {
	var valid []domain.User
	for _, u := range candidates {
		if u.ID != authorID {
			valid = append(valid, u)
		}
	}

	rand.Shuffle(len(valid), func(i, j int) { valid[i], valid[j] = valid[j], valid[i] })

	if len(valid) < count {
		return valid
	}
	return valid[:count]
}

// Merge marks a pull request as merged. The operation is idempotent.
func (s *PRService) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prStorage.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pr: %w", err)
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	if err := s.prStorage.UpdateStatus(ctx, s.db, prID, "MERGED"); err != nil {
		return nil, err
	}

	pr.Status = "MERGED"
	now := time.Now()
	pr.MergedAt = &now

	return pr, nil
}
