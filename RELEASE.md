# Release & Version Management

## Overview

Cobrak is configured for professional release management with:
- ✅ Version information injected at build time
- ✅ GoReleaser configuration for multi-platform builds
- ✅ Makefile for easy build and release workflows
- ✅ Cosign signing support (optional)
- ✅ Checksum generation

---

## Version Information

Version, commit, and build date are set at compile time via ldflags:

```bash
# Development build (defaults)
go build -o cobrak main.go
# cobrak version dev
# commit: unknown
# date: unknown

# Release build with version info
go build \
  -X main.version=v1.0.0 \
  -X main.commit=abc123 \
  -X main.date=2025-02-21T10:30:00Z \
  -o cobrak main.go
```

---

## Building

### Local Development Build

```bash
# Using Makefile
make build

# Manual build
go build -o cobrak main.go

# Check version
./cobrak version
```

### Production Build with Version

```bash
# Using Makefile (automatically extracts from git)
make build

# Manual build with git-derived version
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

go build \
  -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
  -o cobrak main.go
```

---

## Release Process with GoReleaser

### Prerequisites

```bash
# Install GoReleaser
brew install goreleaser

# Optional: Install Cosign for signing
brew install cosign

# Create git tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

### Dry-Run Release (Recommended First)

```bash
# Build all artifacts without uploading
make release-dry

# Or manually
goreleaser release --snapshot --clean --skip=publish,docker,formula

# Outputs to: ./dist/
```

### Full Release

```bash
# Set up GitHub token (if releasing to GitHub)
export GITHUB_TOKEN=your_token_here

# Run release
make release

# Or manually
goreleaser release --clean
```

### Release Artifacts

Generated in `./dist/`:
- `cobrak_v1.0.0_linux_amd64.tar.gz`
- `cobrak_v1.0.0_linux_arm64.tar.gz`
- `cobrak_v1.0.0_darwin_amd64.tar.gz`
- `cobrak_v1.0.0_darwin_arm64.tar.gz`
- `cobrak_v1.0.0_windows_amd64.zip`
- `cobrak_v1.0.0_windows_arm64.zip`
- `cobrak_v1.0.0_checksums.txt`

---

## GoReleaser Configuration

File: `.goreleaser.yaml`

### Build Configuration

- **Main binary:** `./main.go`
- **Output name:** `cobrak`
- **Platforms:** Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64)
- **CGO disabled:** Static binaries
- **Stripped:** Removes debug symbols (-s -w)

### Archive Format

- **Linux/macOS:** `.tar.gz` with directory wrapper
- **Windows:** `.zip` with directory wrapper

### Checksum Generation

- **Algorithm:** SHA256
- **File:** `cobrak_v1.0.0_checksums.txt`

### Release Notes

- **Draft:** false (auto-published)
- **Prerelease:** auto-detected from version
- **Changelog:** Sorted ascending, excludes docs/test/ci

### Optional Features

#### Docker Images

Configured but disabled by default. To enable:

```yaml
dockers:
  - image_templates:
      - 'marcgeld/cobrak:latest'
      - 'marcgeld/cobrak:{{ .Tag }}'
    skip_push: false
```

Then build and push:
```bash
docker buildx build --push -t marcgeld/cobrak .
```

#### Homebrew Formula

Configured but requires separate tap repository. To enable:

1. Create `homebrew-cobrak` repository
2. Set `skip_upload: false`

#### Code Signing (Cosign)

Configured but disabled by default. To enable:

```bash
# Generate cosign keys
cosign generate-key-pair

# Enable signing in .goreleaser.yaml
signs:
  - artifacts: checksum
    cmd: cosign
    args:
      - 'sign-blob'
      - '--key=cosign.key'
      - '${artifact}'
    output: true

# Make release
make release-sign
```

---

## Makefile Commands

### Build & Install

```bash
make build              # Build cobrak binary
make install           # Install to $GOPATH/bin
make clean             # Remove build artifacts
```

### Testing

```bash
make test              # Run all tests
make test-coverage     # Run tests with coverage report
```

### Release

```bash
make version           # Show current version
make release-dry       # Dry-run release process
make release           # Build full release
make release-sign      # Sign checksums with cosign
```

### Code Quality

```bash
make fmt               # Format code
make lint              # Run linter (requires golangci-lint)
```

---

## Version Command

```bash
./cobrak version
# Output:
# cobrak version v1.0.0
# commit: abc123def456
# date: 2025-02-21T10:30:00Z
```

---

## GitHub Actions Workflow (Optional)

Example `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - uses: actions/setup-go@v4
        with:
          go-version: 1.25
      
      - uses: goreleaser/goreleaser-action@v4
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Verification

