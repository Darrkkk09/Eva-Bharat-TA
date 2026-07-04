package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-evabharat/models"
	"go-evabharat/store"
)

// Helper to make a request and return the response recorder
func performRequest(handler http.Handler, method, path string, body []byte, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestHealthRoute(t *testing.T) {
	memStore := store.NewMemoryStore()
	router := SetupRouter(memStore)

	w := performRequest(router, "GET", "/health", nil, "")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", resp["status"])
	}
	if _, ok := resp["time"]; !ok {
		t.Error("Expected time field in response")
	}
}

func TestAuthRegisterRoute(t *testing.T) {
	memStore := store.NewMemoryStore()
	router := SetupRouter(memStore)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := []byte(`{"username": "testuser", "email": "test@example.com", "password": "password123"}`)
		w := performRequest(router, "POST", "/auth/register", reqBody, "")

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		var user models.User
		if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
			t.Fatalf("Failed to decode user: %v", err)
		}

		if user.Username != "testuser" || user.Email != "test@example.com" {
			t.Errorf("Unexpected user fields: %+v", user)
		}
		if user.Password != "" {
			t.Error("Password should not be exposed in JSON response")
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		reqBody := []byte(`{"username": "", "email": "test2@example.com", "password": "password123"}`)
		w := performRequest(router, "POST", "/auth/register", reqBody, "")

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("duplicate email conflict", func(t *testing.T) {
		// Register once
		reqBody := []byte(`{"username": "user1", "email": "dup@example.com", "password": "password123"}`)
		w1 := performRequest(router, "POST", "/auth/register", reqBody, "")
		if w1.Code != http.StatusCreated {
			t.Fatalf("First register failed")
		}

		// Register again
		w2 := performRequest(router, "POST", "/auth/register", reqBody, "")
		if w2.Code != http.StatusConflict {
			t.Errorf("Expected status 409 conflict, got %d", w2.Code)
		}
	})
}

