# Source Fetcher Makefile

# Variables
BINARY_NAME=source-fetcher
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Platforms
PLATFORMS=windows linux darwin
ARCHITECTURES=amd64 arm64

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all build clean test coverage lint help install uninstall run dev deps upgrade

## all: Default target - build for current platform
all: clean build

## build: Build binary for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

## build-all: Build binaries for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHITECTURES),\
			echo "Building $(BINARY_NAME)-$(GOOS)-$(GOARCH)...";\
			GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build $(LDFLAGS) \
				-o bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,) . || true;\
		)\
	)
	@echo "Build complete for all platforms"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_NAME) $(BINARY_NAME).exe bin/ dist/
	@rm -rf *.out *.test *.prof
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

## test-coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=5m

## fmt: Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## mod-tidy: Tidy go modules
mod-tidy:
	@echo "Tidying go modules..."
	go mod tidy
	go mod verify

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

## upgrade: Upgrade dependencies
upgrade:
	@echo "Upgrading dependencies..."
	go get -u ./...
	go mod tidy

## install: Install binary to system
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin (requires sudo)..."
	@sudo mv $(BINARY_NAME) /usr/local/bin/sfer
	@echo "Installed as 'sfer' command"

## uninstall: Remove installed binary
uninstall:
	@echo "Uninstalling sfer from /usr/local/bin (requires sudo)..."
	@sudo rm -f /usr/local/bin/sfer
	@echo "Uninstalled"

## run: Build and run
run: build
	./$(BINARY_NAME) $(ARGS)

## dev: Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@which air > /dev/null || (echo "air not installed. Install: go install github.com/cosmtrek/air@latest" && exit 1)
	air

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

## profile-cpu: CPU profiling
profile-cpu:
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=.
	go tool pprof cpu.prof

## profile-mem: Memory profiling
profile-mem:
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=.
	go tool pprof mem.prof

## release: Create a release build (requires VERSION=vX.Y.Z)
release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Usage: make release VERSION=v1.0.0"; exit 1; fi
	@echo "Creating release $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Release $(VERSION) created and pushed"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):$(VERSION)

## help: Show this help message
help: Makefile
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'

.DEFAULT_GOAL := help
