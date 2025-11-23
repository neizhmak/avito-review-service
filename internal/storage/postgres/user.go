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
	defer func() {
		_ = rows.Close()
	}()

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

// GetUsersByTeam retrieves all users belonging to a specific team.
func (s *UserStorage) GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	query := "SELECT id, username, is_active, team_name FROM users WHERE team_name = $1"
	rows, err := s.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive, &u.TeamName); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateActivity updates the activity status of a user.
func (s *UserStorage) UpdateActivity(ctx context.Context, userID string, isActive bool) error {
	query := "UPDATE users SET is_active = $1 WHERE id = $2"
	res, err := s.db.ExecContext(ctx, query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user activity: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// MassDeactivate sets is_active to false for all users in the specified team.
func (s *UserStorage) MassDeactivate(ctx context.Context, executor QueryExecutor, teamName string) error {
	query := "UPDATE users SET is_active = false WHERE team_name = $1"
	_, err := executor.ExecContext(ctx, query, teamName)
	if err != nil {
		return fmt.Errorf("failed to deactivate users: %w", err)
	}
	return nil
}
