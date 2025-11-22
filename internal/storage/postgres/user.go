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

// Gets a list of the team's active users.
func (s *UserStorage) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	query := "SELECT id, username, is_active, team_name FROM users WHERE team_name = $1 AND is_active = true"

	rows, err := s.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive, &u.TeamName); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// GetByID retrieves a user by their ID.
func (s *UserStorage) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := "SELECT id, username, is_active, team_name FROM users WHERE id = $1"
	
	row := s.db.QueryRowContext(ctx, query, userID)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.IsActive, &u.TeamName); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", userID) 
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &u, nil
}
