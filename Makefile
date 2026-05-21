.PHONY: build test lint fmt run clean deps help vet verify docker-build docker-run migrate migrate-status migrate-prod migrate-prod-status migrate-down predeploy-prod

# Variables
BINARY_NAME=hotlines3-api
BUILD_DIR=bin
GO=go
GOFLAGS=-v

# Default target
.DEFAULT_GOAL := help

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

## run: Run the application in development mode
run:
	@echo "Running application..."
	$(GO) run main.go

## fmt: Format code with gofmt and goimports
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Run: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. See: https://golangci-lint.run/usage/install/"; \
	fi

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## test: Run tests (when tests are added)
test:
	@echo "Running tests..."
	$(GO) test -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## verify: Verify dependencies and code compiles
verify:
	@echo "Verifying..."
	$(GO) mod verify
	$(GO) build ./...

## migrate: Run Goose migrations using config.yaml + .env
migrate:
	@echo "Running database migrations from config.yaml/.env..."
	./scripts/migrate.sh config up

## migrate-status: Show Goose migration status using config.yaml + .env
migrate-status:
	@echo "Checking database migration status..."
	./scripts/migrate.sh config status

## migrate-down: Roll back one Goose migration using config.yaml + .env
migrate-down:
	@echo "Rolling back one database migration..."
	./scripts/migrate.sh config down

## migrate-prod: Run Goose migrations for production; requires GOOSE_DBSTRING or DATABASE_URL
migrate-prod:
	@echo "Running production database migrations..."
	./scripts/migrate.sh prod up

## migrate-prod-status: Show production Goose migration status; requires GOOSE_DBSTRING or DATABASE_URL
migrate-prod-status:
	@echo "Checking production database migration status..."
	./scripts/migrate.sh prod status

## predeploy-prod: Run tests and production Goose migrations before Cloud Run deploy
predeploy-prod:
	@echo "Running predeploy production gate..."
	$(GO) test ./...
	$(MAKE) migrate-prod
	$(MAKE) migrate-prod-status

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker-compose up

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
