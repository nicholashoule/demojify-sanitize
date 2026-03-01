.PHONY: all build test race coverage fmt vet lint clean

# Default: format, vet, then test
all: fmt vet test

# Build check (library -- no binary output)
build:
	go build ./...

# Run all tests
test:
	go test ./...

# Run tests with race detector
race:
	go test -race ./...

# Run tests with coverage report
coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format code
fmt:
	gofmt -s -w .

# Vet
vet:
	go vet ./...

# Lint (requires golangci-lint: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

# Remove build artifacts
clean:
	-rm -f coverage.out coverage.html
