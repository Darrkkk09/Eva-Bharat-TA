
FROM golang:1.22-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY go.mod ./

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ticket-system main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder /app/ticket-system .
USER appuser

EXPOSE 8080

ENTRYPOINT ["./ticket-system"]
