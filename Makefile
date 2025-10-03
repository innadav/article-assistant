.PHONY: test test-all test-unit test-integration test-e2e test-ingestion test-coverage build run clean

# Test commands
test: test-unit test-integration

test-all: test-unit test-integration test-e2e test-ingestion

test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/... ./internal/ingest/...

test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/... -tags=integration

test-e2e:
	@echo "Running e2e tests..."
	go test -v ./tests/e2e/... -tags=integration

test-ingestion:
	@echo "Running ingestion e2e tests..."
	go test -v ./tests/e2e/ingestion_test.go -tags=integration

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build commands
build:
	@echo "Building application..."
	go build -o bin/server ./cmd/server

# Run commands
run-server:
	@echo "Starting server..."
	go run ./cmd/server

# Docker commands
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# Database commands
db-reset:
	@echo "Resetting database..."
	docker-compose exec postgres psql -U postgres -d article_assistant -c "DROP TABLE IF EXISTS articles;"
	docker-compose exec postgres psql -U postgres -d article_assistant -c "CREATE TABLE articles (id SERIAL PRIMARY KEY, url TEXT UNIQUE, title TEXT, summary TEXT, embedding vector(1536), entities TEXT[], keywords TEXT[], sentiment TEXT, tone TEXT);"
	docker-compose exec postgres psql -U postgres -d article_assistant -c "CREATE INDEX articles_embedding_idx ON articles USING ivfflat (embedding vector_cosine_ops);"

# Clean commands
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Development commands
dev-setup:
	@echo "Setting up development environment..."
	cp env.example .env
	@echo "Please update .env with your OpenAI API key"
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	sleep 10
	make db-reset

dev-test:
	@echo "Running development tests..."
	make test-unit
	@echo "Running integration tests (requires API key)..."
	make test-integration

# Help
help:
	@echo "Available commands:"
	@echo "  test              - Run unit and integration tests"
	@echo "  test-all          - Run unit, integration, and e2e tests"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration   - Run integration tests only"
	@echo "  test-e2e          - Run e2e tests only"
	@echo "  test-ingestion    - Run ingestion e2e tests only"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  build             - Build application binaries"
	@echo "  run-server        - Run the server"
	@echo "  docker-build      - Build Docker images"
	@echo "  docker-up         - Start Docker services"
	@echo "  docker-down       - Stop Docker services"
	@echo "  docker-logs       - Show Docker logs"
	@echo "  db-reset          - Reset database schema"
	@echo "  clean             - Clean build artifacts"
	@echo "  dev-setup         - Setup development environment"
	@echo "  dev-test          - Run development tests"
