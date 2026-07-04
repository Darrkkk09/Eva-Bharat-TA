package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"go-evabharat/models"
	"go-evabharat/store"
)

type contextKey string

// UserContextKey is the key used to store/retrieve the User object in/from Request Context.
const UserContextKey contextKey = "user"

// AuthMiddleware extracts the Bearer token, validates it, and loads the user into context.
func AuthMiddleware(s *store.MemoryStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				RespondWithError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			// Expecting "Bearer <token>" format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				RespondWithError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			token := parts[1]
			prefix := "dummy-jwt-token-for-user-"
			if !strings.HasPrefix(token, prefix) {
				RespondWithError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Extract User ID from the dummy token
			idStr := strings.TrimPrefix(token, prefix)
			userID, err := strconv.Atoi(idStr)
			if err != nil {
				RespondWithError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			// Load user from store
			user, err := s.GetUserByID(userID)
			if err != nil {
				RespondWithError(w, http.StatusUnauthorized, "authenticated user no longer exists")
				return
			}

			// Inject user into context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthenticatedUser retrieves the injected User object from request context.
func GetAuthenticatedUser(r *http.Request) (models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(models.User)
	return user, ok
}
