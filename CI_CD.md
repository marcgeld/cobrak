# CI/CD Pipeline Documentation

## Overview

The cobrak project uses GitHub Actions for continuous integration and deployment.

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch

**Jobs:**

#### Test Job
- Runs all unit tests with race detection
- Generates coverage report
- Uploads coverage to Codecov

**Steps:**
1. Checkout code
2. Setup Go 1.25.x
3. Cache Go modules
4. Download dependencies
5. Run tests with coverage
6. Upload coverage report

#### Build Job
- Builds the cobrak binary
- Verifies the binary works
- Runs after test job succeeds

**Steps:**
1. Checkout code
2. Setup Go 1.25.x
3. Cache Go modules
4. Build binary
5. Verify binary with `--help`

#### Lint Job
- Runs golangci-lint for code quality
- Checks code style and common issues

**Steps:**
1. Checkout code
2. Setup Go 1.25.x
3. Run golangci-lint

#### Security Job
- Runs Gosec security scanner
- Checks for common security issues

**Steps:**
1. Checkout code
2. Run Gosec scanner

### 2. Release Workflow (`.github/workflows/release.yml`)

**Triggers:**
- Push of tags matching `v*` (e.g., v1.0.0, v1.2.3)

**Jobs:**

#### GoReleaser Job
- Builds binaries for multiple platforms
- Creates GitHub release
- Uploads artifacts

**Platforms:**
- Linux (amd64, arm64)
- macOS/Darwin (amd64, arm64)
- Windows (amd64, arm64)

**Steps:**
1. Checkout code with full history
2. Setup Go 1.24.x
3. Run GoReleaser
4. Generate checksums
5. Create GitHub release

## Status Badges

The README.md includes badges for:
- Go Version
- License
- CI Status (links to CI workflow)
- Build Status

## Local Testing

Before pushing to main, run tests locally:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run linter
golangci-lint run ./...
```

## CI/CD Flow

```
Push to main
    ↓
GitHub Actions CI
    ↓
    ├─→ Run Tests (with race detection)
    ├─→ Build Binary
    ├─→ Lint Code
    └─→ Security Scan
    ↓
All checks pass ✓
    ↓
Ready for release

Tag with v*
    ↓
GitHub Actions Release
    ↓
    ├─→ Build for all platforms
    ├─→ Generate checksums
    ├─→ Create GitHub Release
    └─→ Upload artifacts
```

## Configuration Files

### `.golangci.yml`
Configures golangci-lint with:
- Enabled linters (errcheck, gosimple, govet, etc.)
- Linter-specific settings
- Issue exclusions for test files
- Timeout and directory settings

### `.github/workflows/ci.yml`
Defines the CI workflow:
- When to run (push/PR to main)
- What jobs to execute
- Dependencies between jobs

### `.github/workflows/release.yml`
Defines the release workflow:
- Triggers on version tags
- Uses GoReleaser for multi-platform builds
- Publishes to GitHub Releases

## Caching

Both workflows use Go module caching to speed up builds:
- Caches `~/.cache/go-build`
- Caches `~/go/pkg/mod`
- Key based on `go.sum` hash

## Permissions

### CI Workflow
- `contents: read` - Read repository contents

### Release Workflow
- `contents: write` - Write releases and tags

## Monitoring

View workflow runs at:
- https://github.com/marcgeld/cobrak/actions

Check specific workflows:
- CI: https://github.com/marcgeld/cobrak/actions/workflows/ci.yml
- Release: https://github.com/marcgeld/cobrak/actions/workflows/release.yml

## Troubleshooting

### CI Fails on Tests
1. Run tests locally: `go test ./...`
2. Fix failing tests
3. Push changes

### CI Fails on Linting
1. Run linter locally: `golangci-lint run ./...`
2. Fix linting issues
3. Push changes

### CI Fails on Build
1. Build locally: `go build .`
2. Fix compilation errors
3. Push changes

### Release Fails
1. Check tag format (must be `v*`)
2. Verify .goreleaser.yaml is valid
3. Check GitHub token permissions
4. Review workflow logs

## Best Practices

1. **Always run tests locally** before pushing
2. **Create pull requests** for significant changes
3. **Tag releases** with semantic versioning (v1.0.0)
4. **Monitor CI results** after pushing
5. **Fix issues immediately** if CI fails

## Future Enhancements

Possible improvements:
- Add integration tests
- Add performance benchmarks
- Add Docker image builds
- Add Homebrew formula updates
- Add notification on failures
- Add deployment to package registries

