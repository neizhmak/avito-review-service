package postgres

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/testutil"
)

func TestTeamStorage_Save(t *testing.T) {
	db := testutil.OpenTestDB(t)
	var err error

	// Init
	storage := NewTeamStorage(db)
	ctx := context.Background()

	teamName := "test-team-1"

	// Clear DB
	if _, err = db.Exec("DELETE FROM teams WHERE name = $1", teamName); err != nil {
		t.Fatalf("failed to cleanup teams: %v", err)
	}

	// Test create record
	if err = storage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test exist record
	var nameInDB string
	if err = db.QueryRow("SELECT name FROM teams WHERE name = $1", teamName).Scan(&nameInDB); err != nil {
		t.Fatalf("failed to find created team: %v", err)
	}

	if nameInDB != teamName {
		t.Errorf("want %s, got %s", teamName, nameInDB)
	}
}

func TestTeamStorage_GetByName(t *testing.T) {
	db := testutil.OpenTestDB(t)
	ctx := context.Background()

	teamStorage := NewTeamStorage(db)
	teamName := "team-get-by-name"

	testutil.CleanupTeamData(t, db, teamName)

	if err := teamStorage.Save(ctx, domain.Team{Name: teamName}); err != nil {
		t.Fatalf("failed to save team: %v", err)
	}

	got, err := teamStorage.GetByName(ctx, teamName)
	if err != nil {
		t.Fatalf("failed to get team: %v", err)
	}
	if got.Name != teamName {
		t.Fatalf("expected team name %s, got %s", teamName, got.Name)
	}
}
