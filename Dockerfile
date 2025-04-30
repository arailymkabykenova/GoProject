# --- Stage 1: Build ---
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk update && apk add --no-cache ca-certificates sqlite-libs gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the application from the main package
# CGO_ENABLED=1 is needed for mattn/go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -v -o /server ./cmd/server

# --- Stage 2: Run ---
FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /server /app/server

# Copy migration files (needed by InitSchema at runtime)
COPY migration /app/migration

# Create data directory where the DB file will live INSIDE the container
RUN mkdir -p /app/data

EXPOSE 8080

# Set default environment variables inside the container
ENV PORT=8080
ENV BASE_URL="http://localhost:8080"
ENV DB_PATH="/app/data/shortener.db"

# Command to run
CMD ["/app/server"]
