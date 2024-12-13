# Stage 1: Build the Go application
FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/main cmd/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app .

RUN chmod +x /app/bin/main

CMD ["/app/bin/main"]
