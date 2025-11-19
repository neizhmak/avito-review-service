package domain

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
