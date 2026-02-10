.PHONY: build test lint fmt run clean deps help vet verify docker-build docker-run

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
