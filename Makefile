.PHONY: help
help:
	@echo "Available targets:"
	@echo "  lint      - run linter"
	@echo "  lint-fix  - run linter and fix"
	@echo "  test      - run tests"
	@echo "  build     - build project"

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build:
	go build -o bin/gosql ./cmd/app