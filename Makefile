.PHONY: build test lint docker-build docker-run clean help

# Variables
APP_NAME=url-shortener
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_COMPOSE_FILE=docker-compose.yml

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o $(APP_NAME) .

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

deps: ## Download dependencies
	go mod download
	go mod tidy

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 \
		-e DATABASE_URL=postgres://user:password@host.docker.internal:5432/urlshortener?sslmode=disable \
		-e REDIS_URL=redis://host.docker.internal:6379 \
		$(DOCKER_IMAGE)

docker-compose-up: ## Start services with Docker Compose
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-compose-down: ## Stop services with Docker Compose
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-compose-logs: ## View Docker Compose logs
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

dev: ## Start development environment
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d postgres redis
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Starting application..."
	@DATABASE_URL=postgres://user:password@localhost:5432/urlshortener?sslmode=disable \
	 REDIS_URL=redis://localhost:6379 \
	 go run .

k8s-apply: ## Apply Kubernetes configurations
	kubectl apply -f k8s/

k8s-delete: ## Delete Kubernetes resources
	kubectl delete -f k8s/

k8s-status: ## Check Kubernetes deployment status
	kubectl get pods -n url-shortener
	kubectl get services -n url-shortener
	kubectl get ingress -n url-shortener

clean: ## Clean build artifacts
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

security-scan: ## Run security scan
	gosec ./...

benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

load-test: ## Run load tests (requires hey tool)
	@which hey > /dev/null || (echo "Please install hey: go install github.com/rakyll/hey@latest" && exit 1)
	hey -n 1000 -c 10 http://localhost:8080/api/v1/health

all: clean deps fmt vet lint test build ## Run all checks and build
