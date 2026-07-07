package store

import (
	"errors"
	"sync"
	"time"

	"go-evabharat/models"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailConflict  = errors.New("email already registered")
	ErrTicketNotFound = errors.New("ticket not found")
)

// This structure implements models real database operations.
type MemoryStore struct {
	usersMu    sync.RWMutex
	users      []models.User
	nextUserID int

	ticketsMu    sync.RWMutex
	tickets      []models.Ticket
	nextTicketID int
}

// NewMemoryStore initializes a new in-memory.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make([]models.User, 0),
		nextUserID:   1,
		tickets:      make([]models.Ticket, 0),
		nextTicketID: 1,
	}
}

// CreateUser registers a new user if the email is not already taken.
func (s *MemoryStore) CreateUser(username, email, password string) (models.User, error) {
	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	// Check email uniqueness
	for _, u := range s.users {
		if u.Email == email {
			return models.User{}, ErrEmailConflict
		}
	}

	user := models.User{
		ID:       s.nextUserID,
		Username: username,
		Email:    email,
		Password: password,
	}
	s.users = append(s.users, user)
	s.nextUserID++

	return user, nil
}

// GetUserByEmail finds a user by their registered email.
func (s *MemoryStore) GetUserByEmail(email string) (models.User, error) {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return models.User{}, ErrUserNotFound
}

// GetUserByID finds a user by their unique ID.
func (s *MemoryStore) GetUserByID(id int) (models.User, error) {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return models.User{}, ErrUserNotFound
}


// CreateTicket creates a support ticket in the in-memory database.
func (s *MemoryStore) CreateTicket(title, description string, createdBy int) (models.Ticket, error) {
	s.ticketsMu.Lock()
	defer s.ticketsMu.Unlock()

	ticket := models.Ticket{
		ID:          s.nextTicketID,
		Title:       title,
		Description: description,
		Status:      "open",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.tickets = append(s.tickets, ticket)
	s.nextTicketID++

	return ticket, nil
}

// GetTicketByID retrieves a specific ticket by its ID.
func (s *MemoryStore) GetTicketByID(id int) (models.Ticket, error) {
	s.ticketsMu.RLock()
	defer s.ticketsMu.RUnlock()

	for _, t := range s.tickets {
		if t.ID == id {
			return t, nil
		}
	}
	return models.Ticket{}, ErrTicketNotFound
}

// GetTicketsByUserID retrieves all tickets created by a specific user.
func (s *MemoryStore) GetTicketsByUserID(userID int) []models.Ticket {
	s.ticketsMu.RLock()
	defer s.ticketsMu.RUnlock()

	var res []models.Ticket
	for _, t := range s.tickets {
		if t.CreatedBy == userID {
			res = append(res, t)
		}
	}
	return res
}

// UpdateTicketStatus transitions the state of a ticket following the transition rules:
// open -> in_progress -> closed. Once closed, the status cannot be changed.
func (s *MemoryStore) UpdateTicketStatus(ticketID int, userID int, newStatus string) (models.Ticket, error) {
	s.ticketsMu.Lock()
	defer s.ticketsMu.Unlock()

	// Find the ticket index
	idx := -1
	
	for i, t := range s.tickets {
		if t.ID == ticketID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return models.Ticket{}, ErrTicketNotFound
	}

	ticket := &s.tickets[idx]

	// Verify ownership: only the creator can update the status
	if ticket.CreatedBy != userID {
		return models.Ticket{}, errors.New("unauthorized to update this ticket")
	}

	// Validate status transition
	current := ticket.Status

	if current == "closed" {
		return models.Ticket{}, errors.New("closed tickets cannot be modified")
	}

	if newStatus != "in_progress" && newStatus != "closed" {
		return models.Ticket{}, errors.New("invalid status: must be 'in_progress' or 'closed'")
	}

	if current == "open" && newStatus != "in_progress" {
		return models.Ticket{}, errors.New("invalid transition: open tickets must transition to 'in_progress' first")
	}

	if current == "in_progress" && newStatus != "closed" {
		return models.Ticket{}, errors.New("invalid transition: 'in_progress' tickets can only transition to 'closed'")
	}

	// Apply transition
	ticket.Status = newStatus
	ticket.UpdatedAt = time.Now()

	return *ticket, nil
}

// ListTickets returns a copy of all stored tickets.
func (s *MemoryStore) ListTickets() []models.Ticket {
	s.ticketsMu.RLock()
	defer s.ticketsMu.RUnlock()

	// Create a safe copy of the slice to avoid race conditions when clients access/read it
	res := make([]models.Ticket, len(s.tickets))
	copy(res, s.tickets)
	return res
}
