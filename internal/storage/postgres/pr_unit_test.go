package postgres

import (
	"context"
	"testing"

	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/testutil"
)

func TestPullRequestStorage_ReviewersAndStats(t *testing.T) {
	db := testutil.OpenTestDB(t)
	ctx := context.Background()

	teamStorage := NewTeamStorage(db)
	userStorage := NewUserStorage(db)
	prStorage := NewPullRequestStorage(db)

	teamName := "storage-pr"
	authorID := "storage-author"
	reviewer1 := "storage-rev1"
	reviewer2 := "storage-rev2"
	prID := "storage-pr-1"

	testutil.CleanupTeamData(t, db, teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}
	users := []domain.User{
		{ID: authorID, Username: "Author", IsActive: true, TeamName: teamName},
		{ID: reviewer1, Username: "Rev1", IsActive: true, TeamName: teamName},
		{ID: reviewer2, Username: "Rev2", IsActive: true, TeamName: teamName},
	}
	for _, u := range users {
		if err := userStorage.Save(ctx, u); err != nil {
			t.Fatalf("failed to save user %s: %v", u.ID, err)
		}
	}

	pr := domain.PullRequest{ID: prID, Title: "Test", AuthorID: authorID, Status: domain.PRStatusOpen}
	if err := prStorage.Save(ctx, db, pr); err != nil {
		t.Fatalf("failed to save pr: %v", err)
	}

	if err := prStorage.SaveReviewer(ctx, db, prID, reviewer1); err != nil {
		t.Fatalf("failed to save reviewer1: %v", err)
	}
	if err := prStorage.SaveReviewer(ctx, db, prID, reviewer2); err != nil {
		t.Fatalf("failed to save reviewer2: %v", err)
	}

	reviewers, err := prStorage.GetReviewers(ctx, prID)
	if err != nil {
		t.Fatalf("failed to get reviewers: %v", err)
	}
	if len(reviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(reviewers))
	}

	if err = prStorage.DeleteReviewer(ctx, db, prID, reviewer1); err != nil {
		t.Fatalf("failed to delete reviewer: %v", err)
	}
	reviewers, err = prStorage.GetReviewers(ctx, prID)
	if err != nil {
		t.Fatalf("failed to get reviewers after delete: %v", err)
	}
	if len(reviewers) != 1 {
		t.Fatalf("expected 1 reviewer after delete, got %d", len(reviewers))
	}

	if err = prStorage.RemoveReviewersByTeam(ctx, db, teamName); err != nil {
		t.Fatalf("failed to remove reviewers by team: %v", err)
	}
	reviewers, err = prStorage.GetReviewers(ctx, prID)
	if err != nil {
		t.Fatalf("failed to get reviewers after team removal: %v", err)
	}
	if len(reviewers) != 0 {
		t.Fatalf("expected 0 reviewers after removal, got %d", len(reviewers))
	}

	stats, err := prStorage.GetSystemStats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalPRs == 0 {
		t.Fatalf("expected stats to include PRs")
	}
}

func TestPullRequestStorage_UpdateStatusAndGetByReviewer(t *testing.T) {
	db := testutil.OpenTestDB(t)
	ctx := context.Background()

	teamStorage := NewTeamStorage(db)
	userStorage := NewUserStorage(db)
	prStorage := NewPullRequestStorage(db)

	teamName := "storage-status"
	authorID := "status-author"
	reviewerID := "status-reviewer"
	prID := "status-pr"

	testutil.CleanupTeamData(t, db, teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}
	if err := userStorage.Save(ctx, domain.User{ID: authorID, Username: "Author", IsActive: true, TeamName: teamName}); err != nil {
		t.Fatalf("failed to save author: %v", err)
	}
	if err := userStorage.Save(ctx, domain.User{ID: reviewerID, Username: "Reviewer", IsActive: true, TeamName: teamName}); err != nil {
		t.Fatalf("failed to save reviewer: %v", err)
	}

	pr := domain.PullRequest{ID: prID, Title: "Status PR", AuthorID: authorID, Status: domain.PRStatusOpen}
	if err := prStorage.Save(ctx, db, pr); err != nil {
		t.Fatalf("failed to save pr: %v", err)
	}
	if err := prStorage.SaveReviewer(ctx, db, prID, reviewerID); err != nil {
		t.Fatalf("failed to save reviewer: %v", err)
	}

	if err := prStorage.UpdateStatus(ctx, db, prID, domain.PRStatusMerged); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	updated, err := prStorage.GetByID(ctx, prID)
	if err != nil {
		t.Fatalf("failed to get updated pr: %v", err)
	}
	if updated.Status != domain.PRStatusMerged || updated.MergedAt == nil {
		t.Fatalf("expected merged status with timestamp, got %+v", updated)
	}

	byReviewer, err := prStorage.GetByReviewerID(ctx, reviewerID)
	if err != nil {
		t.Fatalf("failed to get by reviewer: %v", err)
	}
	if len(byReviewer) != 1 || byReviewer[0].ID != prID {
		t.Fatalf("expected PR by reviewer, got %+v", byReviewer)
	}
}
