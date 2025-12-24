.PHONY: build run dev test cover lint clean migrate docker-up docker-down

# Build
build:
	go build -o bin/api ./cmd/api

# Run
run: build
	./bin/api

# Development with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test -v ./...

# Test coverage
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint (requires golangci-lint)
lint:
	golangci-lint run

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html

# Database migrate
migrate:
	go run ./cmd/api -migrate

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-logs:
	docker-compose logs -f backend

# Generate mocks (requires mockgen)
mocks:
	go generate ./...

# Tidy dependencies
tidy:
	go mod tidy

# All checks
check: lint test

# Help
help:
	@echo "Available commands:"
	@echo "  make build       - Build the binary"
	@echo "  make run         - Build and run"
	@echo "  make dev         - Run with hot reload"
	@echo "  make test        - Run tests"
	@echo "  make cover       - Test coverage"
	@echo "  make lint        - Run linter"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-up   - Start Docker services"
	@echo "  make docker-down - Stop Docker services"
	@echo "  make tidy        - Tidy dependencies"
