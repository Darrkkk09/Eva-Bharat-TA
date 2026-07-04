package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go-evabharat/store"
)

// AuthHandler handles authentication routes.
type AuthHandler struct {
	Store *store.MemoryStore
}

// NewAuthHandler initializes a new AuthHandler with store dependency.
func NewAuthHandler(s *store.MemoryStore) *AuthHandler {
	return &AuthHandler{Store: s}
}

// RegisterRequest holds registration request body.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest holds login credentials.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse defines response format for successful logins.
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}

	user, err := h.Store.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, store.ErrEmailConflict) {
			RespondWithError(w, http.StatusConflict, err.Error())
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, user)
}

// Login handles user login and issues dummy JWT tokens.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.Store.GetUserByEmail(req.Email)
	if err != nil || user.Password != req.Password {
		RespondWithError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Generate a mock JWT string
	dummyToken := fmt.Sprintf("dummy-jwt-token-for-user-%d", user.ID)

	RespondWithJSON(w, http.StatusOK, LoginResponse{
		Token: dummyToken,
		User:  user,
	})
}
