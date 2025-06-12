.PHONY: build run test clean docker docker-run fmt lint deps deps-update help

# Variables
BINARY_NAME=videocraft
DOCKER_IMAGE=videocraft:latest
GO=go
GOFLAGS=-v

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  run          - Build and run the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker       - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  deps         - Download dependencies"
	@echo "  deps-update  - Update dependencies"

# Build
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) cmd/server/main.go

# Run
run: build
	./$(BINARY_NAME)

# Test
test:
	$(GO) test $(GOFLAGS) ./...

# Test with coverage
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Docker build
docker:
	docker build -t $(DOCKER_IMAGE) .

# Docker run
docker-run:
	docker run -p 3002:3002 \
		-v ./generated_videos:/app/generated_videos \
		-v ./temp:/app/temp \
		-v ./config:/app/config \
		$(DOCKER_IMAGE)

# Format code
fmt:
	$(GO) fmt ./...

# Lint (requires golangci-lint to be installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  brew install golangci-lint"; \
		echo "  or"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2"; \
	fi

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Update dependencies
deps-update:
	$(GO) get -u ./...
	$(GO) mod tidy

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	mkdir -p generated_videos temp whisper_cache
	@echo "Development directories created"
	@echo "Install golangci-lint for linting:"
	@echo "  brew install golangci-lint"

# Quick development run with debug logging
dev-run:
	VIDEOCRAFT_LOG_LEVEL=debug ./$(BINARY_NAME)

# Production build (optimized)
build-prod:
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -ldflags '-w -s' -o $(BINARY_NAME) cmd/server/main.go

# Security scan (requires gosec)
security:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with:"; \
		echo "  go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi