# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=url-shortener
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker parameters
DOCKER_IMAGE=url-shortener
DOCKER_TAG=latest

# Test parameters
TEST_TIMEOUT=30s
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

.PHONY: all build clean test coverage lint docker-build docker-run help

all: clean test build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v .

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

# Run tests
test:
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) ./...

# Run tests with race detection
test-race:
	$(GOTEST) -v -race -timeout $(TEST_TIMEOUT) ./...

# Run tests with coverage
coverage:
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

# Run tests with coverage and open in browser (macOS)
coverage-open: coverage
	open $(COVERAGE_HTML)

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install linter
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

# Format code
fmt:
	$(GOCMD) fmt ./...

# Run the application
run:
	$(GOCMD) run .

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up:
	docker-compose up --build -d

docker-compose-down:
	docker-compose down

# Load testing
load-test:
	$(GOCMD) run test-6000-requests-go.go

# Security scan
security:
	gosec ./...

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  build-linux    - Build for Linux"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  coverage       - Run tests with coverage"
	@echo "  coverage-open  - Run tests with coverage and open in browser"
	@echo "  lint           - Run linter"
	@echo "  deps           - Install dependencies"
	@echo "  install-linter - Install golangci-lint"
	@echo "  fmt            - Format code"
	@echo "  run            - Run the application"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-compose-up   - Start with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose"
	@echo "  load-test      - Run load test"
	@echo "  security       - Run security scan"
	@echo "  help           - Show this help"