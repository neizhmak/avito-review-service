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
