# Ticket System Backend in Go (Modularized)

This is a clean, production-grade starter template for a Ticket System backend built in Go, designed for a backend internship. It features a modular directory layout, thread-safe in-memory database store (using dependency injection), and HTTP routing using standard Go 1.22+ features.

## Project Structure

```text
go-evabharat/
├── go.mod            # Go module file
├── main.go           # Server startup, dependency injection & routing setup
├── models/           # Declarations of data models
│   ├── ticket.go
│   └── user.go
├── store/            # In-memory storage layer handling data mutation safely
│   └── memory.go
└── handlers/         # Controller/routing layer parsing JSON requests
    ├── auth.go
    ├── ticket.go
    ├── middleware.go # Authentication middleware
    └── helpers.go    # HTTP response helpers (e.g. JSON encoders)
```

## Requirements

- **Go 1.22 or higher** (uses the enhanced HTTP routing patterns).

## How to Run

1. Clone or navigate to the repository directory.
2. Run the application:
   ```bash
   go run main.go
   ```
3. The server will start listening on port `8080`.

## API Endpoints

### 1. Health Check
- **URL:** `/health`
- **Method:** `GET`
- **Response (200 OK):**
  ```json
  {
    "status": "healthy",
    "time": "2026-07-03T19:00:00+05:30"
  }
  ```

---

### Auth Group (Public)

#### 2. User Registration
- **URL:** `/auth/register`
- **Method:** `POST`
- **Payload:**
  ```json
  {
    "username": "johndoe",
    "email": "johndoe@example.com",
    "password": "securepassword123"
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "id": 1,
    "username": "johndoe",
    "email": "johndoe@example.com"
  }
  ```
  *(Note: Password is kept hidden from JSON payloads automatically using struct tags).*

#### 3. User Login
- **URL:** `/auth/login`
- **Method:** `POST`
- **Payload:**
  ```json
  {
    "email": "johndoe@example.com",
    "password": "securepassword123"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "token": "dummy-jwt-token-for-user-1",
    "user": {
      "id": 1,
      "username": "johndoe",
      "email": "johndoe@example.com"
    }
  }
  ```

---

### Ticket Group (Protected)
*All routes in this group require the HTTP Header: `Authorization: Bearer dummy-jwt-token-for-user-<id>`*

#### 4. Create Ticket
- **URL:** `/tickets`
- **Method:** `POST`
- **Payload:**
  ```json
  {
    "title": "Bug in payments page",
    "description": "Getting a 500 error on checkout page"
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "id": 1,
    "title": "Bug in payments page",
    "description": "Getting a 500 error on checkout page",
    "status": "open",
    "created_by": 1,
    "created_at": "2026-07-03T19:05:00+05:30",
    "updated_at": "2026-07-03T19:05:00+05:30"
  }
  ```

#### 5. List Tickets
- **URL:** `/tickets`
- **Method:** `GET`
- **Description:** Returns only the tickets belonging to the authenticated user.
- **Response (200 OK):**
  ```json
  [
    {
      "id": 1,
      "title": "Bug in payments page",
      "description": "Getting a 500 error on checkout page",
      "status": "open",
      "created_by": 1,
      "created_at": "2026-07-03T19:05:00+05:30",
      "updated_at": "2026-07-03T19:05:00+05:30"
    }
  ]
  ```

#### 6. Get Ticket by ID
- **URL:** `/tickets/{id}` (e.g. `/tickets/1`)
- **Method:** `GET`
- **Description:** Returns the details of a specific ticket. Users can only fetch tickets they created.
- **Response (200 OK):**
  ```json
  {
    "id": 1,
    "title": "Bug in payments page",
    "description": "Getting a 500 error on checkout page",
    "status": "open",
    "created_by": 1,
    "created_at": "2026-07-03T19:05:00+05:30",
    "updated_at": "2026-07-03T19:05:00+05:30"
  }
  ```

#### 7. Update Ticket Status
- **URL:** `/tickets/{id}/status` (e.g. `/tickets/1/status`)
- **Method:** `PATCH`
- **Description:** Updates the ticket status. State transitions must strictly follow: `open` -> `in_progress` -> `closed`. Once a ticket is `closed`, it cannot be reopened or edited.
- **Payload:**
  ```json
  {
    "status": "in_progress"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "id": 1,
    "title": "Bug in payments page",
    "description": "Getting a 500 error on checkout page",
    "status": "in_progress",
    "created_by": 1,
    "created_at": "2026-07-03T19:05:00+05:30",
    "updated_at": "2026-07-03T19:06:00+05:30"
  }
  ```

---

## Testing with `curl`

**1. Create a user:**
```bash
curl -i -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "intern", "email": "intern@example.com", "password": "password123"}'
```

**2. Login to get the dummy token:**
```bash
curl -i -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "intern@example.com", "password": "password123"}'
```

**3. Create a Ticket (linked to authenticated user):**
```bash
curl -i -X POST http://localhost:8080/tickets \
  -H "Authorization: Bearer dummy-jwt-token-for-user-1" \
  -H "Content-Type: application/json" \
  -d '{"title": "Database connection drop", "description": "Losing connection every 5 minutes"}'
```

**4. Retrieve all Tickets (filtered for user 1):**
```bash
curl -i -H "Authorization: Bearer dummy-jwt-token-for-user-1" http://localhost:8080/tickets
```

**5. Get Ticket with ID 1:**
```bash
curl -i -H "Authorization: Bearer dummy-jwt-token-for-user-1" http://localhost:8080/tickets/1
```

**6. Transition ticket status to in_progress:**
```bash
curl -i -X PATCH http://localhost:8080/tickets/1/status \
  -H "Authorization: Bearer dummy-jwt-token-for-user-1" \
  -H "Content-Type: application/json" \
  -d '{"status": "in_progress"}'
```

**7. Transition ticket status to closed:**
```bash
curl -i -X PATCH http://localhost:8080/tickets/1/status \
  -H "Authorization: Bearer dummy-jwt-token-for-user-1" \
  -H "Content-Type: application/json" \
  -d '{"status": "closed"}'
```

**8. Attempt to reopen closed ticket (Expected: 400 Bad Request):**
```bash
curl -i -X PATCH http://localhost:8080/tickets/1/status \
  -H "Authorization: Bearer dummy-jwt-token-for-user-1" \
  -H "Content-Type: application/json" \
  -d '{"status": "in_progress"}'
```

---

## Docker Support

This setup includes a multi-stage `Dockerfile` and a `docker-compose.yml` configuration.

### Running with Docker Compose
If you have Docker installed locally, you can build and run the entire application with one command:
```bash
docker compose up --build -d
```
Verify the server is running by hitting the health check endpoint:
```bash
curl -i http://localhost:8080/health
```

### Running and Testing Without Local Docker (Using Google Cloud Shell)
If you don't have Docker installed on your PC, you can test the Docker setup for free in your browser using **Google Cloud Shell** (which has Docker pre-installed):

1. Open [Google Cloud Shell](https://shell.cloud.google.com).
2. Clone this repository (or drag-and-drop the files into the Cloud Shell editor).
3. Start the container in the background:
   ```bash
   docker compose up --build -d
   ```
4. Test the health check endpoint:
   ```bash
   curl -i http://localhost:8080/health
   ```
5. Test the entire route test suite inside a temporary Go docker container:
   ```bash
   docker run --rm -v $(pwd):/app -w /app golang:1.22-alpine go test -v ./...
   ```
6. Stop the container when finished:
   ```bash
   docker compose down
   ```

