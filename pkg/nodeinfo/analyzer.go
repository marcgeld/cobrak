package nodeinfo

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// AnalyzeNode gathers comprehensive node information from Kubernetes API
func AnalyzeNode(ctx context.Context, client kubernetes.Interface, nodeName string) (*NodeInfo, error) {
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting node: %w", err)
	}

	info := &NodeInfo{
		NodeName:       node.Name,
		OS:             node.Status.NodeInfo.OperatingSystem,
		Kernel:         node.Status.NodeInfo.KernelVersion,
		Architecture:   node.Status.NodeInfo.Architecture,
		KubeletVersion: node.Status.NodeInfo.KubeletVersion,
	}

	// Extract CPU info from node status
	info.CPU = extractCPUInfo(node)

	// Extract GPU info from node labels
	info.GPU = extractGPUInfo(node)

	// Extract container runtime
	info.ContainerRuntime = extractContainerRuntime(node)

	// Detect virtualization
	info.VirtualizationType = detectVirtualization(node)

	// Initialize memory pressure with data from node status
	info.MemoryPressure = analyzeMemoryPressure(node)

	// Initialize filesystem latency
	info.FilesystemLatency = analyzeFilesystemLatency(node)

	return info, nil
}

// AnalyzeAllNodes analyzes all nodes in the cluster
func AnalyzeAllNodes(ctx context.Context, client kubernetes.Interface) ([]NodeInfo, error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	var nodeInfos []NodeInfo
	for _, node := range nodes.Items {
		info, err := AnalyzeNode(ctx, client, node.Name)
		if err != nil {
			// Log but continue with other nodes
			continue
		}
		nodeInfos = append(nodeInfos, *info)
	}

	return nodeInfos, nil
}

// extractCPUInfo extracts CPU information from node
func extractCPUInfo(node *corev1.Node) CPUInfo {
	cpuInfo := CPUInfo{
		Model: "Unknown",
		Count: 0,
	}

	// Try to get CPU model from system-info
	if node.Status.NodeInfo.MachineID != "" {
		cpuInfo.Model = extractCPUModel(node.Status.NodeInfo.MachineID)
	}

	// Get CPU capacity
	if cpu, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
		cpuInfo.Count = int(cpu.Value())
		cpuInfo.Capacity = cpu.MilliValue()
	}

	return cpuInfo
}

// extractCPUModel tries to extract CPU model from various sources
func extractCPUModel(machineID string) string {
	// In a real scenario, this would parse /proc/cpuinfo data
	// For now, we'll provide a reasonable default
	models := map[string]string{
		"intel": "Intel Xeon",
		"amd":   "AMD EPYC",
		"arm":   "ARM Processor",
	}

	for key, model := range models {
		if strings.Contains(strings.ToLower(machineID), key) {
			return model
		}
	}
	return "Generic CPU"
}

// extractGPUInfo extracts GPU information from node labels
func extractGPUInfo(node *corev1.Node) GPUInfo {
	gpuInfo := GPUInfo{
		Available: false,
		GPUs:      []GPU{},
	}

	// Check for NVIDIA GPU labels
	if nvidiaGpus, ok := node.Labels["nvidia.com/gpu"]; ok {
		gpuInfo.Available = true
		gpuInfo.GPUs = append(gpuInfo.GPUs, GPU{
			Index: "nvidia-0",
			Model: nvidiaGpus,
		})
	}

	// Check for AMD GPU labels
	if amdGpus, ok := node.Labels["amd.com/gpu"]; ok {
		gpuInfo.Available = true
		gpuInfo.GPUs = append(gpuInfo.GPUs, GPU{
			Index: "amd-0",
			Model: amdGpus,
		})
	}

	// Check for other GPU labels
	for key, value := range node.Labels {
		if strings.Contains(key, "gpu") && key != "nvidia.com/gpu" && key != "amd.com/gpu" {
			gpuInfo.Available = true
			gpuInfo.GPUs = append(gpuInfo.GPUs, GPU{
				Index: key,
				Model: value,
			})
		}
	}

	return gpuInfo
}

// extractContainerRuntime extracts container runtime information
func extractContainerRuntime(node *corev1.Node) ContainerRuntime {
	runtime := ContainerRuntime{
		Name:    "Unknown",
		Version: "Unknown",
	}

	// Parse container runtime from node status
	runtimeVersion := node.Status.NodeInfo.ContainerRuntimeVersion
	if runtimeVersion != "" {
		parts := strings.Split(runtimeVersion, "://")
		if len(parts) >= 2 {
			runtime.Name = parts[0]
			// Extract version from format like "docker://20.10.12"
			versionParts := strings.Split(parts[1], "-")
			if len(versionParts) > 0 {
				runtime.Version = versionParts[0]
			}
		}
	}

	return runtime
}

