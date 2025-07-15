# Makefile for dbutil-gen project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Project parameters
BINARY_NAME=dbutil-gen
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/dbutil-gen
DOCKER_COMPOSE=docker-compose

# Test parameters
TEST_DB_URL=postgres://dbutil:dbutil_test_password@localhost:5432/dbutil_test?sslmode=disable
TEST_TIMEOUT=30s

# Default target - show help
.PHONY: default
default: help

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "✅ Binary built: $(BINARY_PATH)"

# Run tests (includes integration tests if database is available)
.PHONY: test
test:
	@echo "Running tests..."
	$(GOMOD) tidy
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) ./...
	@echo "✅ Tests completed"

# Run linter and formatter
.PHONY: lint
lint:
	@echo "Running linter and formatter..."
	go fmt ./...
	$(GOLINT) run ./...
	@echo "✅ Linting completed"

# Setup development environment
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Starting PostgreSQL database..."
	$(DOCKER_COMPOSE) up -d postgres
	@echo "Waiting for database to be ready..."
	@bash -c 'for i in {1..30}; do if pg_isready -h localhost -p 5432 -U dbutil -d dbutil_test >/dev/null 2>&1; then break; fi; sleep 1; done'
	@echo "Running test data migrations..."
	@./test/run_migrations.sh
	@echo "✅ Development environment ready!"
	@echo "Database URL: $(TEST_DB_URL)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf test_output/
	@rm -rf test-output/
	@$(DOCKER_COMPOSE) down -v >/dev/null 2>&1 || true
	@echo "✅ Cleanup completed"

# Show help
.PHONY: help
help:
	@echo ""
	@echo "🔧 dbutil-gen - Database-first code generator for PostgreSQL"
	@echo ""
	@echo "📋 USAGE:"
	@echo "  make <target>    Run a specific target"
	@echo "  make             Show this help message"
	@echo ""
	@echo "🚀 ESSENTIAL TARGETS:"
	@echo "  build            Build the dbutil-gen binary"
	@echo "  test             Run all tests (unit + integration)"
	@echo "  lint             Run linter and code formatter"
	@echo "  dev-setup        Setup development environment with database"
	@echo "  clean            Remove build artifacts and stop services"
	@echo ""
	@echo "💡 QUICK START:"
	@echo "  make dev-setup   # Setup database and run migrations"
	@echo "  make build       # Build the tool"
	@echo "  make test        # Run tests"
	@echo ""
	@echo "📚 MORE INFO:"
	@echo "  ./bin/dbutil-gen --help    # CLI tool usage and options"
	@echo "  https://github.com/nhalm/dbutil    # Documentation"
	@echo "" 