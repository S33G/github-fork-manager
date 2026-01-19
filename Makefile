.PHONY: build build-all test test-coverage clean install uninstall dev fmt tidy deps lint

BINARY_NAME=github-fork-manager
VERSION?=dev
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
GO?=$(shell which go || echo go)
GOFLAGS=-trimpath
PREFIX?=$(HOME)/.local
INSTALL_DIR=$(PREFIX)/bin

# Build for current platform
build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/github-fork-manager

# Build for common platforms
build-all:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/github-fork-manager
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/github-fork-manager
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/github-fork-manager
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/github-fork-manager
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/github-fork-manager

# Run tests
test:
	$(GO) test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html

# Install to ~/.local/bin (or custom PREFIX)
install: build
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "Ensure $(INSTALL_DIR) is on your PATH"

# Uninstall from installation directory
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Uninstalled"
	@echo "To remove configuration: rm -rf ~/.github-fork-manager"

# Run in development mode with optional ARGS
dev:
	$(GO) run ./cmd/github-fork-manager $(ARGS)

# Format code
fmt:
	$(GO) fmt ./...
	gofmt -s -w .

# Tidy dependencies
tidy:
	$(GO) mod tidy

# Download dependencies
deps:
	$(GO) mod download

# Lint (requires golangci-lint installed)
lint:
	golangci-lint run