### Verify Checksum

```bash
# Download artifact and checksum
wget https://github.com/marcgeld/cobrak/releases/download/v1.0.0/cobrak_v1.0.0_linux_amd64.tar.gz
wget https://github.com/marcgeld/cobrak/releases/download/v1.0.0/cobrak_v1.0.0_checksums.txt

# Verify
sha256sum -c cobrak_v1.0.0_checksums.txt | grep linux_amd64

# Expected output:
# cobrak_v1.0.0_linux_amd64.tar.gz: OK
```

### Verify Cosign Signature (if enabled)

```bash
# Download signature
wget https://github.com/marcgeld/cobrak/releases/download/v1.0.0/cobrak_v1.0.0_checksums.txt.sig

# Verify with public key
cosign verify-blob \
  --key cosign.pub \
  --signature cobrak_v1.0.0_checksums.txt.sig \
  cobrak_v1.0.0_checksums.txt
```

---

## Release Strategy

### Semantic Versioning

Version format: `vMAJOR.MINOR.PATCH`

- **MAJOR:** Breaking changes
- **MINOR:** New features (backwards compatible)
- **PATCH:** Bug fixes

### Pre-release Versions

- `v1.0.0-alpha` - Early development
- `v1.0.0-beta` - Feature complete, testing
- `v1.0.0-rc1` - Release candidate

### Tagging

```bash
# Create version tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push to remote
git push origin v1.0.0

# Trigger release workflow
# (GitHub Actions will run automatically)
```

---

## CI/CD Integration

### GitHub Actions

GoReleaser Action automatically triggers on version tags.

### Custom CI/CD

```bash
# In your CI/CD pipeline
GITHUB_TOKEN=$CI_TOKEN goreleaser release --clean
```

### Local Development

```bash
# Test release process locally
make release-dry

# Verify artifacts
ls -lah dist/
```

---

## Structure for Multiple Binaries (Future)

Currently configured for single binary (`cobrak`). To add multiple binaries:

```yaml
builds:
  - id: cobrak
    main: ./main.go
    binary: cobrak
  
  - id: cobrak-agent
    main: ./cmd/agent/main.go
    binary: cobrak-agent
```

---

## Troubleshooting

### Build fails with "version not found"

**Solution:** Version variables default to "dev". For CI/CD, set via ldflags.

### GoReleaser not installing dependencies

```bash
# Add go.mod tidy to before hooks in .goreleaser.yaml
before:
  hooks:
    - go mod tidy
    - go generate ./...
```

### GitHub token issues

```bash
# Ensure token has repo access
export GITHUB_TOKEN=github_pat_xxxxx
goreleaser release --clean
```

### Checksum verification fails

```bash
# Ensure you're using the correct checksum file
# Check the tag/release matches

sha256sum -c cobrak_v1.0.0_checksums.txt
```

---

## Best Practices

✅ Always test with `make release-dry` first  
✅ Use semantic versioning for versions  
✅ Write meaningful release notes  
✅ Sign releases with cosign for production  
✅ Verify checksums before distribution  
✅ Keep .goreleaser.yaml in version control  
✅ Document breaking changes in release notes  
✅ Tag releases with git for traceability  

---

## Next Steps

1. **Local Testing**
   ```bash
   make build
   ./cobrak version
   make test
   ```

2. **Dry-Run Release**
   ```bash
   make release-dry
   ls dist/
   ```

3. **Setup GitHub Release (Optional)**
   - Create GitHub Personal Access Token
   - Export GITHUB_TOKEN
   - Push version tag
   - Run `make release`

4. **Setup Cosign Signing (Optional)**
   - Generate cosign keypair: `cosign generate-key-pair`
   - Enable signing in .goreleaser.yaml
   - Verify signatures: `cosign verify-blob ...`

---

## References

- [GoReleaser Documentation](https://goreleaser.com/)
- [Cosign Documentation](https://docs.sigstore.dev/cosign/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Releases API](https://docs.github.com/en/rest/releases)

