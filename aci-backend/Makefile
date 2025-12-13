.PHONY: build test lint migrate-up migrate-down docker-build docker-up docker-down clean

# Build configuration
BINARY_NAME=aci-backend
BUILD_DIR=./bin
MAIN_PATH=./cmd/server

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report: coverage.html"

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@go test -v -race ./tests/unit/...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test -v -race ./tests/integration/...

# Run linter
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Run database migrations up
migrate-up:
	@echo "Running migrations up..."
	@migrate -path ./migrations -database "$(DATABASE_URL)" up

# Run database migrations down
migrate-down:
	@echo "Running migrations down..."
	@migrate -path ./migrations -database "$(DATABASE_URL)" down

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest -f deployments/docker/Dockerfile .

# Start Docker Compose
docker-up:
	@echo "Starting Docker Compose..."
	@docker-compose -f deployments/docker/docker-compose.yml up -d

# Stop Docker Compose
docker-down:
	@echo "Stopping Docker Compose..."
	@docker-compose -f deployments/docker/docker-compose.yml down

# View Docker logs
docker-logs:
	@docker-compose -f deployments/docker/docker-compose.yml logs -f

# Run the application locally
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Development dependencies installed"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application binary"
	@echo "  test           - Run all tests with coverage"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Tidy Go modules"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Run database migrations down"
	@echo "  migrate-create - Create a new migration"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-up      - Start Docker Compose services"
	@echo "  docker-down    - Stop Docker Compose services"
	@echo "  docker-logs    - View Docker Compose logs"
	@echo "  run            - Run the application locally"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev-deps       - Install development dependencies"
