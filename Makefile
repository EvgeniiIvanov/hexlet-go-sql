.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build              - build the binary"
	@echo "  run                - run the application (requires .env)"
	@echo ""
	@echo "Testing:"
	@echo "  test               - run unit tests"
	@echo "  test-verbose       - run unit tests with verbose output"
	@echo "  test-cover         - run unit tests with coverage report"
	@echo "  test-models        - run model tests only"
	@echo "  test-repo          - run repository tests only"
	@echo "  test-service       - run service tests only"
	@echo "  test-integration   - run integration tests"
	@echo "  test-integration-v - run integration tests with verbose output"
	@echo "  test-all           - run both unit and integration tests"
	@echo ""
	@echo "Docker & Database:"
	@echo "  docker-up          - start PostgreSQL with Docker Compose"
	@echo "  docker-down        - stop and remove PostgreSQL containers"
	@echo "  docker-logs        - view PostgreSQL logs"
	@echo "  docker-ps          - show running containers"
	@echo "  docker-clean       - remove containers and volumes (DESTROYS DATA)"
	@echo "  db-shell           - open PostgreSQL shell (psql)"
	@echo "  db-migrate         - run database migrations"
	@echo "  db-reset           - drop and recreate database (DESTROYS DATA)"
	@echo ""
	@echo "Development:"
	@echo "  fmt                - format Go code"
	@echo "  sqlc-gen           - regenerate sqlc code"
	@echo "  clean              - remove build artifacts"
	@echo "  setup              - initial project setup (copy .env.example)"
	@echo ""
	@echo "Documentation:"
	@echo "  README.md          - User guide and usage examples"
	@echo "  ARCHITECTURE.md    - System design and architecture"
	@echo "  DEVELOPMENT.md     - Developer guide and workflows"

.PHONY: build
build:
	go build -o bin/gosql ./cmd/app

.PHONY: test
test:
	go test ./...

.PHONY: test-verbose
test-verbose:
	go test ./... -v

.PHONY: test-cover
test-cover:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-models
test-models:
	go test ./internal/models/... -v

.PHONY: test-repo
test-repo:
	go test ./internal/repository/... -v

.PHONY: test-service
test-service:
	go test ./internal/service/... -v

.PHONY: test-integration
test-integration:
	go test -tags=integration ./internal/tests/... -v

.PHONY: test-integration-v
test-integration-v:
	go test -tags=integration ./internal/tests/... -v -count=1

.PHONY: test-all
test-all:
	@echo "Running unit tests..."
	@go test ./...
	@echo ""
	@echo "Running integration tests..."
	@go test -tags=integration ./internal/tests/... -v

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: sqlc-gen
sqlc-gen:
	sqlc generate

.PHONY: clean
clean:
	rm -f bin/gosql
	rm -f data.db
	rm -f coverage.out coverage.html

# Docker Compose command (try both docker-compose and docker compose)
DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null)
ifndef DOCKER_COMPOSE
	DOCKER_COMPOSE := docker compose
endif

# Docker commands
.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@$(DOCKER_COMPOSE) ps

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

.PHONY: docker-logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f postgres

.PHONY: docker-ps
docker-ps:
	$(DOCKER_COMPOSE) ps

.PHONY: docker-clean
docker-clean:
	@echo "WARNING: This will remove all containers and volumes (all data will be lost)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER_COMPOSE) down -v; \
		echo "Containers and volumes removed."; \
	else \
		echo "Cancelled."; \
	fi

# Database commands
.PHONY: db-shell
db-shell:
	$(DOCKER_COMPOSE) exec postgres psql -U $$(grep POSTGRES_USER .env | cut -d '=' -f2) -d $$(grep POSTGRES_DB .env | cut -d '=' -f2)

.PHONY: db-migrate
db-migrate:
	@echo "Running migrations..."
	@# This will be implemented with a migration tool later
	@echo "Migration support coming soon. For now, run migrations manually."

.PHONY: db-reset
db-reset:
	@echo "WARNING: This will drop and recreate the database (all data will be lost)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER_COMPOSE) exec postgres psql -U $$(grep POSTGRES_USER .env | cut -d '=' -f2) -c "DROP DATABASE IF EXISTS $$(grep POSTGRES_DB .env | cut -d '=' -f2);"; \
		$(DOCKER_COMPOSE) exec postgres psql -U $$(grep POSTGRES_USER .env | cut -d '=' -f2) -c "CREATE DATABASE $$(grep POSTGRES_DB .env | cut -d '=' -f2);"; \
		echo "Database reset complete."; \
	else \
		echo "Cancelled."; \
	fi

# Setup commands
.PHONY: setup
setup:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created. Please update with your settings."; \
	else \
		echo ".env file already exists."; \
	fi

.PHONY: run
run: build
	./bin/gosql