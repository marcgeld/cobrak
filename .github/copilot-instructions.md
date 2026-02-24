# Copilot Development Instructions for cobrak

## Project Overview

**cobrak** is a Kubernetes cluster analysis CLI tool written in Go that provides:
- Resource usage analysis
- Node health monitoring
- Cluster pressure analysis
- Pod inventory and capacity planning

## Build & Test Requirements

### Go Version
- **Minimum**: Go 1.24
- **Tested**: Go 1.24.x
- Keep `go.mod` aligned with GitHub Actions build environment

### Test Coverage Goal
- **Minimum**: 75% coverage across all packages
- **Current**: 29.3% (growing)
- **Strategy**: Integration tests using `fake.NewSimpleClientset()` for Kubernetes mocking

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt

# Run specific package
go test -v ./pkg/config
go test -v ./pkg/resources
```

## Code Quality Standards

### Code Structure
- **Packages**: Organize by domain (config, k8s, resources, nodeinfo, output)
- **Commands**: Place CLI commands in `cmd/` package
- **Naming**: Use clear, descriptive names following Go conventions

### Complexity & Readability
- **Cognitive Complexity**: Keep functions under 15 complexity score
- **Extract Helpers**: Break complex functions into smaller, testable pieces
- **Error Handling**: Always check and wrap errors with context
- **Comments**: Document public functions and non-obvious logic

Example refactoring pattern:
```go
// Bad: Complex function with multiple concerns
func analyzeNode(node *corev1.Node) {
  // 50+ lines of mixed logic
}

// Good: Simple function with clear helpers
func analyzeNode(node *corev1.Node) NodeAnalysis {
  memPressure := analyzeMemoryPressure(node)
  cpuPressure := analyzeCPUPressure(node)
  health := calculateNodeHealth(memPressure, cpuPressure)
  return NodeAnalysis{...}
}
```

## Testing Strategy

### Test Types

**1. Unit Tests** (For utility functions)
- Test data transformation functions
- Test output formatting
- Test configuration parsing
- Use simple test data, no external dependencies

Example:
```go
func TestRenderTable(t *testing.T) {
  data := []TestStruct{...}
  result := RenderTable(data)
  // Assert result contains expected strings
}
```

**2. Integration Tests** (For Kubernetes interactions)
- Use `fake.NewSimpleClientset()` for mocking K8s API
- Mock metrics using simple interfaces
- Test with realistic pod/node fixtures
- Located in files ending with `_integration_test.go`

Example:
```go
func TestBuildPodSummaries_Integration(t *testing.T) {
  pod := &corev1.Pod{...}
  client := fake.NewSimpleClientset(pod)
  ctx := context.Background()
  
  summaries, err := BuildPodSummaries(ctx, client, "")
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  // Assert summaries are correct
}
```

### Coverage Goals by Package

| Package | Goal | Current | Notes |
|---------|------|---------|-------|
| config | 75%+ | 78.9% | ✅ Exceeds goal |
| kubeconfig | 75%+ | 81.2% | ✅ Exceeds goal |
| nodeinfo | 75%+ | 38.5% | ⚠️ Needs integration tests |
| output | 75%+ | 53.8% | ⚠️ Add render function tests |
| resources | 75%+ | 38.8% | ⚠️ Add pod summary tests |
| k8s | 75%+ | 26.9% | ⚠️ Low coverage, hard to test |
| capacity | 75%+ | 0.0% | ⚠️ No tests yet |
| cmd | 50%+ | 0.0% | ℹ️ CLI tests require complex setup |

### Metrics & Coverage Tools

```bash
# Generate coverage HTML report
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt

# Check coverage per function
go tool cover -func=coverage.txt | grep "0.0%"

