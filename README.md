# Ticket System Backend

A robust Golang-based ticket management system featuring authentication, ownership-based authorization, and containerized deployment.

## 🚀 Deployment
*   **Deployed URL:** [https://ranjit-eva-bharat-assingment.onrender.com](https://ranjit-eva-bharat-assingment.onrender.com)
*   **Health Check:** [https://ranjit-eva-bharat-assingment.onrender.com/health](https://ranjit-eva-bharat-assingment.onrender.com/health)

## 📋 API Endpoints
All protected endpoints require an `Authorization: Bearer <token>` header.

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Service health check |
| `POST` | `/auth/register` | Register a new user |
| `POST` | `/auth/login` | Login and receive JWT |
| `POST` | `/tickets` | Create a new ticket |
| `GET` | `/tickets` | List logged-in user's tickets |
| `GET` | `/tickets/{id}` | Get specific ticket details |
| `PATCH` | `/tickets/{id}/status` | Update ticket status |

## 🛠 Prerequisites
- Go 1.22+
- Docker

## 💻 Local Development

### Run natively
```bash
go mod tidy
go run main.go