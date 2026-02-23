package capacity

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PressureLevel indicates resource pressure: LOW, MEDIUM, HIGH, SATURATED
type PressureLevel string

const (
	PressureLow       PressureLevel = "LOW"
	PressureMedium    PressureLevel = "MEDIUM"
	PressureHigh      PressureLevel = "HIGH"
	PressureSaturated PressureLevel = "SATURATED"
)

// NodePressure holds pressure information for a single node
type NodePressure struct {
	NodeName       string
	CPUPressure    PressureLevel
	CPUUtilization float64
	MemPressure    PressureLevel
	MemUtilization float64
}

// NamespacePressure holds pressure information for a namespace
type NamespacePressure struct {
	Namespace  string
	CPUPercent float64
	MemPercent float64
	CPUStatus  string
	MemStatus  string
}

// ClusterPressure holds overall cluster pressure
type ClusterPressure struct {
	Overall            PressureLevel
	CPUUtilization     float64
	MemUtilization     float64
	NodePressures      []NodePressure
	NamespacePressures []NamespacePressure
}

// PressureThresholds defines the pressure level thresholds
type PressureThresholds struct {
	Low       float64
	Medium    float64
	High      float64
	Saturated float64
}

// DefaultPressureThresholds returns the default pressure thresholds
func DefaultPressureThresholds() PressureThresholds {
	return PressureThresholds{
		Low:       50.0,
		Medium:    75.0,
		High:      90.0,
		Saturated: 100.0,
	}
}

// CalculatePressure analyzes cluster resources and returns pressure status using default thresholds
func CalculatePressure(ctx context.Context, client kubernetes.Interface, namespace string) (*ClusterPressure, error) {
	return CalculatePressureWithThresholds(ctx, client, namespace, DefaultPressureThresholds())
}

// CalculatePressureWithThresholds analyzes cluster resources with custom thresholds
func CalculatePressureWithThresholds(ctx context.Context, client kubernetes.Interface, namespace string, thresholds PressureThresholds) (*ClusterPressure, error) {
	pressure := &ClusterPressure{
		NodePressures:      []NodePressure{},
		NamespacePressures: []NamespacePressure{},
	}

	// Fetch cluster resources
	nodes, pods, err := fetchClusterResources(ctx, client, namespace)
	if err != nil {
		return nil, err
	}

	// Calculate per-node and per-namespace pressure
	calculateNodePressures(pressure, nodes, pods, thresholds)
	calculateNamespacePressures(pressure, nodes, pods, thresholds)
	calculateClusterPressure(pressure, nodes, pods)

	return pressure, nil
}

// fetchClusterResources retrieves nodes and pods from the cluster
func fetchClusterResources(ctx context.Context, client kubernetes.Interface, namespace string) ([]corev1.Node, []corev1.Pod, error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("listing nodes: %w", err)
	}

	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("listing pods: %w", err)
	}

	return nodes.Items, pods.Items, nil
}

// calculateNodePressures computes pressure for all nodes
func calculateNodePressures(pressure *ClusterPressure, nodes []corev1.Node, pods []corev1.Pod, thresholds PressureThresholds) {
	for i := range nodes {
		nodePressure := computeNodePressure(&nodes[i], pods, thresholds)
		pressure.NodePressures = append(pressure.NodePressures, nodePressure)
	}
}

// computeNodePressure calculates pressure for a single node with custom thresholds
func computeNodePressure(node *corev1.Node, pods []corev1.Pod, thresholds PressureThresholds) NodePressure {
	np := NodePressure{NodeName: node.Name}

	// Get node allocatable resources
	cpuAllocatable := node.Status.Allocatable.Cpu()
	memAllocatable := node.Status.Allocatable.Memory()

	// Sum resource requests for pods on this node
	var nodeCPURequest, nodeMemRequest int64
	for i := range pods {
		if pods[i].Spec.NodeName == node.Name {
			addPodResourcesForNode(&nodeCPURequest, &nodeMemRequest, &pods[i])
		}
	}

	// Calculate CPU pressure
	if cpuAllocatable != nil && cpuAllocatable.MilliValue() > 0 {
		np.CPUUtilization = (float64(nodeCPURequest) / float64(cpuAllocatable.MilliValue())) * 100
		np.CPUPressure = getPressureLevel(np.CPUUtilization, thresholds)
	}

	// Calculate Memory pressure
	if memAllocatable != nil && memAllocatable.Value() > 0 {
		np.MemUtilization = (float64(nodeMemRequest) / float64(memAllocatable.Value())) * 100
		np.MemPressure = getPressureLevel(np.MemUtilization, thresholds)
	}

	return np
}

// addPodResourcesForNode adds a pod's resource requests to node totals
func addPodResourcesForNode(cpuRequest, memRequest *int64, pod *corev1.Pod) {
	for i := range pod.Spec.Containers {
		c := &pod.Spec.Containers[i]
		if c.Resources.Requests != nil {
			if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				*cpuRequest += cpuReq.MilliValue()
			}
			if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
				*memRequest += memReq.Value()
			}
		}
	}
}

