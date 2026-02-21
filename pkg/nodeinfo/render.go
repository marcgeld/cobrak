package nodeinfo

import (
	"fmt"
	"strings"
)

// RenderNodeInfo renders detailed node information
func RenderNodeInfo(info *NodeInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Node: %s\n", info.NodeName))
	sb.WriteString(fmt.Sprintf("  OS: %s\n", info.OS))
	sb.WriteString(fmt.Sprintf("  Kernel: %s\n", info.Kernel))
	sb.WriteString(fmt.Sprintf("  Architecture: %s\n", info.Architecture))
	sb.WriteString(fmt.Sprintf("  Kubelet Version: %s\n\n", info.KubeletVersion))

	// CPU Info
	sb.WriteString("  CPU Information:\n")
	sb.WriteString(fmt.Sprintf("    Model: %s\n", info.CPU.Model))
	sb.WriteString(fmt.Sprintf("    Cores: %d\n", info.CPU.Count))
	sb.WriteString(fmt.Sprintf("    Capacity: %dm\n\n", info.CPU.Capacity))

	// GPU Info
	sb.WriteString("  GPU Information:\n")
	if info.GPU.Available {
		sb.WriteString(fmt.Sprintf("    Available: Yes (%d GPU(s))\n", len(info.GPU.GPUs)))
		for _, gpu := range info.GPU.GPUs {
			sb.WriteString(fmt.Sprintf("      - %s: %s\n", gpu.Index, gpu.Model))
		}
	} else {
		sb.WriteString("    Available: No\n")
	}
	sb.WriteString("\n")

	// Memory Pressure
	sb.WriteString("  Memory Pressure:\n")
	sb.WriteString(fmt.Sprintf("    Total: %.2f GB\n", float64(info.MemoryPressure.Total)/(1024*1024*1024)))
	sb.WriteString(fmt.Sprintf("    Utilization: %.1f%%\n", info.MemoryPressure.UtilizationRatio*100))
	sb.WriteString(fmt.Sprintf("    Pressure: %s\n", info.MemoryPressure.Pressure))
	if info.MemoryPressure.PageCacheRatio > 0 {
		sb.WriteString(fmt.Sprintf("    Page Cache Ratio: %.1f%%\n", info.MemoryPressure.PageCacheRatio*100))
	}
	sb.WriteString("\n")

	// Filesystem Latency
	sb.WriteString("  Filesystem:\n")
	sb.WriteString(fmt.Sprintf("    Root FS Latency: %dms\n", info.FilesystemLatency.RootFSLatency))
	sb.WriteString(fmt.Sprintf("    Root FS Inodes Used: %.1f%%\n", info.FilesystemLatency.RootFSInodesUsed))
	sb.WriteString(fmt.Sprintf("    Root FS Capacity Used: %.1f%%\n\n", info.FilesystemLatency.RootFSCapacityUsed))

	// Container Runtime
	sb.WriteString("  Container Runtime:\n")
	sb.WriteString(fmt.Sprintf("    Name: %s\n", info.ContainerRuntime.Name))
	sb.WriteString(fmt.Sprintf("    Version: %s\n\n", info.ContainerRuntime.Version))

	// Virtualization
	sb.WriteString("  Virtualization:\n")
	sb.WriteString(fmt.Sprintf("    Type: %s\n", info.VirtualizationType))

	return strings.TrimRight(sb.String(), "\n")
}

// RenderNodeInfoCompact renders a compact version of node information
func RenderNodeInfoCompact(info *NodeInfo) string {
	var sb strings.Builder

	gpuStatus := "No"
	if info.GPU.Available {
		gpuStatus = fmt.Sprintf("Yes (%d)", len(info.GPU.GPUs))
	}

	sb.WriteString(fmt.Sprintf("%s | %s | %s | CPU:%dc | GPU:%s | Mem:%s | Runtime:%s | Virt:%s\n",
		info.NodeName,
		info.OS,
		info.Architecture,
		info.CPU.Count,
		gpuStatus,
		info.MemoryPressure.Pressure,
		info.ContainerRuntime.Name,
		info.VirtualizationType,
	))

	return strings.TrimRight(sb.String(), "\n")
}

// RenderNodeHealth renders node health status
func RenderNodeHealth(status *NodeHealthStatus) string {
	var sb strings.Builder

	statusSymbol := "✓"
	if status.Status == "WARNING" {
		statusSymbol = "⚠"
	} else if status.Status == "CRITICAL" {
		statusSymbol = "✗"
	}

	sb.WriteString(fmt.Sprintf("%s Node: %s [%s]\n", statusSymbol, status.NodeName, status.Status))

	if len(status.Issues) > 0 {
		sb.WriteString("  Issues:\n")
		for _, issue := range status.Issues {
			sb.WriteString(fmt.Sprintf("    - %s\n", issue))
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

		sb.WriteString(fmt.Sprintf("%s | %s | %s | %dc | %s | %s | %s | %s\n",
			info.NodeName,
			info.OS,
			info.Architecture,
			info.CPU.Count,
			gpuStatus,
			info.MemoryPressure.Pressure,
			info.ContainerRuntime.Name,
			info.VirtualizationType,
		))
	}

	return strings.TrimRight(sb.String(), "\n")
}
