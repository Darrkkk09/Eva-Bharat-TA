# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

# Install CA certificates and build dependencies
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy dependency definition
COPY go.mod ./

# Copy the rest of the project source files
COPY . .

# Compile the application as a statically linked binary
# CGO_ENABLED=0 disables C dependencies, allowing the binary to run on any Linux environment
# -ldflags="-w -s" strips debug symbols and debugging information to minimize size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ticket-system main.go

# Stage 2: Create the minimal production image
FROM alpine:latest

# Install certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Create a non-root user and group for running the container securely
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set runtime working directory
WORKDIR /home/appuser

# Copy the compiled binary from the builder stage
COPY --from=builder /app/ticket-system .

# Use the secure non-root user
USER appuser

# Expose default HTTP Port
EXPOSE 8080

# Start the application
ENTRYPOINT ["./ticket-system"]
