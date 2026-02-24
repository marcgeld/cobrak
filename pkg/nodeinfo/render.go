package nodeinfo

import (
	"fmt"
	"strings"
)

// RenderNodeInfo renders detailed node information
func RenderNodeInfo(info *NodeInfo) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Node: %s\n", info.NodeName)
	fmt.Fprintf(&sb, "  OS: %s\n", info.OS)
	fmt.Fprintf(&sb, "  Kernel: %s\n", info.Kernel)
	fmt.Fprintf(&sb, "  Architecture: %s\n", info.Architecture)
	fmt.Fprintf(&sb, "  Kubelet Version: %s\n\n", info.KubeletVersion)

	// CPU Info
	sb.WriteString("  CPU Information:\n")
	fmt.Fprintf(&sb, "    Model: %s\n", info.CPU.Model)
	fmt.Fprintf(&sb, "    Cores: %d\n", info.CPU.Count)
	fmt.Fprintf(&sb, "    Capacity: %dm\n\n", info.CPU.Capacity)

	// GPU Info
	sb.WriteString("  GPU Information:\n")
	if info.GPU.Available {
		fmt.Fprintf(&sb, "    Available: Yes (%d GPU(s))\n", len(info.GPU.GPUs))
		for _, gpu := range info.GPU.GPUs {
			fmt.Fprintf(&sb, "      - %s: %s\n", gpu.Index, gpu.Model)
		}
	} else {
		sb.WriteString("    Available: No\n")
	}
	sb.WriteString("\n")

	// Memory Pressure
	sb.WriteString("  Memory Pressure:\n")
	fmt.Fprintf(&sb, "    Total: %.2f GB\n", float64(info.MemoryPressure.Total)/(1024*1024*1024))
	fmt.Fprintf(&sb, "    Utilization: %.1f%%\n", info.MemoryPressure.UtilizationRatio*100)
	fmt.Fprintf(&sb, "    Pressure: %s\n", info.MemoryPressure.Pressure)
	if info.MemoryPressure.PageCacheRatio > 0 {
		fmt.Fprintf(&sb, "    Page Cache Ratio: %.1f%%\n", info.MemoryPressure.PageCacheRatio*100)
	}
	sb.WriteString("\n")

	// Filesystem Latency
	sb.WriteString("  Filesystem:\n")
	fmt.Fprintf(&sb, "    Root FS Latency: %dms\n", info.FilesystemLatency.RootFSLatency)
	fmt.Fprintf(&sb, "    Root FS Inodes Used: %.1f%%\n", info.FilesystemLatency.RootFSInodesUsed)
	fmt.Fprintf(&sb, "    Root FS Capacity Used: %.1f%%\n\n", info.FilesystemLatency.RootFSCapacityUsed)

	// Container Runtime
	sb.WriteString("  Container Runtime:\n")
	fmt.Fprintf(&sb, "    Name: %s\n", info.ContainerRuntime.Name)
	fmt.Fprintf(&sb, "    Version: %s\n\n", info.ContainerRuntime.Version)

	// Virtualization
	sb.WriteString("  Virtualization:\n")
	fmt.Fprintf(&sb, "    Type: %s\n", info.VirtualizationType)

	return strings.TrimRight(sb.String(), "\n")
}

// RenderNodeInfoCompact renders a compact version of node information
func RenderNodeInfoCompact(info *NodeInfo) string {
	var sb strings.Builder

	gpuStatus := "No"
	if info.GPU.Available {
		gpuStatus = fmt.Sprintf("Yes (%d)", len(info.GPU.GPUs))
	}

	fmt.Fprintf(&sb, "%s | %s | %s | CPU:%dc | GPU:%s | Mem:%s | Runtime:%s | Virt:%s\n",
		info.NodeName,
		info.OS,
		info.Architecture,
		info.CPU.Count,
		gpuStatus,
		info.MemoryPressure.Pressure,
		info.ContainerRuntime.Name,
		info.VirtualizationType,
	)

	return strings.TrimRight(sb.String(), "\n")
}

// RenderNodeHealth renders node health status
func RenderNodeHealth(status *NodeHealthStatus) string {
	var sb strings.Builder

	var statusSymbol string
	switch status.Status {
	case "WARNING":
		statusSymbol = "⚠"
	case "CRITICAL":
		statusSymbol = "✗"
	default:
		statusSymbol = "✓"
	}

	fmt.Fprintf(&sb, "%s Node: %s [%s]\n", statusSymbol, status.NodeName, status.Status)

	if len(status.Issues) > 0 {
		sb.WriteString("  Issues:\n")
		for _, issue := range status.Issues {
			fmt.Fprintf(&sb, "    - %s\n", issue)
		}
	} else {
		sb.WriteString("  No issues detected\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// RenderMultipleNodeInfoCompact renders multiple nodes in compact format
func RenderMultipleNodeInfoCompact(infos []NodeInfo) string {
	if len(infos) == 0 {
		return "No nodes found."
	}

	var sb strings.Builder
	sb.WriteString("NODE | OS | ARCH | CPU | GPU | MEM | RUNTIME | VIRTUALIZATION\n")
	sb.WriteString(strings.Repeat("-", 100) + "\n")

	for _, info := range infos {
		gpuStatus := "No"
		if info.GPU.Available {
			gpuStatus = fmt.Sprintf("Yes(%d)", len(info.GPU.GPUs))
		}

		fmt.Fprintf(&sb, "%s | %s | %s | %dc | %s | %s | %s | %s\n",
			info.NodeName,
			info.OS,
			info.Architecture,
			info.CPU.Count,
			gpuStatus,
			info.MemoryPressure.Pressure,
			info.ContainerRuntime.Name,
			info.VirtualizationType,
		)
	}

	return strings.TrimRight(sb.String(), "\n")
}
