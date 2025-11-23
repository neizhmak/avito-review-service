package domain

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

// Team represents a group of users working together.
type Team struct {
	Name    string `json:"team_name"`
	Members []User `json:"members,omitempty"`
}

// User represents an individual user in the system.
type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name"`
}

// PullRequest represents a pull request in the system.
type PullRequest struct {
	ID        string     `json:"pull_request_id"`
	Title     string     `json:"pull_request_name"`
	AuthorID  string     `json:"author_id"`
	Status    PRStatus   `json:"status"`
	Reviewers []string   `json:"assigned_reviewers"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	MergedAt  *time.Time `json:"mergedAt,omitempty"`
}

// ReviewerStats represents statistics for a reviewer.
type ReviewerStats struct {
	ReviewerID string `json:"reviewer_id"`
	Count      int    `json:"review_count"`
}

// SystemStats represents overall system statistics.
type SystemStats struct {
	TotalPRs     int             `json:"total_prs"`
	TopReviewers []ReviewerStats `json:"top_reviewers"`
}

// PullRequestShort represents a summarized view of a pull request.
type PullRequestShort struct {
	ID       string   `json:"pull_request_id"`
	Title    string   `json:"pull_request_name"`
	AuthorID string   `json:"author_id"`
	Status   PRStatus `json:"status"`
}
