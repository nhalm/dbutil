# Makefile for dbutil-gen project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Project parameters
BINARY_NAME=dbutil-gen
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/dbutil-gen
DOCKER_COMPOSE=docker-compose

# Test parameters
TEST_DB_URL=postgres://dbutil:dbutil_test_password@localhost:5432/dbutil_test?sslmode=disable
TEST_TIMEOUT=30s

# Default target
.PHONY: all
all: clean build test

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Binary built: $(BINARY_PATH)"

# Run unit tests
.PHONY: test
test:
	@echo "Running unit tests..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) ./...

# Run integration tests (requires database)
.PHONY: integration-test
integration-test: test-setup
	@echo "Running integration tests..."
	@echo "Waiting for database to be ready..."
	@sleep 5
	TEST_DATABASE_URL=$(TEST_DB_URL) $(GOTEST) -v -timeout $(TEST_TIMEOUT) -tags=integration ./...

# Set up test infrastructure
.PHONY: test-setup
test-setup:
	@echo "Setting up test infrastructure..."
	@echo "Starting PostgreSQL container..."
	$(DOCKER_COMPOSE) up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	@echo "PostgreSQL container started successfully"

# Clean up all artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf ./bin
	@echo "Stopping and removing Docker containers..."
	$(DOCKER_COMPOSE) down -v
	@echo "Cleanup completed"

# Development helpers
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -w .

.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run

.PHONY: tidy
tidy:
	@echo "Tidying Go modules..."
	$(GOMOD) tidy

# Database helpers
.PHONY: db-up
db-up:
	@echo "Starting database..."
	$(DOCKER_COMPOSE) up -d postgres

.PHONY: db-down
db-down:
	@echo "Stopping database..."
	$(DOCKER_COMPOSE) down

.PHONY: db-logs
db-logs:
	@echo "Showing database logs..."
	$(DOCKER_COMPOSE) logs -f postgres

.PHONY: db-shell
db-shell:
	@echo "Connecting to database shell..."
	$(DOCKER_COMPOSE) exec postgres psql -U dbutil -d dbutil_test

# Development workflow
.PHONY: dev-setup
dev-setup: test-setup
	@echo "Development environment setup complete"
	@echo "Database URL: $(TEST_DB_URL)"
	@echo "Run 'make build' to build the binary"
	@echo "Run 'make test' to run unit tests"
	@echo "Run 'make integration-test' to run integration tests"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOGET) -v ./...
	$(GOMOD) tidy

# Generate code (for testing the generator itself)
.PHONY: generate
generate: build
	@echo "Running code generation..."
	DATABASE_URL=$(TEST_DB_URL) $(BINARY_PATH) --tables --output=./test_output

# Run the generator with test database
.PHONY: test-generate
test-generate: build test-setup
	@echo "Testing code generation against test database..."
	@sleep 5
	DATABASE_URL=$(TEST_DB_URL) $(BINARY_PATH) --tables --output=./test_output --verbose
	@echo "Generated code available in ./test_output"

# Benchmark tests
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Coverage report
.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           - Build the dbutil-gen binary"
	@echo "  test            - Run unit tests"
	@echo "  integration-test - Run integration tests (requires database)"
	@echo "  test-setup      - Start PostgreSQL test container"
	@echo "  clean           - Clean up all artifacts and containers"
	@echo "  fmt             - Format Go code"
	@echo "  lint            - Run linter"
	@echo "  tidy            - Tidy Go modules"
	@echo "  db-up           - Start database container"
	@echo "  db-down         - Stop database container"
	@echo "  db-logs         - Show database logs"
	@echo "  db-shell        - Connect to database shell"
	@echo "  dev-setup       - Set up development environment"
	@echo "  deps            - Install dependencies"
	@echo "  generate        - Run code generation"
	@echo "  test-generate   - Test code generation against test database"
	@echo "  bench           - Run benchmarks"
	@echo "  coverage        - Generate coverage report"
	@echo "  help            - Show this help message"

# Check if Docker is available
.PHONY: check-docker
check-docker:
	@which docker > /dev/null || (echo "Docker is required but not installed" && exit 1)
	@which docker-compose > /dev/null || (echo "Docker Compose is required but not installed" && exit 1)

# Ensure Docker is available for database operations
test-setup db-up integration-test: check-docker 