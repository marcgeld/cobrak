package capacity

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

// CalculatePressure analyzes cluster resources and returns pressure status
func CalculatePressure(ctx context.Context, client kubernetes.Interface, namespace string) (*ClusterPressure, error) {
	// Use default thresholds
	thresholds := PressureThresholds{
		Low:       50.0,
		Medium:    75.0,
		High:      90.0,
		Saturated: 100.0,
	}
	return CalculatePressureWithThresholds(ctx, client, namespace, thresholds)
}

// PressureThresholds defines the pressure level thresholds
type PressureThresholds struct {
	Low       float64
	Medium    float64
	High      float64
	Saturated float64
}

// CalculatePressureWithThresholds analyzes cluster resources with custom thresholds
func CalculatePressureWithThresholds(ctx context.Context, client kubernetes.Interface, namespace string, thresholds PressureThresholds) (*ClusterPressure, error) {
	pressure := &ClusterPressure{
		NodePressures:      []NodePressure{},
		NamespacePressures: []NamespacePressure{},
	}

	// Get nodes and their capacity/allocatable
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	// Get all pods (or specific namespace)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	// Calculate per-node pressure
	for _, node := range nodes.Items {
		nodePressure := calculateNodePressureWithThresholds(&node, pods, thresholds)
		pressure.NodePressures = append(pressure.NodePressures, nodePressure)
	}

	// Calculate per-namespace pressure
	nsMap := make(map[string]*NamespacePressure)
	for _, pod := range pods.Items {
		ns := pod.Namespace
		if _, exists := nsMap[ns]; !exists {
			nsMap[ns] = &NamespacePressure{
				Namespace: ns,
			}
		}

		// Sum requests per namespace
		for _, c := range pod.Spec.Containers {
			if c.Resources.Requests != nil {
				if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
					nsMap[ns].CPUPercent += float64(cpuReq.MilliValue())
				}
				if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
					nsMap[ns].MemPercent += float64(memReq.Value())
				}
			}
		}
	}

	// Get total allocatable to calculate percentages
	var totalCPUAllocatable, totalMemAllocatable int64
	for _, node := range nodes.Items {
		if cpu := node.Status.Allocatable.Cpu(); cpu != nil {
			totalCPUAllocatable += cpu.MilliValue()
		}
		if mem := node.Status.Allocatable.Memory(); mem != nil {
			totalMemAllocatable += mem.Value()
		}
	}

	// Convert to percentages and set status
	for _, nsPressure := range nsMap {
		if totalCPUAllocatable > 0 {
			nsPressure.CPUPercent = (nsPressure.CPUPercent / float64(totalCPUAllocatable)) * 100
		}
		if totalMemAllocatable > 0 {
			nsPressure.MemPercent = (nsPressure.MemPercent / float64(totalMemAllocatable)) * 100
		}

		// Set status strings
		if nsPressure.CPUPercent >= thresholds.High {
			nsPressure.CPUStatus = fmt.Sprintf("CPU %.0f%%", nsPressure.CPUPercent)
		}
		if nsPressure.MemPercent >= thresholds.High {
			nsPressure.MemStatus = fmt.Sprintf("Memory %.0f%%", nsPressure.MemPercent)
		}

		pressure.NamespacePressures = append(pressure.NamespacePressures, *nsPressure)
	}

	// Calculate overall cluster pressure
	var maxCPUPressure, maxMemPressure PressureLevel
	for _, np := range pressure.NodePressures {
		if prioritizePressure(np.CPUPressure, maxCPUPressure) == np.CPUPressure {
			maxCPUPressure = np.CPUPressure
		}
		if prioritizePressure(np.MemPressure, maxMemPressure) == np.MemPressure {
			maxMemPressure = np.MemPressure
		}
	}

	// Overall pressure is the worse of CPU/Memory
	pressure.Overall = prioritizePressure(maxCPUPressure, maxMemPressure)
	if pressure.Overall == "" {
		pressure.Overall = PressureLow
	}

	// Calculate utilization percentages
	var totalCPURequest, totalMemRequest int64
	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			if c.Resources.Requests != nil {
				if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
					totalCPURequest += cpuReq.MilliValue()
				}
				if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
					totalMemRequest += memReq.Value()
				}
			}
		}
	}

	if totalCPUAllocatable > 0 {
		pressure.CPUUtilization = (float64(totalCPURequest) / float64(totalCPUAllocatable)) * 100
	}
	if totalMemAllocatable > 0 {
		pressure.MemUtilization = (float64(totalMemRequest) / float64(totalMemAllocatable)) * 100
	}

	return pressure, nil
}

// calculateNodePressure calculates pressure for a single node based on requested resources
func calculateNodePressure(node *corev1.Node, pods *corev1.PodList) NodePressure {
	thresholds := PressureThresholds{
		Low:       50.0,
		Medium:    75.0,
		High:      90.0,
		Saturated: 100.0,
	}
	return calculateNodePressureWithThresholds(node, pods, thresholds)
}

// calculateNodePressureWithThresholds calculates pressure for a single node with custom thresholds
func calculateNodePressureWithThresholds(node *corev1.Node, pods *corev1.PodList, thresholds PressureThresholds) NodePressure {
	np := NodePressure{
		NodeName: node.Name,
	}

	// Get node allocatable
	cpuAllocatable := node.Status.Allocatable.Cpu()
	memAllocatable := node.Status.Allocatable.Memory()

	var nodeCPURequest, nodeMemRequest *resource.Quantity
	nodeCPURequest = resource.NewMilliQuantity(0, resource.DecimalSI)
	nodeMemRequest = resource.NewQuantity(0, resource.BinarySI)

	// Sum requests for pods on this node
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == node.Name {
			for _, c := range pod.Spec.Containers {
				if c.Resources.Requests != nil {
					if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
						nodeCPURequest.Add(cpuReq)
					}
					if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
						nodeMemRequest.Add(memReq)
					}
				}
			}
		}
	}

	// Calculate CPU utilization and pressure
	if cpuAllocatable != nil && cpuAllocatable.MilliValue() > 0 {
		np.CPUUtilization = (float64(nodeCPURequest.MilliValue()) / float64(cpuAllocatable.MilliValue())) * 100
		np.CPUPressure = calculatePressureLevelWithThresholds(np.CPUUtilization, thresholds)
	}

	// Calculate Memory utilization and pressure
	if memAllocatable != nil && memAllocatable.Value() > 0 {
		np.MemUtilization = (float64(nodeMemRequest.Value()) / float64(memAllocatable.Value())) * 100
		np.MemPressure = calculatePressureLevelWithThresholds(np.MemUtilization, thresholds)
	}

	return np
}

// calculatePressureLevel determines pressure level based on utilization percentage using default thresholds
func calculatePressureLevel(utilization float64) PressureLevel {
	thresholds := PressureThresholds{
		Low:       50.0,
		Medium:    75.0,
		High:      90.0,
		Saturated: 100.0,
	}
	return calculatePressureLevelWithThresholds(utilization, thresholds)
}

// calculatePressureLevelWithThresholds determines pressure level based on utilization and custom thresholds
func calculatePressureLevelWithThresholds(utilization float64, thresholds PressureThresholds) PressureLevel {
	if utilization >= thresholds.Saturated {
		return PressureSaturated
	}
	if utilization >= thresholds.High {
		return PressureHigh
	}
	if utilization >= thresholds.Medium {
		return PressureMedium
	}
	if utilization >= thresholds.Low {
		return PressureLow
	}
	return PressureLow
}

// prioritizePressure returns the worse of two pressure levels
func prioritizePressure(a, b PressureLevel) PressureLevel {
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
