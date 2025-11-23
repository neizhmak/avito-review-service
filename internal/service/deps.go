package service

import (
	"context"

	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/storage"
)

// PullRequestRepository defines persistence operations for pull requests.
type PullRequestRepository interface {
	Save(ctx context.Context, executor storage.QueryExecutor, pr domain.PullRequest) error
	GetByID(ctx context.Context, id string) (*domain.PullRequest, error)
	UpdateStatus(ctx context.Context, executor storage.QueryExecutor, id string, status domain.PRStatus) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	DeleteReviewer(ctx context.Context, executor storage.QueryExecutor, prID string, userID string) error
	SaveReviewer(ctx context.Context, executor storage.QueryExecutor, prID, reviewerID string) error
	GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
	RemoveReviewersByTeam(ctx context.Context, executor storage.QueryExecutor, teamName string) error
	GetSystemStats(ctx context.Context) (*domain.SystemStats, error)
}

// UserRepository defines persistence operations for users.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	Save(ctx context.Context, user domain.User) error
	UpdateActivity(ctx context.Context, userID string, isActive bool) error
	GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	MassDeactivate(ctx context.Context, executor storage.QueryExecutor, teamName string) error
}

// TeamRepository defines persistence operations for teams.
type TeamRepository interface {
	GetByName(ctx context.Context, name string) (*domain.Team, error)
	Save(ctx context.Context, team domain.Team) error
}
