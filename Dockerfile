# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor/
RUN go mod verify

# Install migrate CLI
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o ./out/app ./cmd/app/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy migrate CLI from builder
COPY --from=builder /go/bin/migrate /usr/local/bin/

COPY --from=builder /app/out/app .
COPY migrations ./migrations

EXPOSE 8080

CMD ["./app"]