# Find functions needing tests
go tool cover -func=coverage.txt | awk '$NF < 75 {print}'
```

## Features & Functionality

### Resource Analysis
- `./cobrak resources` - Show pod resources and totals
- `./cobrak resources simple` - Quick pressure summary
- `./cobrak resources inventory` - Namespace resource coverage
- `./cobrak resources usage` - Actual CPU/memory usage (needs metrics-server)
- `./cobrak resources diff` - Compare usage vs requests

### Flags & Configuration
- `--top=N` - Limit output to top N items (must work on all commands)
- `--output=json|yaml|text` - Output format
- `--namespace=NAME` - Filter by namespace
- `--nocolor` - Disable colored output
- Configuration file: `~/.cobrak/settings.toml`

### Color Support
- Default: Enabled if terminal supports it
- Override: `--nocolor` flag or `NO_COLOR=1` environment variable
- Config: `color = true/false` in settings.toml
- Pressure levels colored: GREEN(LOW), YELLOW(MEDIUM), MAGENTA(HIGH), RED(SATURATED)

## Common Tasks

### Adding a New Command
1. Create `cmd/newcommand.go`
2. Add function `newNewCommandCmd()` returning `*cobra.Command`
3. Add RunE handler `runNewCommand()`
4. Register in appropriate parent command
5. Add integration tests in `pkg/*_integration_test.go`

### Adding Tests
1. For utilities: Create `*_test.go` file
2. For K8s integration: Create `*_integration_test.go` file
3. Use `fake.NewSimpleClientset()` for mocking
4. Use `mock.DeepCopyObject()` for realistic fixtures
5. Run `go test -cover ./...` to verify coverage

### Refactoring for Complexity
1. Measure complexity: Check error message in IDE
2. Extract helpers: Move logic to separate functions
3. Test helpers: Create unit tests for each helper
4. Verify: Ensure complexity drops below 15

Example:
```go
// Reduce complexity by extracting concerns
// Before: 1 function, 55 complexity
// After: 1 main + 3 helpers, each <15 complexity
```

## CI/CD & Release

### GitHub Actions Workflows
- **CI** (`.github/workflows/ci.yml`): Runs on every push to main
    - Tests: `go test ./...`
    - Build: `go build .`
    - Lint: `golangci-lint run ./...`
    - Security: `gosec ./...`

- **Release** (`.github/workflows/release.yml`): Runs on version tags
    - Builds for: Linux, macOS, Windows (amd64, arm64)
    - Generates: Checksums, release notes

### Making a Release
1. Commit changes: `git commit -m "feat: describe changes"`
2. Create tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
3. Push: `git push origin main v1.0.0`
4. GitHub Actions builds and publishes automatically

## Dependencies & Imports

### Core Dependencies
- `k8s.io/client-go` - Kubernetes API client
- `k8s.io/api` - Kubernetes API types
- `k8s.io/metrics` - Metrics API
- `github.com/spf13/cobra` - CLI framework
- `github.com/BurntSushi/toml` - Config parsing
- `github.com/fatih/color` - Color output

### Testing Dependencies
- `k8s.io/client-go/kubernetes/fake` - Fake K8s client
- `testing` - Standard testing

### Import Organization
```go
// Standard library
import (
  "context"
  "fmt"
)

// External libraries
import (
  corev1 "k8s.io/api/core/v1"
  "github.com/spf13/cobra"
)

// Internal packages
import (
  "github.com/marcgeld/cobrak/pkg/config"
  "github.com/marcgeld/cobrak/pkg/k8s"
)
```

## Common Patterns

### Error Handling
```go
func doSomething() error {
  if err != nil {
    return fmt.Errorf("operation failed: %w", err)
  }
  return nil
}
```

### Configuration Management
```go
settings, err := config.LoadSettings()
if err != nil {
  return fmt.Errorf("loading config: %w", err)
}
colorEnabled := settings.Color && !nocolor
output.SetGlobalColorEnabled(colorEnabled)
```

### Kubernetes Client Setup
```go
cfg, err := k8s.NewRestConfig(kubeconfig, kubeCtx)
if err != nil {
  return fmt.Errorf("building rest config: %w", err)
}

client, err := k8s.NewClientFromConfig(cfg)
if err != nil {
  return fmt.Errorf("creating k8s client: %w", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()
```

### Testing with Fake Client
```go
pod := &corev1.Pod{
  ObjectMeta: metav1.ObjectMeta{
    Name:      "test-pod",
    Namespace: "default",
  },
  Spec: corev1.PodSpec{...},
}

client := fake.NewSimpleClientset(pod)
ctx := context.Background()

// Use client in function being tested
result, err := BuildPodSummaries(ctx, client, "")
```

## Linting & Formatting

### Code Style
- Run `go fmt ./...` to format code
- Use goimports for import management
- Follow Go naming conventions (CamelCase, package names lowercase)

### Linting
- Configuration in `.golangci.yml`
- Run via: `golangci-lint run ./...`
- Focus on: errcheck, govet, ineffassign, unused

## Documentation

### README
- Located at root
- Should cover: Features, Installation, Quick Start, Commands, Configuration
- Update whenever: New features added, breaking changes made, version bumped

### Code Comments
- Public functions: Always document purpose
- Complex logic: Explain the "why"
- Private functions: Add if logic is non-obvious

Example:
```go
// RenderTable formats data as a tab-separated table.
// It limits output to 'top' items if top > 0.
func RenderTable(data []DataItem, top int) string {
  if top > 0 && len(data) > top {
    data = data[:top]  // Limit to top N items
  }
  // ... rest of implementation
}
```

## Version Management

### Setting Version
- Edit `go.mod`: Current language version (1.24)
- Version flags set via ldflags in build
- GoReleaser handles version injection automatically

### Version Command
```bash
./cobrak version
# Output:
# cobrak version v1.0.0
# commit: abc123def
# date: 2026-02-24T15:30:00Z
```

## Architecture Decisions

### Modular Organization
- `pkg/config` - Settings and configuration
- `pkg/k8s` - Kubernetes API utilities
- `pkg/kubeconfig` - Kubeconfig parsing
- `pkg/nodeinfo` - Node system information
- `pkg/output` - Output formatting and rendering
- `pkg/resources` - Pod resource analysis
- `pkg/capacity` - Cluster capacity calculation

### Output Design
- Supports multiple formats: text, JSON, YAML
- Uses same data structures for all formats
- Rendering decoupled from data collection

### Color Support
- Global flag-based control
- Respects `NO_COLOR` environment variable
- Configuration file override
- Terminal capability detection

## Performance Considerations

- **Timeouts**: 20-30 seconds per operation (avoid hanging)
- **Memory**: ~50MB typical usage
- **API Calls**: Minimize Kubernetes API calls (use caching if needed)
- **Parallelization**: Consider for large clusters, but keep simple

---

**Last Updated**: February 24, 2026
**Version**: 1.0.0
**Go Version**: 1.24+
**Test Coverage Target**: 75%+

