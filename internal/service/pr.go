package service

import (
	"context"
	"database/sql"
	"errors"
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
	pr.Status = "OPEN"

	// Validate author
	author, err := s.userStorage.GetByID(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, notFound("author not found")
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	// Check duplicates
	existing, getErr := s.prStorage.GetByID(ctx, pr.ID)
	if getErr == nil && existing != nil {
		return nil, conflict(ErrCodePRExists, "PR id already exists")
	} else if getErr != nil && !errors.Is(getErr, postgres.ErrNotFound) {
		return nil, fmt.Errorf("failed to check PR existence: %w", getErr)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err = s.prStorage.Save(ctx, tx, pr); err != nil {
		return nil, fmt.Errorf("failed to save pr: %w", err)
	}

	candidates, err := s.userStorage.GetActiveUsersByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	reviewers := selectRandomReviewers(candidates, pr.AuthorID, 2)

	for _, r := range reviewers {
		if err = s.prStorage.SaveReviewer(ctx, tx, pr.ID, r.ID); err != nil {
			return nil, fmt.Errorf("failed to save reviewer: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
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
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, notFound("pr not found")
		}
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

// Reassign replaces an existing reviewer on a pull request with a new one from the same team.
func (s *PRService) Reassign(ctx context.Context, prID, oldUserID string) (string, error) {
	pr, err := s.prStorage.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return "", notFound("pr not found")
		}
		return "", fmt.Errorf("pr not found: %w", err)
	}
	if pr.Status == "MERGED" {
		return "", conflict(ErrCodePRMerged, "cannot reassign on merged PR")
	}

	currentReviewers, err := s.prStorage.GetReviewers(ctx, prID)
	if err != nil {
		return "", err
	}
	isAssigned := false
	for _, id := range currentReviewers {
		if id == oldUserID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return "", conflict(ErrCodeNotAssigned, "reviewer is not assigned to this PR")
	}

	oldUser, err := s.userStorage.GetByID(ctx, oldUserID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return "", notFound("user not found")
		}
		return "", fmt.Errorf("failed to get old reviewer info: %w", err)
	}

	candidates, err := s.userStorage.GetActiveUsersByTeam(ctx, oldUser.TeamName)
	if err != nil {
		return "", err
	}

	var validCandidates []domain.User
	for _, cand := range candidates {
		if cand.ID == pr.AuthorID {
			continue
		}
		if cand.ID == oldUserID {
			continue
		}

		alreadyReviewing := false
		for _, rID := range currentReviewers {
			if rID == cand.ID {
				alreadyReviewing = true
				break
			}
		}
		if alreadyReviewing {
			continue
		}

		validCandidates = append(validCandidates, cand)
	}

	if len(validCandidates) == 0 {
		return "", conflict(ErrCodeNoCandidate, "no active replacement candidate in team")
	}

	newReviewer := selectRandomReviewers(validCandidates, "", 1)[0]

	// transactional update
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err = s.prStorage.DeleteReviewer(ctx, tx, prID, oldUserID); err != nil {
		return "", err
	}
	if err = s.prStorage.SaveReviewer(ctx, tx, prID, newReviewer.ID); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return newReviewer.ID, nil
}

// CreateTeam creates a new team along with its members.
func (s *PRService) CreateTeam(ctx context.Context, team domain.Team) (*domain.Team, error) {
	if _, err := s.teamStorage.GetByName(ctx, team.Name); err == nil {
		return nil, newServiceError(ErrCodeTeamExists, "team_name already exists")
	} else if err != nil && !errors.Is(err, postgres.ErrNotFound) {
		return nil, err
	}

	if err := s.teamStorage.Save(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to save team: %w", err)
	}

	for i, u := range team.Members {
		u.TeamName = team.Name
		if err := s.userStorage.Save(ctx, u); err != nil {
			return nil, fmt.Errorf("failed to save user %s: %w", u.ID, err)
		}
		team.Members[i] = u
	}

	return &team, nil
}

// GetPR retrieves a pull request by its ID.
func (s *PRService) GetPR(ctx context.Context, id string) (*domain.PullRequest, error) {
	pr, err := s.prStorage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, notFound("pr not found")
		}
		return nil, err
	}
	return pr, nil
}

// GetTeam retrieves a team by its name along with its members.
func (s *PRService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamStorage.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, notFound("team not found")
		}
		return nil, err
	}

	members, err := s.userStorage.GetUsersByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}
	team.Members = members

	return team, nil
}

// SetUserActive sets the active status of a user.
func (s *PRService) SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	if err := s.userStorage.UpdateActivity(ctx, userID, isActive); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, notFound("user not found")
		}
		return nil, err
	}

	return s.userStorage.GetByID(ctx, userID)
}

// GetUserReviews retrieves all pull requests assigned to a specific reviewer.
func (s *PRService) GetUserReviews(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	if _, err := s.userStorage.GetByID(ctx, reviewerID); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, notFound("user not found")
		}
		return nil, err
	}

	prs, err := s.prStorage.GetByReviewerID(ctx, reviewerID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		result = append(result, domain.PullRequestShort{
			ID:       pr.ID,
			Title:    pr.Title,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	return result, nil
}

// DeactivateTeam deactivates all users in a team and removes them from open pull requests.
func (s *PRService) DeactivateTeam(ctx context.Context, teamName string) error {
	_, err := s.teamStorage.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return notFound("team not found")
		}
		return err
	}

	// Transactional operation
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.userStorage.MassDeactivate(ctx, tx, teamName); err != nil {
		return err
	}

	if err := s.prStorage.RemoveReviewersByTeam(ctx, tx, teamName); err != nil {
		return err
	}

	return tx.Commit()
}

// GetStats retrieves system statistics related to pull requests.
func (s *PRService) GetStats(ctx context.Context) (*domain.SystemStats, error) {
	return s.prStorage.GetSystemStats(ctx)
}
