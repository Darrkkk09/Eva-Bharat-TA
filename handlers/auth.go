package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-evabharat/store"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "your_jwt_secret_key"
	}
	return secret
}

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

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.Store.CreateUser(req.Username, req.Email, string(hashedPassword))
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

// Login handles user login and issues JWT tokens.
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
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		RespondWithError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Generate JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.ID),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	RespondWithJSON(w, http.StatusOK, LoginResponse{
		Token: tokenString,
		User:  user,
	})
}
