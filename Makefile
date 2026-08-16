.PHONY: build run dev test cover lint clean migrate docker-up docker-down docker-prod

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

# Test with race detection
test-race:
	go test -v -race ./...

# Test coverage
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint (requires golangci-lint)
lint:
	golangci-lint run

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html api swiftlead-backend backend

# Database migrate
migrate:
	go run ./cmd/api -migrate

# Docker (Development)
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-logs:
	docker-compose logs -f backend

docker-reset:
	docker-compose down -v && docker-compose up -d

# Docker (Production)
docker-prod-up:
	docker-compose -f docker-compose.prod.yml up -d

docker-prod-down:
	docker-compose -f docker-compose.prod.yml down

docker-prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

docker-prod-pull:
	docker-compose -f docker-compose.prod.yml pull

# Build for production
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/api ./cmd/api

# Build Docker image
docker-image:
	docker build -t swiftlet-backend:latest .

# Generate mocks (requires mockgen)
mocks:
	go generate ./...

# Tidy dependencies
tidy:
	go mod tidy

# Verify dependencies
verify:
	go mod verify

# All checks before commit
check: lint test

# Full CI check
ci: tidy verify lint test-race build

# Help
help:
	@echo "Swiftlet Backend - Make Commands"
	@echo ""
	@echo "Development:"
	@echo "  make build         - Build the binary"
	@echo "  make run           - Build and run"
	@echo "  make dev           - Run with hot reload"
	@echo "  make test          - Run tests"
	@echo "  make cover         - Test coverage"
	@echo "  make lint          - Run linter"
	@echo ""
	@echo "Docker (Dev):"
	@echo "  make docker-up     - Start Docker services"
	@echo "  make docker-down   - Stop Docker services"
	@echo "  make docker-reset  - Reset (remove volumes)"
	@echo ""
	@echo "Docker (Prod):"
	@echo "  make docker-prod-up    - Start production"
	@echo "  make docker-prod-down  - Stop production"
	@echo "  make docker-prod-logs  - View logs"
	@echo ""
	@echo "CI/CD:"
	@echo "  make ci            - Full CI check"
	@echo "  make build-prod    - Production build"
