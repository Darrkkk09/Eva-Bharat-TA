package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-evabharat/models"
	"go-evabharat/store"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// UserContextKey is the key used to store/retrieve the User object in/from Request Context.
const UserContextKey contextKey = "user"

func AuthMiddleware(s *store.MemoryStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				RespondWithError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				RespondWithError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenStr := parts[1]
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				RespondWithError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				RespondWithError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			user, err := s.GetUserByID(userID)
			if err != nil {
				RespondWithError(w, http.StatusUnauthorized, "authenticated user no longer exists")
				return
			}

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
