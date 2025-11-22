package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/neizhmak/avito-review-service/internal/domain"
)

type TeamStorage struct {
	db *sql.DB
}

func NewTeamStorage(db *sql.DB) *TeamStorage {
	return &TeamStorage{db: db}
}

// Save saves a new team to the database.
func (s *TeamStorage) Save(ctx context.Context, team domain.Team) error {
	query := "INSERT INTO teams (name) VALUES ($1)"

	_, err := s.db.ExecContext(ctx, query, team.Name)
	if err != nil {
		return fmt.Errorf("failed to insert team: %w", err)
	}

	return nil
}

// GetByName retrieves a team by its name.
func (s *TeamStorage) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	query := "SELECT name FROM teams WHERE name = $1"

	row := s.db.QueryRowContext(ctx, query, name)

	var t domain.Team
	if err := row.Scan(&t.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &t, nil
}
