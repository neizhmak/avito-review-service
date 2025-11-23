package postgres

import (
	"context"
	"testing"

	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/testutil"
)

func TestUserStorage_UpdateAndMassDeactivate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	ctx := context.Background()

	teamStorage := NewTeamStorage(db)
	userStorage := NewUserStorage(db)
	var err error

	teamName := "storage-user"
	userActive := "storage-user-active"
	userPassive := "storage-user-passive"

	testutil.CleanupTeamData(t, db, teamName)

	if err = teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}

	users := []domain.User{
		{ID: userActive, Username: "Active", IsActive: true, TeamName: teamName},
		{ID: userPassive, Username: "Passive", IsActive: false, TeamName: teamName},
	}
	for _, u := range users {
		if err = userStorage.Save(ctx, u); err != nil {
			t.Fatalf("failed to save user %s: %v", u.ID, err)
		}
	}

	if err = userStorage.UpdateActivity(ctx, userActive, false); err != nil {
		t.Fatalf("failed to update activity: %v", err)
	}

	var isActive bool
	if err = db.QueryRowContext(ctx, "SELECT is_active FROM users WHERE id = $1", userActive).Scan(&isActive); err != nil {
		t.Fatalf("failed to query user: %v", err)
	}
	if isActive {
		t.Fatalf("expected user to be inactive after update")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to start tx: %v", err)
	}
	if err = userStorage.MassDeactivate(ctx, tx, teamName); err != nil {
		t.Fatalf("failed to mass deactivate: %v", err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	usersByTeam, err := userStorage.GetUsersByTeam(ctx, teamName)
	if err != nil {
		t.Fatalf("failed to get users by team: %v", err)
	}
	if len(usersByTeam) != 2 {
		t.Fatalf("expected 2 users, got %d", len(usersByTeam))
	}
	for _, u := range usersByTeam {
		if u.IsActive {
			t.Fatalf("expected all users inactive after mass deactivate")
		}
	}
}
