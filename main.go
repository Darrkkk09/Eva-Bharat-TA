package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"go-evabharat/handlers"
	"go-evabharat/store"
)

// handleHealth serves the GET /health route directly.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	handlers.RespondWithJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// SetupRouter configures the ServeMux for the ticket system.
func SetupRouter(memStore *store.MemoryStore) *http.ServeMux {
	authHandler := handlers.NewAuthHandler(memStore)
	ticketHandler := handlers.NewTicketHandler(memStore)

	mux := http.NewServeMux()

	// System health check
	mux.HandleFunc("GET /health", handleHealth)

	// Authentication endpoints
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Ticket management endpoints (using Go 1.22 path routing syntax & wrapped with AuthMiddleware)
	auth := handlers.AuthMiddleware(memStore)
	mux.Handle("POST /tickets", auth(http.HandlerFunc(ticketHandler.Create)))
	mux.Handle("GET /tickets", auth(http.HandlerFunc(ticketHandler.List)))
	mux.Handle("GET /tickets/{id}", auth(http.HandlerFunc(ticketHandler.GetByID)))
	mux.Handle("PATCH /tickets/{id}/status", auth(http.HandlerFunc(ticketHandler.UpdateStatus)))

	return mux
}

func main() {
	// 1. Initialize the storage layer
	memStore := store.NewMemoryStore()

	// 2. Setup standard Go 1.22+ ServeMux
	mux := SetupRouter(memStore)

	// 3. Build and configure the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting Ticket System server on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
