.PHONY: help build test clean release release-dry-sign version

# Variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

help:
	@echo "cobrak - Kubernetes cluster analysis CLI"
	@echo ""
	@echo "Available targets:"
	@echo "  make build              - Build cobrak binary"
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make release            - Build release with GoReleaser"
	@echo "  make release-dry        - Dry run of release process"
	@echo "  make release-sign       - Sign checksums with cosign"
	@echo "  make version            - Show current version"
	@echo "  make install            - Install cobrak locally"
	@echo ""

build:
	@echo "Building cobrak $(VERSION)..."
	go build -o cobrak $(LDFLAGS) ./main.go
	@echo "✓ Binary built: ./cobrak"

install:
	@echo "Installing cobrak $(VERSION)..."
	go install $(LDFLAGS) ./main.go
	@echo "✓ Installation complete"

test:
	@echo "Running tests..."
	go test ./... -v

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

clean:
	@echo "Cleaning build artifacts..."
	rm -f cobrak coverage.out coverage.html
	rm -rf dist/
	@echo "✓ Clean complete"

version:
	@echo "cobrak version $(VERSION)"
	@echo "commit: $(COMMIT)"
	@echo "date: $(DATE)"

release-dry:
	@echo "Running GoReleaser in dry-run mode..."
	goreleaser release --snapshot --clean --skip=publish,docker,formula

release:
	@echo "Building release for $(VERSION)..."
	goreleaser release --clean
	@echo "✓ Release complete"

release-sign:
	@echo "Signing checksums with cosign..."
	cosign sign-blob --key cosign.key dist/cobrak_*_checksums.txt > dist/cobrak_*_checksums.txt.sig
	@echo "✓ Checksums signed"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Format complete"

lint:
	@echo "Running linter..."
	golangci-lint run ./...
	@echo "✓ Lint complete"

.DEFAULT_GOAL := help

