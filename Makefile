.PHONY: all help build test race coverage fmt fmt-check vet lint pre-commit hooks driver clean

# Default: format, vet, then test
all: fmt vet test

# List the available targets
help:
	@echo "Targets:"
	@echo "  all         fmt + vet + test (default)"
	@echo "  build       go build ./..."
	@echo "  test        go test ./..."
	@echo "  race        go test -race ./... (needs CGO)"
	@echo "  coverage    test with HTML coverage report"
	@echo "  fmt         gofmt -s -w ."
	@echo "  fmt-check   fail if any file needs gofmt"
	@echo "  vet         go vet ./..."
	@echo "  lint        golangci-lint run ./..."
	@echo "  pre-commit  run the git hook's gate: gofmt + vet + lint + test (cross-platform)"
	@echo "  hooks       install scripts/hooks/pre-commit into .git/hooks/"
	@echo "  driver      build and run the driver example"
	@echo "  clean       remove build artifacts"

# Build check (library -- no binary output)
build:
	go build ./...

# Run all tests
test:
	go test ./...

# Run tests with race detector
# NOTE: The race detector requires CGO (and gcc on Windows).
# On Windows without gcc: set CGO_ENABLED=1 or install gcc via TDM-GCC / MSYS2.
# If CGO is unavailable locally, `make race` will fail -- CI (Linux) runs it on
# every push via .github/workflows/ci.yml.
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

# Check formatting without writing (exits non-zero if any files need
# formatting). Delegates to a Go helper so the target works whether make
# runs recipes via sh or cmd.exe; it only reports (never gofmt -w).
fmt-check:
	go run scripts/fmtcheck.go

# Vet
vet:
	go vet ./...

# Lint (requires golangci-lint: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

# Run the same Go quality gate the git pre-commit hook enforces, by invoking
# the canonical cross-platform implementation directly. It runs gofmt (auto-
# fixing and re-staging any unformatted files, like the installed hook), go
# vet, golangci-lint, and the full test suite. Delegating to the Go program
# (rather than a fmt-check/vet/lint/test recipe chain) keeps this target
# working regardless of whether make runs recipes via sh or cmd.exe.
pre-commit:
	go run scripts/hooks/pre-commit.go

# Install git hooks from scripts/hooks/ into .git/hooks/. Delegates to a Go
# helper so it works without cp/chmod (cross-platform, incl. cmd.exe make).
hooks:
	go run scripts/installhooks.go

# Build and run the driver example program.
# The driver exercises every major public API as a downstream consumer would.
driver:
	go run ./docs/examples/driver/

# Remove build artifacts (cross-platform; missing files are not an error).
clean:
	go run scripts/clean.go
