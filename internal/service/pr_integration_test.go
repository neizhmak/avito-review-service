package service

import (
	"context"
	"errors"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/storage/postgres"
	"github.com/neizhmak/avito-review-service/internal/testutil"
)

func TestPRService_Create_Happy(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "service-test-team"
	authorID := "u-author"
	reviewer1ID := "u-rev-1"
	reviewer2ID := "u-rev-2"
	prID := "pr-service-1"

	testutil.CleanupTeamData(t, db, teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("prep failed: %v", err)
	}

	users := []domain.User{
		{ID: authorID, Username: "Author", IsActive: true, TeamName: teamName},
		{ID: reviewer1ID, Username: "Rev1", IsActive: true, TeamName: teamName},
		{ID: reviewer2ID, Username: "Rev2", IsActive: true, TeamName: teamName},
	}

	for _, u := range users {
		if err := userStorage.Save(ctx, u); err != nil {
			t.Fatalf("prep failed saving user %s: %v", u.ID, err)
		}
	}

	createdPR, err := service.Create(ctx, domain.PullRequest{
		ID:       prID,
		Title:    "Service Test PR",
		AuthorID: authorID,
		Status:   domain.PRStatusOpen,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if createdPR.ID != prID {
		t.Errorf("want pr id %s, got %s", prID, createdPR.ID)
	}
	if len(createdPR.Reviewers) != 2 {
		t.Errorf("want 2 reviewers, got %d: %v", len(createdPR.Reviewers), createdPR.Reviewers)
	}
}

func TestPRService_Merge_Happy(t *testing.T) {
	db := testutil.OpenTestDB(t)

	prStorage := postgres.NewPullRequestStorage(db)
	userStorage := postgres.NewUserStorage(db)
	teamStorage := postgres.NewTeamStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "merge-team"
	authorID := "merge-author"
	prID := "merge-pr-1"

	testutil.CleanupTeamData(t, db, teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}
	if err := userStorage.Save(ctx, domain.User{ID: authorID, Username: "Auth", IsActive: true, TeamName: teamName}); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	originalPR := domain.PullRequest{ID: prID, Title: "Merge Me", AuthorID: authorID, Status: domain.PRStatusOpen}
	if err := prStorage.Save(ctx, db, originalPR); err != nil {
		t.Fatalf("failed to save pr: %v", err)
	}

	mergedPR, err := service.Merge(ctx, prID)
	if err != nil {
		t.Fatalf("first merge failed: %v", err)
	}
	if mergedPR.Status != domain.PRStatusMerged {
		t.Errorf("want MERGED, got %s", mergedPR.Status)
	}
	if mergedPR.MergedAt == nil {
		t.Error("want MergedAt not nil")
	}
}

func TestPRService_Reassign_Happy(t *testing.T) {
	db := testutil.OpenTestDB(t)

	prStorage := postgres.NewPullRequestStorage(db)
	userStorage := postgres.NewUserStorage(db)
	teamStorage := postgres.NewTeamStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "reassign-team"
	authorID := "reassign-author"
	oldReviewerID := "reassign-old-rev"
	newReviewerID := "reassign-new-rev"
	prID := "reassign-pr-1"

	testutil.CleanupTeamData(t, db, teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}
	if err := userStorage.Save(ctx, domain.User{ID: authorID, Username: "Auth", IsActive: true, TeamName: teamName}); err != nil {
		t.Fatalf("failed to save author: %v", err)
	}
	if err := userStorage.Save(ctx, domain.User{ID: oldReviewerID, Username: "OldRev", IsActive: true, TeamName: teamName}); err != nil {
		t.Fatalf("failed to save old reviewer: %v", err)
	}
	if err := userStorage.Save(ctx, domain.User{ID: newReviewerID, Username: "NewRev", IsActive: true, TeamName: teamName}); err != nil {
		t.Fatalf("failed to save new reviewer: %v", err)
	}

	originalPR := domain.PullRequest{ID: prID, Title: "Reassign Me", AuthorID: authorID, Status: domain.PRStatusOpen}
	if err := prStorage.Save(ctx, db, originalPR); err != nil {
		t.Fatalf("failed to save pr: %v", err)
	}
	if err := prStorage.SaveReviewer(ctx, db, prID, oldReviewerID); err != nil {
		t.Fatalf("failed to save reviewer: %v", err)
	}

	newRevID, err := service.Reassign(ctx, prID, oldReviewerID)
	if err != nil {
		t.Fatalf("reassign failed: %v", err)
	}
	if newRevID == oldReviewerID {
		t.Fatalf("expected new reviewer, got same")
	}
}

func TestPRService_Create_DuplicatePR(t *testing.T) {
	db := testutil.OpenTestDB(t)

	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "dup-team"
	prID := "dup-pr"
	authorID := "dup-author"

	testutil.CleanupTeamData(t, db, teamName)

	testutil.SeedTeam(t, teamStorage, userStorage, teamName, []domain.User{
		{ID: authorID, Username: "Author", IsActive: true},
	})

	existing := domain.PullRequest{ID: prID, Title: "Existing", AuthorID: authorID, Status: domain.PRStatusOpen}
	if err := prStorage.Save(ctx, db, existing); err != nil {
		t.Fatalf("failed to seed pr: %v", err)
	}

	_, err := service.Create(ctx, domain.PullRequest{ID: prID, Title: "Duplicate", AuthorID: authorID})
	if err == nil {
		t.Fatalf("expected duplicate error, got nil")
	}

	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodePRExists {
		t.Fatalf("expected service error %s, got %v", ErrCodePRExists, err)
	}
}

func TestPRService_Reassign_NotAssigned(t *testing.T) {
	db := testutil.OpenTestDB(t)

	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "reassign-negative"
	authorID := "neg-author"
	assignedReviewer := "neg-assigned"
	unassigned := "neg-unassigned"
	prID := "neg-pr"

	testutil.CleanupTeamData(t, db, teamName)

	testutil.SeedTeam(t, teamStorage, userStorage, teamName, []domain.User{
		{ID: authorID, Username: "Author", IsActive: true},
		{ID: assignedReviewer, Username: "Reviewer", IsActive: true},
		{ID: unassigned, Username: "Free", IsActive: true},
	})

	pr := domain.PullRequest{ID: prID, Title: "PR", AuthorID: authorID, Status: domain.PRStatusOpen}
	testutil.SeedPR(t, prStorage, db, pr, assignedReviewer)

	_, err := service.Reassign(ctx, prID, unassigned)
	if err == nil {
		t.Fatalf("expected not assigned error, got nil")
	}

	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodeNotAssigned {
		t.Fatalf("expected %s, got %v", ErrCodeNotAssigned, err)
	}
}

func TestPRService_Merge_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)

	prStorage := postgres.NewPullRequestStorage(db)
	userStorage := postgres.NewUserStorage(db)
	teamStorage := postgres.NewTeamStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)

	_, err := service.Merge(context.Background(), "missing-pr")
	if err == nil {
		t.Fatalf("expected not found error, got nil")
	}

	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodeNotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestPRService_DeactivateTeam_RemovesReviewers(t *testing.T) {
	db := testutil.OpenTestDB(t)

	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "deactivate-team"
	authorID := "deactivate-author"
	reviewerID := "deactivate-reviewer"
	prID := "deactivate-pr"

	testutil.CleanupTeamData(t, db, teamName)

	testutil.SeedTeam(t, teamStorage, userStorage, teamName, []domain.User{
		{ID: authorID, Username: "Author", IsActive: true},
		{ID: reviewerID, Username: "Reviewer", IsActive: true},
	})
	testutil.SeedPR(t, prStorage, db, domain.PullRequest{ID: prID, Title: "Deactivate", AuthorID: authorID}, reviewerID)

	if err := service.DeactivateTeam(ctx, teamName); err != nil {
		t.Fatalf("deactivate team failed: %v", err)
	}

	var isActive bool
	if err := db.QueryRowContext(ctx, "SELECT is_active FROM users WHERE id = $1", reviewerID).Scan(&isActive); err != nil {
		t.Fatalf("failed to query user: %v", err)
	}
	if isActive {
		t.Fatalf("expected user to be deactivated")
	}

	var reviewerCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pr_reviewers WHERE pull_request_id = $1", prID).Scan(&reviewerCount); err != nil {
		t.Fatalf("failed to query reviewers: %v", err)
	}
	if reviewerCount != 0 {
		t.Fatalf("expected reviewers removed, got %d", reviewerCount)
	}
}

