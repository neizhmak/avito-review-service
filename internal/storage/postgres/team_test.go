package postgres

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
)

func TestTeamStorage_Save(t *testing.T) {
	connStr := "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Check connections
	if err = db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v. Make sure Docker is running!", err)
	}

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
