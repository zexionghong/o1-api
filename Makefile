# AI API Gateway Makefile

.PHONY: help build run test clean migrate-up migrate-down docker-build docker-run

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"

# Build the application
build:
	@echo "Building AI API Gateway..."
	@go build -o bin/server cmd/server/main.go
	@go build -o bin/migrate cmd/migrate/main.go

# Run the application
run:
	@echo "Starting AI API Gateway..."
	@go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf data/

# Run database migrations up
migrate-up:
	@echo "Running migrations up..."
	@go run cmd/migrate/main.go up

# Run database migrations down
migrate-down:
	@echo "Running migrations down..."
	@go run cmd/migrate/main.go down

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t ai-api-gateway .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 ai-api-gateway

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Generate mocks
mocks:
	@echo "Generating mocks..."
	@go generate ./...
