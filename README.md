# Ticket System Backend

A Golang-based ticket management backend with JWT authentication, ownership-based authorization, and Docker support.

---

## 🚀 Deployment

**Base URL**

https://ranjit-eva-bharat-assingment.onrender.com

**Health Check**

https://ranjit-eva-bharat-assingment.onrender.com/health

---

## 💻 Run Locally

### Prerequisites

- Go 1.22+
- Docker

### Install Dependencies

```bash
go mod tidy
```

### Run

```bash
go run main.go
```

The server will start on:

```
http://localhost:8080
```

---

## 🐳 Docker

### Build Image

```bash
docker build -t ticket-system .
```

### Run Container

```bash
docker run -p 8080:8080 ticket-system
```

---

## 📝 Assumptions

- JWT is used for authentication.
- Passwords are securely hashed before storage.
- Data is stored in memory and will be lost when the server restarts.
- Users can only access and modify their own tickets.
- Ticket status follows the workflow:
  - `open → in_progress → closed`
- Once a ticket is marked as `closed`, it cannot be reopened.
- Protected endpoints require:
  ```
  Authorization: Bearer <JWT_TOKEN>
  ```