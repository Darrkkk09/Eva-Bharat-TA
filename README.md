# Ticket System Backend

## Project Overview

This repository contains a **ticket‑system backend** written in Go. It provides RESTful APIs for:
- User registration and login 
- Ticket creation, listing, retrieval, and status updates
- Ownership checks so users can only modify their own tickets

The service is built with the standard `net/http` ServeMux (Go 1.22+), uses an in‑memory store (`store.MemoryStore`), and is fully containerised via Docker.

## Local Development

### Prerequisites
- Go 1.22 or newer
- Docker (for container builds)

### Run the server locally (no Docker)
```bash
# Install dependencies (if any)
 go mod tidy

# Start the server
 go run main.go
```
The server will listen on `localhost:8080` (or the port set in the `PORT` environment variable).

### Health check
```bash
curl http://localhost:8080/health
# Expected output
# {"status":"ok"}
```

## Docker

### Build the image
```bash
docker build -t ticket-system .
```

### Run the container
```bash
docker run -p 8080:8080 ticket-system
```
You can now hit the same endpoints as in local mode, e.g. `curl http://localhost:8080/health`.

## Deployment

The application can be deployed to any free‑tier hosting platform that supports Docker images (Render, Railway, Fly.io, etc.).

- **Deployed URL (placeholder):** `https://<your‑service>.onrender.com`
- The `/health` endpoint must be publicly accessible and return `{ "status": "ok" }`.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT`   | Port the server binds to | `8080` |

Create a `.env.example` file with the above variables to illustrate required configuration.

## Assumptions & Notes
- The JWT is a dummy token; no real signing is performed.
- The in‑memory store means data is lost on server restart – suitable for a prototype or demo.
- No external database or caching layer is used to keep the implementation simple.
- The health endpoint returns a static JSON payload `{ "status": "ok" }` as required by the deployment contract.

## License

MIT – feel free to fork, modify, and deploy.
