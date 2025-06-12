.PHONY: build run test clean docker-build docker-run lint fmt vet install-tools coverage benchmark security help

# Variables
BINARY_NAME=videocraft
DOCKER_IMAGE=videocraft
DOCKER_TAG=latest
VERSION?=$(shell cat VERSION)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) cmd/server/main.go

# Build for multiple platforms
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				ext=".exe"; \
			else \
				ext=""; \
			fi; \
			echo "Building $$os/$$arch..."; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) \
				-o dist/$(BINARY_NAME)-$$os-$$arch$$ext cmd/server/main.go; \
		done; \
	done

# Run the application
run: ## Run the application
	go run $(LDFLAGS) cmd/server/main.go

# Install development tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# Format code
fmt: ## Format Go code
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found, install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Lint code
lint: ## Run linters
	golangci-lint run
	staticcheck ./...

# Vet code
vet: ## Run go vet
	go vet ./...

# Run tests
test: ## Run tests
	@mkdir -p generated_videos temp whisper_cache
	go test -v -race ./...

# Run tests with coverage
coverage: ## Run tests with coverage
	@mkdir -p generated_videos temp whisper_cache
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

# Run integration tests
test-integration: ## Run integration tests
	@mkdir -p generated_videos temp whisper_cache
	go test -v -tags=integration ./...

# Run benchmarks
benchmark: ## Run benchmarks
	@mkdir -p generated_videos temp whisper_cache
	go test -bench=. -benchmem ./...

# Security scan
security: ## Run security scan
	govulncheck ./...
	gosec ./...
	@echo "Security scan completed"

# Clean build artifacts
clean: ## Clean build artifacts
	go clean
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -rf generated_videos/ temp/ whisper_cache/

# Setup development environment
dev-setup: install-tools ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	go mod verify
	pip3 install -r scripts/requirements.txt
	@mkdir -p generated_videos temp whisper_cache
	@echo "Development environment ready!"

# Build Docker image
docker-build: ## Build Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg VCS_REF=$(GIT_COMMIT) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):$(VERSION) .

# Run Docker container
docker-run: ## Run Docker container
	docker run -p 3002:3002 \
		-v $(PWD)/generated_videos:/app/generated_videos \
		-v $(PWD)/temp:/app/temp \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Build and run with Docker Compose
docker-compose-up: ## Run with Docker Compose
	docker-compose up --build

# Stop Docker Compose
docker-compose-down: ## Stop Docker Compose
	docker-compose down

# Check code quality
quality-check: fmt vet lint test security ## Run all quality checks

# Release build (used by CI)
release-build: ## Build release version
	@echo "Building release $(VERSION)..."
	CGO_ENABLED=0 go build \
		-ldflags "-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)" \
		-o $(BINARY_NAME) cmd/server/main.go

# Show version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Generate documentation
docs: ## Generate documentation
	@echo "Documentation is available in CLAUDE.md files"
	@echo "Main documentation: README.md"
	@echo "Technical docs: CLAUDE.md"
	@echo "API docs: internal/api/CLAUDE.md"

# Local development server with auto-reload
dev: ## Run development server with auto-reload
	@echo "Starting development server..."
	@echo "Note: Install 'air' for auto-reload: go install github.com/cosmtrek/air@latest"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found, running without auto-reload..."; \
		$(MAKE) run; \
	fi