func TestPRService_GetTeamAndUserReviews(t *testing.T) {
	db := testutil.OpenTestDB(t)

	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "reviews-team"
	authorID := "reviews-author"
	reviewerID := "reviews-reviewer"
	prID := "reviews-pr"

	testutil.CleanupTeamData(t, db, teamName)

	testutil.SeedTeam(t, teamStorage, userStorage, teamName, []domain.User{
		{ID: authorID, Username: "Author", IsActive: true},
		{ID: reviewerID, Username: "Reviewer", IsActive: true},
	})
	testutil.SeedPR(t, prStorage, db, domain.PullRequest{ID: prID, Title: "PR", AuthorID: authorID}, reviewerID)

	team, err := service.GetTeam(ctx, teamName)
	if err != nil {
		t.Fatalf("GetTeam failed: %v", err)
	}
	if len(team.Members) != 2 {
		t.Fatalf("expected 2 team members, got %d", len(team.Members))
	}

	reviews, err := service.GetUserReviews(ctx, reviewerID)
	if err != nil {
		t.Fatalf("GetUserReviews failed: %v", err)
	}
	if len(reviews) != 1 || reviews[0].ID != prID || reviews[0].Status != domain.PRStatusOpen {
		t.Fatalf("unexpected reviews result: %+v", reviews)
	}
}

func TestPRService_SetUserActiveAndStats(t *testing.T) {
	db := testutil.OpenTestDB(t)

	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "stats-team"
	authorID := "stats-author"
	reviewerID := "stats-reviewer"
	prID := "stats-pr"

	testutil.CleanupTeamData(t, db, teamName)

	testutil.SeedTeam(t, teamStorage, userStorage, teamName, []domain.User{
		{ID: authorID, Username: "Author", IsActive: true},
		{ID: reviewerID, Username: "Reviewer", IsActive: true},
	})
	testutil.SeedPR(t, prStorage, db, domain.PullRequest{ID: prID, Title: "PR", AuthorID: authorID}, reviewerID)

	updatedUser, err := service.SetUserActive(ctx, reviewerID, false)
	if err != nil {
		t.Fatalf("SetUserActive failed: %v", err)
	}
	if updatedUser.IsActive {
		t.Fatalf("expected user to be inactive")
	}

	stats, err := service.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.TotalPRs == 0 {
		t.Fatalf("expected TotalPRs > 0, got %d", stats.TotalPRs)
	}
}

func TestPRService_CreateTeam_Duplicate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "dup-team-service"
	testutil.CleanupTeamData(t, db, teamName)

	if _, err := service.CreateTeam(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("first create team failed: %v", err)
	}
	_, err := service.CreateTeam(ctx, domain.Team{Name: teamName})
	if err == nil {
		t.Fatalf("expected duplicate team error")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodeTeamExists {
		t.Fatalf("expected ErrCodeTeamExists, got %v", err)
	}
}

func TestPRService_Create_NoAuthor(t *testing.T) {
	db := testutil.OpenTestDB(t)
	service := NewPRService(
		postgres.NewPullRequestStorage(db),
		postgres.NewUserStorage(db),
		postgres.NewTeamStorage(db),
		db,
	)

	_, err := service.Create(context.Background(), domain.PullRequest{ID: "no-author", Title: "PR"})
	if err == nil {
		t.Fatalf("expected author not found error")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodeNotFound {
		t.Fatalf("expected ErrCodeNotFound, got %v", err)
	}
}

func TestPRService_CreateTeam_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "create-team-success"
	testutil.CleanupTeamData(t, db, teamName)

	team, err := service.CreateTeam(ctx, domain.Team{
		Name: teamName,
		Members: []domain.User{
			{ID: "cts-1", Username: "One", IsActive: true},
			{ID: "cts-2", Username: "Two", IsActive: false},
		},
	})
	if err != nil {
		t.Fatalf("CreateTeam failed: %v", err)
	}
	if team.Name != teamName || len(team.Members) != 2 {
		t.Fatalf("unexpected team result: %+v", team)
	}
}

