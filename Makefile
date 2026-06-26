.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - build the binary"
	@echo "  test          - run all tests"
	@echo "  test-verbose  - run all tests with verbose output"
	@echo "  test-cover    - run tests with coverage report"
	@echo "  test-models   - run model tests only"
	@echo "  test-repo     - run repository tests only"
	@echo "  test-service  - run service tests only"
	@echo "  fmt           - format Go code"
	@echo "  clean         - remove build artifacts"
	@echo "  sqlc-gen      - regenerate sqlc code"

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