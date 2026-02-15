.PHONY: help build test test-coverage fmt vet lint clean install

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the noms binary"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format all Go files"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run golangci-lint"
	@echo "  clean         - Remove build artifacts"
	@echo "  install       - Install noms binary to GOPATH/bin"
	@echo "  ci            - Run all CI checks (fmt, vet, lint, test)"

# Build the binary
build:
	go build -o noms cmd/main.go

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	gofmt -s -w .

# Run go vet
vet:
	go vet ./...

# Run golangci-lint (requires golangci-lint to be installed)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

# Clean build artifacts
clean:
	rm -f noms
	rm -f coverage.out coverage.html
	rm -f noms-*

# Install to GOPATH/bin
install:
	go install cmd/main.go

# Run all CI checks
ci: fmt vet test
	@echo "All CI checks passed!"
