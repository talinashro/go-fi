.PHONY: test test-race test-coverage build clean examples

# Default target
all: test build

# Run tests (excluding examples)
test:
	go test -v .

# Run tests with race detector
test-race:
	go test -race -v .

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out -covermode=atomic .
	go tool cover -html=coverage.out -o coverage.html

# Build the library
build:
	go build -v .

# Build examples
examples:
	@echo "Building examples..."
	@for dir in examples/*/; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Building $$dir"; \
			go build -o "$$dir/$(notdir $$dir)" "$$dir"; \
		fi \
	done

# Clean build artifacts
clean:
	rm -f *.test
	rm -f coverage.out
	rm -f coverage.html
	rm -f examples/*/$(notdir examples/*/)
	@echo "Cleaned build artifacts"

# Run vet
vet:
	go vet .

# Run staticcheck
staticcheck:
	staticcheck .

# Run all checks
check: vet staticcheck test

# Install dependencies
deps:
	go mod download
	go install honnef.co/go/tools/cmd/staticcheck@latest

# Help
help:
	@echo "Available targets:"
	@echo "  test          - Run tests (excluding examples)"
	@echo "  test-race     - Run tests with race detector"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  build         - Build the library"
	@echo "  examples      - Build all examples"
	@echo "  clean         - Clean build artifacts"
	@echo "  vet           - Run go vet"
	@echo "  staticcheck   - Run staticcheck"
	@echo "  check         - Run vet, staticcheck, and tests"
	@echo "  deps          - Install dependencies"
	@echo "  help          - Show this help"
