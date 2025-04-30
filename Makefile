BINARY_NAME=go-url-shortener
CMD_PATH=./cmd/server

all: help

build:
	@echo "Building Go application..."
	@go build -o bin/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)"

run: build
	@echo "Running Go application locally..."
	@./bin/$(BINARY_NAME)

clean:
	@echo "Cleaning build artifacts..."
	@rm -f bin/$(BINARY_NAME)
	@echo "Clean complete."

docker-build:
	@echo "Building Docker image..."
	@docker compose build
	@echo "Docker image built."

docker-up:
	@echo "Starting Docker container..."
	@docker compose up -d
	@echo "Container started."

docker-down:
	@echo "Stopping Docker container..."
	@docker compose down
	@echo "Container stopped."

docker-logs:
	@echo "Showing logs for Docker container..."
	@docker compose logs -f

help:
	@echo "Available commands:"
	@echo "  make build         - Build the Go application locally"
	@echo "  make run           - Build and run the Go application locally"
	@echo "  make clean         - Remove local build artifacts"
	@echo "  make docker-build  - Build the Docker image"
	@echo "  make docker-up     - Start the container using Docker Compose"
	@echo "  make docker-down   - Stop the container using Docker Compose"
	@echo "  make docker-logs   - View logs from the running container"

.PHONY: all build run clean docker-build docker-up docker-down docker-logs help