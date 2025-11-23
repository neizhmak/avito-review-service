package domain

import "time"

// Team represents a team in the system.
type Team struct {
	Name    string `json:"team_name"`
	Members []User `json:"members,omitempty"`
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
	Reviewers []string   `json:"assigned_reviewers"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	MergedAt  *time.Time `json:"mergedAt,omitempty"`
}

type ReviewerStats struct {
	ReviewerID string `json:"reviewer_id"`
	Count      int    `json:"review_count"`
}

type SystemStats struct {
	TotalPRs     int             `json:"total_prs"`
	TopReviewers []ReviewerStats `json:"top_reviewers"`
}

type PullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Title    string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}
