.PHONY: build test clean server provider install-deps fmt vet lint acceptance-test chaos-test help

# Variables
BINARY_NAME_SERVER=dirtcloud-server

VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: build

## Help
help: ## Show this help message
	@echo 'Management commands for DirtCloud:'
	@echo
	@echo 'Usage:'
	@echo '  make [target]'
	@echo
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) printf "  %-20s%s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)

## Dependencies
install-deps: ## Install Go dependencies
	go mod download
	go mod tidy

## Building
build: server ## Build server binary

server: ## Build the DirtCloud server
	go build $(LDFLAGS) -o bin/$(BINARY_NAME_SERVER) ./cmd/server



## Development
run-server: ## Run the server locally
	go run ./cmd/server

run-server-with-chaos: ## Run the server with chaos enabled
	DIRT_CHAOS_ENABLED=true \
	DIRT_LATENCY_GLOBAL_MS=10-100 \
	DIRT_ERRRATE_PROJECTS=0.1 \
	DIRT_ERRRATE_INSTANCES=0.05 \
	DIRT_ERRRATE_METADATA=0.05 \
	go run ./cmd/server

## Testing
test: ## Run unit tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

chaos-test: ## Run chaos engineering tests
	DIRT_CHAOS_ENABLED=true go test -v -tags=chaos ./...

test-all: test chaos-test ## Run all tests

## Code quality
fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: install-golangci-lint ## Run golangci-lint
	golangci-lint run

install-golangci-lint:
	@which golangci-lint > /dev/null || \
		(echo "Installing golangci-lint..." && \
		 curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin)

## Database
clean-db: ## Remove the SQLite database file
	rm -f dirt.db dirt.db-shm dirt.db-wal

reset-db: clean-db ## Reset the database (clean and restart server to recreate)

## Docker
docker-build: ## Build Docker image
	docker build -t dirtcloud:$(VERSION) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 dirtcloud:$(VERSION)



## Benchmarking
benchmark: ## Run benchmark tests
	go test -bench=. -benchmem ./...

load-test: ## Run basic load test (requires hey)
	@which hey > /dev/null || (echo "Please install hey: go install github.com/rakyll/hey@latest" && exit 1)
	hey -n 1000 -c 10 http://localhost:8080/v1/projects

## Documentation
docs: ## Generate documentation (placeholder)
	@echo "API documentation generation not implemented yet"

## Cleanup
clean: ## Clean build artifacts and temporary files
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f dirt.db dirt.db-shm dirt.db-wal


## Release preparation
release-prep: clean test-all lint docs ## Prepare for release (run all checks)

## Development workflow shortcuts
dev-setup: install-deps ## Set up development environment
	@echo "Development environment ready!"
	@echo "Run 'make run-server' to start the server"
	@echo "Test the API with curl commands"

quick-test: fmt vet test ## Quick development test cycle