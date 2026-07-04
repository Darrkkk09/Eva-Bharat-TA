package models

import "time"

// Ticket represents a support ticket in the system.
type Ticket struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // e.g., "open", "in-progress", "closed"
	CreatedBy   int       `json:"created_by"`
	AssignedTo  int       `json:"assigned_to,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
