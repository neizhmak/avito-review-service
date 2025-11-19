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
	defer db.Close()

	// Check connections
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v. Make sure Docker is running!", err)
	}

	storage := NewTeamStorage(db)
	ctx := context.Background()

	teamName := "test-team-1"

	// Clear DB
	_, _ = db.Exec("DELETE FROM teams WHERE name = $1", teamName)

	// Test create record
	err = storage.Save(ctx, domain.Team{Name: teamName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test exist record
	var nameInDB string
	err = db.QueryRow("SELECT name FROM teams WHERE name = $1", teamName).Scan(&nameInDB)
	if err != nil {
		t.Fatalf("failed to find created team: %v", err)
	}

	if nameInDB != teamName {
		t.Errorf("want %s, got %s", teamName, nameInDB)
	}
}