// detectVirtualization detects the virtualization type
func detectVirtualization(node *corev1.Node) string {
	// Check provider ID to detect virtualization
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return "Bare Metal"
	}

	virtualizationTypes := map[string]string{
		"aws":       "AWS EC2",
		"gce":       "Google Compute Engine",
		"azure":     "Microsoft Azure",
		"vsphere":   "VMware vSphere",
		"kvm":       "KVM",
		"xen":       "Xen",
		"hyperv":    "Hyper-V",
		"openstack": "OpenStack",
	}

	providerLower := strings.ToLower(providerID)
	for key, virtType := range virtualizationTypes {
		if strings.Contains(providerLower, key) {
			return virtType
		}
	}

	// Check labels for virtualization hints
	for key := range node.Labels {
		if strings.Contains(key, "karpenter.sh") {
			return "Karpenter Provisioned"
		}
		if strings.Contains(key, "node.kubernetes.io/instance-type") {
			return "Cloud Instance"
		}
	}

	return "Virtual Machine"
}

// analyzeMemoryPressure analyzes memory pressure from node conditions
func analyzeMemoryPressure(node *corev1.Node) MemoryPressure {
	memPressure := MemoryPressure{
		Pressure: "UNKNOWN",
	}

	// Get total memory from node status
	memPressure.Total = getNodeMemoryTotal(node)

	// Check for memory pressure conditions
	memPressure.Pressure = getMemoryPressureStatus(node)

	// Calculate memory utilization ratio
	memPressure.UtilizationRatio = calculateMemoryUtilization(node, memPressure.Total)

	return memPressure
}

// getNodeMemoryTotal extracts total memory from node status
func getNodeMemoryTotal(node *corev1.Node) int64 {
	if memory, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
		return memory.Value()
	}
	return 0
}

// getMemoryPressureStatus determines memory pressure status from node conditions
func getMemoryPressureStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeMemoryPressure {
			if condition.Status == corev1.ConditionTrue {
				return "HIGH"
			}
			return "LOW"
		}
	}
	// Default to LOW if no explicit pressure condition found
	return "LOW"
}

// calculateMemoryUtilization calculates memory utilization ratio
func calculateMemoryUtilization(node *corev1.Node, totalMemory int64) float64 {
	if totalMemory <= 0 || node.Status.Allocatable == nil {
		return 0.0
	}

	allocatable, ok := node.Status.Allocatable[corev1.ResourceMemory]
	if !ok {
		return 0.0
	}

	allocated := totalMemory - allocatable.Value()
	if allocated <= 0 {
		return 0.0
	}

	return float64(allocated) / float64(totalMemory)
}

// analyzeFilesystemLatency analyzes filesystem information
func analyzeFilesystemLatency(node *corev1.Node) FilesystemLatency {
	fsLatency := FilesystemLatency{
		RootFSLatency:      0, // Would need kubelet metrics for actual latency
		RootFSInodesUsed:   0,
		RootFSCapacityUsed: 0,
	}

	// Check for disk pressure conditions
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeDiskPressure {
			if condition.Status == corev1.ConditionTrue {
				fsLatency.RootFSLatency = 100 // High latency indicator
			}
		}
	}

	return fsLatency
}

// GetNodeHealthStatus evaluates overall node health
func GetNodeHealthStatus(ctx context.Context, client kubernetes.Interface, nodeName string) (*NodeHealthStatus, error) {
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting node: %w", err)
	}

	status := &NodeHealthStatus{
		NodeName: node.Name,
		Status:   "HEALTHY",
		Issues:   []string{},
	}

	// Check node conditions
	for _, condition := range node.Status.Conditions {
		if condition.Status != corev1.ConditionTrue {
			continue
		}

		switch condition.Type {
		case corev1.NodeReady:
			if condition.Status != corev1.ConditionTrue {
				status.Status = "CRITICAL"
				status.Issues = append(status.Issues, "Node not ready")
			}
		case corev1.NodeMemoryPressure:
			status.Status = "WARNING"
			status.Issues = append(status.Issues, "Memory pressure detected")
		case corev1.NodeDiskPressure:
			status.Status = "WARNING"
			status.Issues = append(status.Issues, "Disk pressure detected")
		case corev1.NodePIDPressure:
			status.Status = "WARNING"
			status.Issues = append(status.Issues, "PID pressure detected")
		case corev1.NodeNetworkUnavailable:
			status.Status = "WARNING"
			status.Issues = append(status.Issues, "Network unavailable")
		}
	}

	return status, nil
}
