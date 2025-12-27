.PHONY: help build run test lint fmt coverage clean mod-tidy mod-download

# Default target
help:
	@echo "Available targets:"
	@echo "  make build        - Build the ynab_import binary"
	@echo "  make run          - Run the application directly"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run go vet linter"
	@echo "  make fmt          - Format code with gofmt"
	@echo "  make coverage     - Generate test coverage report"
	@echo "  make clean        - Remove build artifacts"

# Build the binary
build:
	mkdir -p bin
	go build -o bin/ynab_import

# Run the application
run:
	go run .

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	go vet ./...

# Format code
fmt:
	gofmt -w .

# Generate test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -rf bin
	rm -f coverage.out coverage.html
