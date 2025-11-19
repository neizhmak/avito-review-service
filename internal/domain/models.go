package domain

import "time"

// Team represents a team in the system.
type Team struct {
	Name string `json:"team_name"`
}

// User represents a team member.
type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name"`
}

type PullRequest struct {
	ID        string     `json:"pull_request_id"`
	Title     string     `json:"pull_request_name"`
	AuthorID  string     `json:"author_id"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	MergedAt  *time.Time `json:"merged_at"` // Use a pointer (*) because the date can be NULL.
}