func TestAuthLoginRoute(t *testing.T) {
	memStore := store.NewMemoryStore()
	router := SetupRouter(memStore)

	// Pre-register user
	_, err := memStore.CreateUser("loginuser", "login@example.com", "secret123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("successful login", func(t *testing.T) {
		reqBody := []byte(`{"email": "login@example.com", "password": "secret123"}`)
		w := performRequest(router, "POST", "/auth/login", reqBody, "")

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp struct {
			Token string      `json:"token"`
			User  models.User `json:"user"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		expectedToken := "dummy-jwt-token-for-user-1"
		if resp.Token != expectedToken {
			t.Errorf("Expected token %s, got %s", expectedToken, resp.Token)
		}
		if resp.User.Email != "login@example.com" {
			t.Errorf("Expected email login@example.com, got %s", resp.User.Email)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		reqBody := []byte(`{"email": "login@example.com", "password": "wrongpassword"}`)
		w := performRequest(router, "POST", "/auth/login", reqBody, "")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		reqBody := []byte(`{"email": "", "password": "password"}`)
		w := performRequest(router, "POST", "/auth/login", reqBody, "")

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestTicketsRoutes(t *testing.T) {
	memStore := store.NewMemoryStore()
	router := SetupRouter(memStore)

	// Create two users in the system
	user1, err := memStore.CreateUser("userone", "one@example.com", "pass123")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	user2, err := memStore.CreateUser("usertwo", "two@example.com", "pass123")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	token1 := fmt.Sprintf("dummy-jwt-token-for-user-%d", user1.ID)
	token2 := fmt.Sprintf("dummy-jwt-token-for-user-%d", user2.ID)

	var ticketID int

	t.Run("create ticket - success", func(t *testing.T) {
		reqBody := []byte(`{"title": "Test Ticket", "description": "Need assistance with testing"}`)
		w := performRequest(router, "POST", "/tickets", reqBody, token1)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		var ticket models.Ticket
		if err := json.Unmarshal(w.Body.Bytes(), &ticket); err != nil {
			t.Fatalf("Failed to decode ticket: %v", err)
		}

		if ticket.Title != "Test Ticket" || ticket.Description != "Need assistance with testing" {
			t.Errorf("Unexpected ticket content: %+v", ticket)
		}
		if ticket.CreatedBy != user1.ID {
			t.Errorf("Expected created_by to be %d, got %d", user1.ID, ticket.CreatedBy)
		}
		if ticket.Status != "open" {
			t.Errorf("Expected status to be open, got %s", ticket.Status)
		}
		ticketID = ticket.ID
	})

	t.Run("create ticket - missing fields", func(t *testing.T) {
		reqBody := []byte(`{"title": "", "description": "some description"}`)
		w := performRequest(router, "POST", "/tickets", reqBody, token1)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("create ticket - unauthorized", func(t *testing.T) {
		reqBody := []byte(`{"title": "Valid Title", "description": "Valid Description"}`)
		w := performRequest(router, "POST", "/tickets", reqBody, "invalid-token")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("list tickets - success", func(t *testing.T) {
		// List for user 1 (should return 1 ticket)
		w1 := performRequest(router, "GET", "/tickets", nil, token1)
		if w1.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w1.Code)
		}

		var tickets1 []models.Ticket
		if err := json.Unmarshal(w1.Body.Bytes(), &tickets1); err != nil {
			t.Fatalf("Failed to decode tickets: %v", err)
		}
		if len(tickets1) != 1 {
			t.Errorf("Expected 1 ticket for user1, got %d", len(tickets1))
		}

		// List for user 2 (should return 0 tickets)
		w2 := performRequest(router, "GET", "/tickets", nil, token2)
		if w2.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w2.Code)
		}

		var tickets2 []models.Ticket
		if err := json.Unmarshal(w2.Body.Bytes(), &tickets2); err != nil {
			t.Fatalf("Failed to decode tickets: %v", err)
		}
		if len(tickets2) != 0 {
			t.Errorf("Expected 0 tickets for user2, got %d", len(tickets2))
		}
	})

	t.Run("get ticket by ID - success", func(t *testing.T) {
		path := fmt.Sprintf("/tickets/%d", ticketID)
		w := performRequest(router, "GET", path, nil, token1)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var ticket models.Ticket
		if err := json.Unmarshal(w.Body.Bytes(), &ticket); err != nil {
			t.Fatalf("Failed to decode ticket: %v", err)
		}

		if ticket.ID != ticketID {
			t.Errorf("Expected ID %d, got %d", ticketID, ticket.ID)
		}
	})

	t.Run("get ticket by ID - forbidden (different user)", func(t *testing.T) {
		path := fmt.Sprintf("/tickets/%d", ticketID)
		w := performRequest(router, "GET", path, nil, token2)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("get ticket by ID - not found", func(t *testing.T) {
		w := performRequest(router, "GET", "/tickets/9999", nil, token1)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("get ticket by ID - invalid ID parameter", func(t *testing.T) {
		w := performRequest(router, "GET", "/tickets/abc", nil, token1)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("update status - valid transition open -> in_progress", func(t *testing.T) {
		path := fmt.Sprintf("/tickets/%d/status", ticketID)
		reqBody := []byte(`{"status": "in_progress"}`)
		w := performRequest(router, "PATCH", path, reqBody, token1)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var ticket models.Ticket
		if err := json.Unmarshal(w.Body.Bytes(), &ticket); err != nil {
			t.Fatalf("Failed to decode ticket: %v", err)
		}

		if ticket.Status != "in_progress" {
			t.Errorf("Expected status in_progress, got %s", ticket.Status)
		}
	})

	t.Run("update status - forbidden (different user)", func(t *testing.T) {
		path := fmt.Sprintf("/tickets/%d/status", ticketID)
		reqBody := []byte(`{"status": "closed"}`)
		w := performRequest(router, "PATCH", path, reqBody, token2)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("update status - invalid transition in_progress -> open (should fail)", func(t *testing.T) {
		path := fmt.Sprintf("/tickets/%d/status", ticketID)
		reqBody := []byte(`{"status": "open"}`)
		w := performRequest(router, "PATCH", path, reqBody, token1)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("update status - valid transition in_progress -> closed", func(t *testing.T) {
		path := fmt.Sprintf("/tickets/%d/status", ticketID)
		reqBody := []byte(`{"status": "closed"}`)
		w := performRequest(router, "PATCH", path, reqBody, token1)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var ticket models.Ticket
		if err := json.Unmarshal(w.Body.Bytes(), &ticket); err != nil {
			t.Fatalf("Failed to decode ticket: %v", err)
		}

		if ticket.Status != "closed" {
			t.Errorf("Expected status closed, got %s", ticket.Status)
		}
	})

	t.Run("update status - modification of closed ticket (should fail)", func(t *testing.T) {
		path := fmt.Sprintf("/tickets/%d/status", ticketID)
		reqBody := []byte(`{"status": "in_progress"}`)
		w := performRequest(router, "PATCH", path, reqBody, token1)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
