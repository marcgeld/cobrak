# 🔍 cobrak

[![Go Version](https://img.shields.io/badge/go-1.25.0+-blue.svg)](https://golang.org/doc/devel/release)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/status-stable-brightgreen.svg)](#)

A modular, lightweight, and fast analytical CLI tool for inspecting and analyzing Kubernetes cluster state, resource usage, node health, and capacity planning.

**cobrak** helps DevOps engineers and cluster administrators quickly identify resource constraints, node health issues, pressure points, and capacity problems in their Kubernetes clusters.

## ✨ Features

### 📊 Resource Analysis
- **Pod-level resource details** - CPU/Memory requests and limits per pod
- **Cluster capacity summaries** - Total CPU and memory allocatable/capacity
- **Resource inventories** - Namespace-wide resource coverage and missing requests/limits
- **Usage tracking** - Actual CPU/Memory usage per container (requires metrics-server)
- **Usage diffs** - Compare actual usage vs. requested resources to find waste

### ⚡ Quick Pressure Summary
- **Cluster pressure levels** - Overall cluster resource pressure (LOW, MEDIUM, HIGH, SATURATED)
- **Per-node pressure** - CPU and memory saturation per node
- **Per-namespace pressure** - Resource utilization by namespace
- **One-liner format** - Quick status overview at a glance

### 🖥️ Node Information
- **OS & Kernel details** - Operating system, kernel version, architecture
- **CPU information** - CPU model, core count, capacity
- **GPU detection** - Identifies NVIDIA and AMD GPUs
- **Memory pressure** - Total memory, usage, utilization percentage, pressure level
- **Filesystem latency** - Root FS latency, inode usage, disk pressure
- **Container runtime** - Docker, containerd, CRI-O detection with versions
- **Virtualization detection** - AWS EC2, GCP, Azure, VMware, KVM, Bare Metal, etc.
- **Health status** - Overall node health with issue detection

### ⚙️ Node Capacity
- **Per-node capacity** - CPU and memory allocatable/capacity for each node
- **Simple overview** - Quick capacity view by node

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/marcgeld/cobrak.git
cd cobrak

# Build the binary
go build -o cobrak

# Optional: Install to $GOPATH/bin
go install
```

### Basic Usage

```bash
# Show resource summary for all pods
./cobrak resources

# Show quick cluster pressure summary
./cobrak resources simple

# Show detailed node information
./cobrak nodeinfo

# Show node capacity overview
./cobrak capacity

# Get help for any command
./cobrak --help
./cobrak resources --help
./cobrak nodeinfo --help
```

## 📚 Commands

### `cobrak resources`

Analyze pod resources and cluster capacity.

```bash
# Default: show pod resource details with totals
./cobrak resources

# Show quick pressure summary
./cobrak resources simple

# Show namespace resource inventory
./cobrak resources inventory

# Show actual CPU/Memory usage (requires metrics-server)
./cobrak resources usage

# Compare usage vs. requests/limits
./cobrak resources diff

# Filter by namespace
./cobrak resources --namespace=production

# Show top 50 offenders
./cobrak resources --top=50

# JSON output
./cobrak resources --output=json
```

#### Output Example (default):
```
=== CLUSTER CAPACITY SUMMARY ===
CPU Capacity:          12
CPU Allocatable:       11.85
CPU Requests:          1.53
CPU Limits:            0.25

Memory Capacity:       32Gi
Memory Allocatable:    28Gi
Memory Requests:       4Gi
Memory Limits:         8Gi

=== POD RESOURCE DETAILS ===
NAMESPACE    POD              CPU REQUEST  CPU LIMIT   MEM REQUEST   MEM LIMIT
default      nginx-1          500m         1           256Mi         512Mi
default      postgres-1       1000m        2           1Gi           2Gi

=== TOTALS ===
Total CPU Usage:       2100m
Total CPU Requests:    1700m
Total CPU Limits:      3700m
```

#### Output Example (simple):
```
Cluster Pressure: LOW
Node worker-1: CPU SATURATED (95%)
Node worker-2: Memory HIGH (82%)
Namespace monitoring: CPU 72% requested
Namespace production: Memory 85% requested
```

### `cobrak nodeinfo`

Get detailed system information about nodes.

```bash
# Show detailed info for all nodes
./cobrak nodeinfo

# Compact format (table)
./cobrak nodeinfo --compact

# Show specific node
./cobrak nodeinfo --node=worker-1

# Health status only
./cobrak nodeinfo --health

# Specific node health status
./cobrak nodeinfo --node=worker-1 --health
```

#### Output Example (detailed):
```
Node: worker-1
  OS: linux
  Kernel: 5.15.0-56-generic
  Architecture: amd64
  Kubelet Version: v1.26.0

  CPU Information:
    Model: Intel Xeon
    Cores: 8
    Capacity: 8000m

  GPU Information:
    Available: Yes (1 GPU(s))
      - nvidia-0: tesla-v100

  Memory Pressure:
    Total: 15.69 GB
    Utilization: 65.3%
    Pressure: MEDIUM

  Container Runtime:
    Name: containerd
    Version: 1.6.8

  Virtualization:
    Type: AWS EC2
```

#### Output Example (compact):
```
NODE | OS | ARCH | CPU | GPU | MEM | RUNTIME | VIRTUALIZATION
worker-1 | linux | amd64 | 8c | Yes(1) | MEDIUM | containerd | AWS EC2
worker-2 | linux | amd64 | 4c | No | LOW | docker | GCP
```

### `cobrak capacity`

Show CPU and memory capacity for each node.

```bash
# Show capacity for all nodes
./cobrak capacity

# With specific context
./cobrak capacity --context=production-cluster
```

## 🎯 Use Cases

### Resource Planning
```bash
# Identify over-provisioned pods
./cobrak resources diff --top=20

# Find pods without resource limits
./cobrak resources --namespace=production
```

### Cluster Health Checks
```bash
# Quick cluster health overview
./cobrak resources simple

# Detailed node health status
./cobrak nodeinfo --health

# CPU/Memory saturation per node
./cobrak nodeinfo --compact
```

### Capacity Planning
```bash
# See current capacity utilization
./cobrak capacity

# Get detailed pod resource usage
./cobrak resources

# Identify memory-pressured nodes
./cobrak nodeinfo --health
```

## 🔧 Requirements

- **Go 1.25.0+** - To build from source
- **Kubernetes cluster 1.24+** - Target cluster
- **kubectl** - Configured kubeconfig
- **metrics-server** (optional) - For `usage` and `diff` commands

## 📋 Configuration

### Kubeconfig
cobrak respects the standard Kubernetes kubeconfig configuration:

```bash
# Use specific kubeconfig
./cobrak resources --kubeconfig=/path/to/config.yaml

# Use KUBECONFIG environment variable
export KUBECONFIG=$HOME/.kube/config
./cobrak nodeinfo

# Use specific context
./cobrak resources --context=my-cluster
```

### Filtering
```bash
# Analyze specific namespace
./cobrak resources --namespace=kube-system

# Show all namespaces (default)
./cobrak resources --all-namespaces

# Show top N resources
./cobrak resources --top=50
```

## 🏗️ Architecture

cobrak is organized into modular packages:

```
pkg/
├── capacity/        # Node capacity analysis and pressure calculation
├── k8s/            # Kubernetes API client utilities
├── kubeconfig/     # Kubeconfig resolution
├── nodeinfo/       # Node system information
├── output/         # Output rendering and formatting
└── resources/      # Pod resources, inventory, usage, and diff analysis

cmd/
├── root.go         # Root command
├── capacity.go     # Capacity command
├── nodeinfo.go     # Node info command
└── resources.go    # Resources command and subcommands
```

## 📤 Output Formats

cobrak supports **multiple output formats** for easy integration and automation:

### Text Format (Default)
```bash
./cobrak resources --output=text
```
Human-readable table format with clear sections and summaries.

### JSON Format
```bash
./cobrak resources --output=json
```
Structured JSON output perfect for automation, scripting, and tool integration.

### YAML Format
```bash
./cobrak resources --output=yaml
```
YAML format ideal for configuration management and GitOps workflows.

**Example with jq filtering:**
```bash
# Get high-memory pods
./cobrak resources --output=json | jq '.pod_details[] | select(.mem_request | endswith("Gi"))'

# Count pods per namespace
./cobrak resources --output=json | jq '.pod_details | group_by(.namespace) | map({namespace: .[0].namespace, count: length})'
```

## 🏗️ Architecture

cobrak is organized into modular packages:

```
pkg/
├── capacity/        # Node capacity analysis and pressure calculation
├── k8s/            # Kubernetes API client utilities
├── kubeconfig/     # Kubeconfig resolution
├── nodeinfo/       # Node system information
├── output/         # Output rendering and formatting
└── resources/      # Pod resources, inventory, usage, and diff analysis

cmd/
├── root.go         # Root command
├── capacity.go     # Capacity command
├── nodeinfo.go     # Node info command
└── resources.go    # Resources command and subcommands
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test -v ./pkg/nodeinfo/...
```

Current test coverage includes:
- 21+ unit tests across all packages
- Mock Kubernetes client tests
- Output formatting tests
- Node analysis tests

## 🐛 Troubleshooting

### No metrics available
If `usage` and `diff` commands show no data, ensure metrics-server is installed:
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

### Kubeconfig errors
Ensure your kubeconfig is valid and accessible:
```bash
kubectl cluster-info  # Test connectivity
echo $KUBECONFIG      # Check env var
```

### Permission issues
Some operations require specific RBAC permissions:
```bash
# Check your current permissions
kubectl auth can-i get nodes
kubectl auth can-i get pods --all-namespaces
```

## 📊 Performance

- **Fast startup** - Minimal dependencies
- **Low memory usage** - ~50MB typical
- **Efficient API calls** - Minimal cluster load
- **Timeout protection** - 20-30 second timeout per operation

## 🤝 Contributing

Contributions are welcome! Please feel free to submit Pull Requests.

### Development Setup
```bash
git clone https://github.com/marcgeld/cobrak.git
cd cobrak
go mod download
go build -o cobrak
./cobrak --help
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Marcus Geld**
- GitHub: [@marcgeld](https://github.com/marcgeld)

## 🔗 Related Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [metrics-server](https://github.com/kubernetes-sigs/metrics-server)
- [kubectl](https://kubernetes.io/docs/reference/kubectl/)

## 📬 Support

For issues, feature requests, or questions:
1. Open an [issue](https://github.com/marcgeld/cobrak/issues)
2. Check existing [discussions](https://github.com/marcgeld/cobrak/discussions)
3. Review the [documentation](#-commands)

---

Made with ❤️for Kubernetes cluster operators
