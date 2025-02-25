FROM golang:1.24 AS builder

WORKDIR /app

COPY ../go.mod ../go.sum ./

RUN go mod download

COPY .. .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o chat-app ./cmd

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app .

RUN chmod +x /app/chat-app

CMD ["/app/chat-app"]