.PHONY: test test-unit test-integration test-coverage test-all

ALL_PACKAGES := $(shell go list ./...)
TEST_PACKAGES := $(filter-out %/dist %dist/%,$(ALL_PACKAGES))

# Run all tests (excluding integration tests)
test:
	go test ./... -short

# Run unit tests only
test-unit:
	go test ./... -short -tags '!integration'

# Run integration tests only
test-integration:
	go test ./... -tags 'integration'

# Run tests with coverage
test-coverage:
	go test ./... -short -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run all tests including integration
test-all:
	go test ./... -tags 'integration'

# Run tests with verbose output
test-verbose:
	go test ./... -short -v

# Run tests for a specific package
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-pkg PKG=./internal/templates"; \
		exit 1; \
	fi
	go test -v $(PKG)

# Run tests with race detection
test-race:
	go test ./... -short -race

# Run tests for CI (with race detection and coverage)
test-ci:
	mkdir -p dist
	go test -v -race -coverprofile=dist/coverage.out -covermode=atomic $(TEST_PACKAGES)

# Clean coverage files
clean:
	rm -f coverage.out coverage.html
