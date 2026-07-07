FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN go build -o ticket-system main.go

EXPOSE 8080

CMD ["./ticket-system"]