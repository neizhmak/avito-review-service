package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/neizhmak/avito-review-service/internal/domain"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

// Save saves a new user to the database.
func (s *UserStorage) Save(ctx context.Context, user domain.User) error {
	query := "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)"

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Username, user.IsActive, user.TeamName)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}
