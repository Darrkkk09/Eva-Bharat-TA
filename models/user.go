package models

// User represents a user in the ticket system.
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	// Password is kept hidden from JSON output.
	Password string `json:"-"`
}