// calculateNamespacePressures computes pressure for all namespaces
func calculateNamespacePressures(pressure *ClusterPressure, nodes []corev1.Node, pods []corev1.Pod, thresholds PressureThresholds) {
	// Aggregate resources per namespace
	nsMap := aggregateNamespaceResources(pods)

	// Get total allocatable to calculate percentages
	totalAllocatable := getTotalAllocatable(nodes)

	// Convert to percentages and set status
	for ns := range nsMap {
		if totalAllocatable.CPU > 0 {
			nsMap[ns].CPUPercent = (nsMap[ns].CPUPercent / float64(totalAllocatable.CPU)) * 100
		}
		if totalAllocatable.Memory > 0 {
			nsMap[ns].MemPercent = (nsMap[ns].MemPercent / float64(totalAllocatable.Memory)) * 100
		}

		// Set status strings for high utilization
		if nsMap[ns].CPUPercent >= thresholds.High {
			nsMap[ns].CPUStatus = fmt.Sprintf("CPU %.0f%%", nsMap[ns].CPUPercent)
		}
		if nsMap[ns].MemPercent >= thresholds.High {
			nsMap[ns].MemStatus = fmt.Sprintf("Memory %.0f%%", nsMap[ns].MemPercent)
		}

		pressure.NamespacePressures = append(pressure.NamespacePressures, *nsMap[ns])
	}
}

// aggregateNamespaceResources sums resource requests by namespace
func aggregateNamespaceResources(pods []corev1.Pod) map[string]*NamespacePressure {
	nsMap := make(map[string]*NamespacePressure)

	for i := range pods {
		ns := pods[i].Namespace
		if _, exists := nsMap[ns]; !exists {
			nsMap[ns] = &NamespacePressure{Namespace: ns}
		}
		aggregatePodResourcesByNamespace(nsMap[ns], &pods[i])
	}

	return nsMap
}

// aggregatePodResourcesByNamespace adds pod resources to namespace totals
func aggregatePodResourcesByNamespace(nsPressure *NamespacePressure, pod *corev1.Pod) {
	for j := range pod.Spec.Containers {
		c := &pod.Spec.Containers[j]
		if c.Resources.Requests != nil {
			if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				nsPressure.CPUPercent += float64(cpuReq.MilliValue())
			}
			if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
				nsPressure.MemPercent += float64(memReq.Value())
			}
		}
	}
}

// AllocatableResources holds total cluster allocatable resources
type AllocatableResources struct {
	CPU    int64
	Memory int64
}

// getTotalAllocatable sums allocatable resources across all nodes
func getTotalAllocatable(nodes []corev1.Node) AllocatableResources {
	var total AllocatableResources

	for i := range nodes {
		if cpu := nodes[i].Status.Allocatable.Cpu(); cpu != nil {
			total.CPU += cpu.MilliValue()
		}
		if mem := nodes[i].Status.Allocatable.Memory(); mem != nil {
			total.Memory += mem.Value()
		}
	}

	return total
}

// calculateClusterPressure computes overall cluster pressure
func calculateClusterPressure(pressure *ClusterPressure, nodes []corev1.Node, pods []corev1.Pod) {
	// Find maximum pressure across all nodes
	maxCPUPressure, maxMemPressure := findMaxNodePressures(pressure.NodePressures)
	pressure.Overall = combinePressureLevels(maxCPUPressure, maxMemPressure)

	// Calculate cluster utilization percentages
	totalAllocatable := getTotalAllocatable(nodes)
	totalRequested := getTotalRequested(pods)

	if totalAllocatable.CPU > 0 {
		pressure.CPUUtilization = (float64(totalRequested.CPU) / float64(totalAllocatable.CPU)) * 100
	}
	if totalAllocatable.Memory > 0 {
		pressure.MemUtilization = (float64(totalRequested.Memory) / float64(totalAllocatable.Memory)) * 100
	}
}

// findMaxNodePressures finds the worst CPU and Memory pressure across all nodes
func findMaxNodePressures(nodePressures []NodePressure) (PressureLevel, PressureLevel) {
	var maxCPU, maxMem PressureLevel

	for _, np := range nodePressures {
		maxCPU = combinePressureLevels(np.CPUPressure, maxCPU)
		maxMem = combinePressureLevels(np.MemPressure, maxMem)
	}

	return maxCPU, maxMem
}

// getTotalRequested sums all resource requests across the cluster
func getTotalRequested(pods []corev1.Pod) AllocatableResources {
	var total AllocatableResources

	for i := range pods {
		for j := range pods[i].Spec.Containers {
			c := &pods[i].Spec.Containers[j]
			if c.Resources.Requests != nil {
				if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
					total.CPU += cpuReq.MilliValue()
				}
				if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
					total.Memory += memReq.Value()
				}
			}
		}
	}

	return total
}

// getPressureLevel determines pressure level based on utilization and thresholds
func getPressureLevel(utilization float64, thresholds PressureThresholds) PressureLevel {
	switch {
	case utilization >= thresholds.Saturated:
		return PressureSaturated
	case utilization >= thresholds.High:
		return PressureHigh
	case utilization >= thresholds.Medium:
		return PressureMedium
	case utilization >= thresholds.Low:
		return PressureLow
	default:
		return PressureLow
	}
}

// combinePressureLevels returns the worse of two pressure levels
func combinePressureLevels(a, b PressureLevel) PressureLevel {
	pressureOrder := map[PressureLevel]int{
		PressureLow:       0,
		PressureMedium:    1,
		PressureHigh:      2,
		PressureSaturated: 3,
	}

	if pressureOrder[a] >= pressureOrder[b] {
		return a
	}
	return b
}