func TestPRService_Reassign_NoCandidate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "no-candidate"
	authorID := "nc-author"
	reviewerID := "nc-reviewer"
	prID := "nc-pr"

	testutil.CleanupTeamData(t, db, teamName)

	testutil.SeedTeam(t, teamStorage, userStorage, teamName, []domain.User{
		{ID: authorID, Username: "Author", IsActive: true},
		{ID: reviewerID, Username: "Reviewer", IsActive: true},
	})

	pr := domain.PullRequest{ID: prID, Title: "No Candidate", AuthorID: authorID, Status: domain.PRStatusOpen}
	testutil.SeedPR(t, prStorage, db, pr, reviewerID)

	_, err := service.Reassign(ctx, prID, reviewerID)
	if err == nil {
		t.Fatalf("expected no candidate error")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodeNoCandidate {
		t.Fatalf("expected ErrCodeNoCandidate, got %v", err)
	}
}

func TestPRService_Reassign_OnMerged(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "merged-reassign"
	authorID := "mr-author"
	reviewerID := "mr-reviewer"
	prID := "mr-pr"

	testutil.CleanupTeamData(t, db, teamName)

	testutil.SeedTeam(t, teamStorage, userStorage, teamName, []domain.User{
		{ID: authorID, Username: "Author", IsActive: true},
		{ID: reviewerID, Username: "Reviewer", IsActive: true},
	})

	pr := domain.PullRequest{ID: prID, Title: "PR", AuthorID: authorID, Status: domain.PRStatusMerged}
	testutil.SeedPR(t, prStorage, db, pr, reviewerID)

	_, err := service.Reassign(ctx, prID, reviewerID)
	if err == nil {
		t.Fatalf("expected merged error")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodePRMerged {
		t.Fatalf("expected ErrCodePRMerged, got %v", err)
	}
}

func TestPRService_GetPR(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)
	service := NewPRService(prStorage, userStorage, teamStorage, db)
	ctx := context.Background()

	teamName := "getpr-team"
	prID := "getpr-id"
	authorID := "getpr-author"

	testutil.CleanupTeamData(t, db, teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}
	if err := userStorage.Save(ctx, domain.User{ID: authorID, Username: "Author", IsActive: true, TeamName: teamName}); err != nil {
		t.Fatalf("failed to save author: %v", err)
	}

	pr := domain.PullRequest{ID: prID, Title: "Title", AuthorID: authorID, Status: domain.PRStatusOpen}
	testutil.SeedPR(t, prStorage, db, pr)

	got, err := service.GetPR(ctx, prID)
	if err != nil {
		t.Fatalf("GetPR failed: %v", err)
	}
	if got.ID != prID {
		t.Fatalf("expected id %s, got %s", prID, got.ID)
	}
}

func TestPRService_GetTeam_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	service := NewPRService(
		postgres.NewPullRequestStorage(db),
		postgres.NewUserStorage(db),
		postgres.NewTeamStorage(db),
		db,
	)

	_, err := service.GetTeam(context.Background(), "missing-team")
	if err == nil {
		t.Fatalf("expected error")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodeNotFound {
		t.Fatalf("expected ErrCodeNotFound, got %v", err)
	}
}

func TestPRService_DeactivateTeam_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	service := NewPRService(
		postgres.NewPullRequestStorage(db),
		postgres.NewUserStorage(db),
		postgres.NewTeamStorage(db),
		db,
	)
	err := service.DeactivateTeam(context.Background(), "missing-team")
	if err == nil {
		t.Fatalf("expected error")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != ErrCodeNotFound {
		t.Fatalf("expected ErrCodeNotFound, got %v", err)
	}
}
