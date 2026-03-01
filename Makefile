.PHONY: all build test race coverage fmt fmt-check vet lint hooks clean

# Default: format, vet, then test
all: fmt vet test

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

# Check formatting without writing (exits non-zero if any files need formatting)
fmt-check:
	@out=$$(gofmt -s -l .); if [ -n "$$out" ]; then echo "Unformatted files:\n$$out"; exit 1; fi

# Vet
vet:
	go vet ./...

# Lint (requires golangci-lint: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

# Install git hooks from scripts/hooks/ into .git/hooks/
# chmod is a no-op on Windows but harmless; the sh wrapper runs via
# Git for Windows' bundled sh.exe on all platforms.
hooks:
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit 2>/dev/null || true
	@echo "[PASS] pre-commit hook installed"

# Remove build artifacts
clean:
	-rm -f coverage.out coverage.html
