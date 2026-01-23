.PHONY: build test install clean release

# Version information
VERSION ?= dev
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -X github.com/adriancodes/gumloop/internal/cli.Version=$(VERSION) \
           -X github.com/adriancodes/gumloop/internal/cli.GitCommit=$(GIT_COMMIT) \
           -X github.com/adriancodes/gumloop/internal/cli.BuildDate=$(BUILD_DATE)

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o bin/gumloop ./cmd/gumloop

# Run tests
test:
	go test ./...

# Install to user's local bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/gumloop

# Build a release binary: make release VERSION=v1.1.0
release:
	@test -n "$(VERSION)" || (echo "Usage: make release VERSION=v1.1.0" && exit 1)
	@test "$(VERSION)" != "dev" || (echo "Usage: make release VERSION=v1.1.0" && exit 1)
	go test ./...
	go build -ldflags "$(LDFLAGS)" -o bin/gumloop ./cmd/gumloop
	@echo ""
	@./bin/gumloop version
	@echo ""
	@echo "Release binary ready at bin/gumloop"
	@echo "Next: git add bin/gumloop && git commit && git push"

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Development build with race detector
dev:
	go build -race -ldflags "$(LDFLAGS)" -o bin/gumloop ./cmd/gumloop

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy
