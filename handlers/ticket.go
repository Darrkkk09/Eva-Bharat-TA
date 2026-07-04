package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go-evabharat/models"
	"go-evabharat/store"
)

// TicketHandler handles all ticket-related routes.
type TicketHandler struct {
	Store *store.MemoryStore
}

// NewTicketHandler initializes a new TicketHandler with store dependency.
func NewTicketHandler(s *store.MemoryStore) *TicketHandler {
	return &TicketHandler{Store: s}
}

// CreateTicketRequest defines request body for ticket creation.
// Note: created_by is removed from the JSON request because it is linked
// directly to the authenticated user from the context.
type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateStatusRequest defines request body for patch status.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// Create creates a new ticket linked to the authenticated user.
func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Extract the authenticated user from context
	user, ok := GetAuthenticatedUser(r)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" || req.Description == "" {
		RespondWithError(w, http.StatusBadRequest, "title and description are required")
		return
	}

	// Link ticket to the logged-in user (user.ID)
	ticket, err := h.Store.CreateTicket(req.Title, req.Description, user.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to create ticket")
		return
	}

	RespondWithJSON(w, http.StatusCreated, ticket)
}

// List retrieves all tickets belonging to the authenticated user.
func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := GetAuthenticatedUser(r)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tickets := h.Store.GetTicketsByUserID(user.ID)
	if tickets == nil {
		tickets = make([]models.Ticket, 0)
	}

	RespondWithJSON(w, http.StatusOK, tickets)
}

// GetByID retrieves a single ticket belonging to the authenticated user.
func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := GetAuthenticatedUser(r)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// r.PathValue is a Go 1.22+ feature to parse path parameters.
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	ticket, err := h.Store.GetTicketByID(id)
	if err != nil {
		if errors.Is(err, store.ErrTicketNotFound) {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "failed to retrieve ticket")
		return
	}

	// Enforce ownership: users can only view their own tickets
	if ticket.CreatedBy != user.ID {
		RespondWithError(w, http.StatusForbidden, "unauthorized to view this ticket")
		return
	}

	RespondWithJSON(w, http.StatusOK, ticket)
}

// UpdateStatus handles PATCH /tickets/{id}/status for state transitions.
func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := GetAuthenticatedUser(r)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Status == "" {
		RespondWithError(w, http.StatusBadRequest, "status is required")
		return
	}

	updatedTicket, err := h.Store.UpdateTicketStatus(id, user.ID, req.Status)
	if err != nil {
		if errors.Is(err, store.ErrTicketNotFound) {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		// Check for permission issues vs invalid state transition issues
		if err.Error() == "unauthorized to update this ticket" {
			RespondWithError(w, http.StatusForbidden, err.Error())
			return
		}
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, updatedTicket)
}
