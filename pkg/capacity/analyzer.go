package capacity

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NodeCapacity holds allocatable and total capacity data for a single node.
type NodeCapacity struct {
	Name           string
	CPUAllocatable resource.Quantity
	CPUCapacity    resource.Quantity
	MemAllocatable resource.Quantity
	MemCapacity    resource.Quantity
}

// ClusterCapacitySummary holds aggregated capacity and request data for the entire cluster.
type ClusterCapacitySummary struct {
	// Capacity and allocatable from nodes
	TotalCPUCapacity    resource.Quantity
	TotalCPUAllocatable resource.Quantity
	TotalMemCapacity    resource.Quantity
	TotalMemAllocatable resource.Quantity

	// Requested/Limited resources from pods
	TotalCPURequests resource.Quantity
	TotalCPULimits   resource.Quantity
	TotalMemRequests resource.Quantity
	TotalMemLimits   resource.Quantity
}

// Analyze lists all nodes and returns their capacity data sorted by node name.
func Analyze(ctx context.Context, client kubernetes.Interface) ([]NodeCapacity, error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	result := make([]NodeCapacity, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		nc := NodeCapacity{
			Name:           node.Name,
			CPUAllocatable: node.Status.Allocatable.Cpu().DeepCopy(),
			CPUCapacity:    node.Status.Capacity.Cpu().DeepCopy(),
			MemAllocatable: node.Status.Allocatable.Memory().DeepCopy(),
			MemCapacity:    node.Status.Capacity.Memory().DeepCopy(),
		}
		result = append(result, nc)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// AnalyzeSummary aggregates all node capacity and pod requests/limits into a cluster summary.
func AnalyzeSummary(ctx context.Context, client kubernetes.Interface, namespace string) (*ClusterCapacitySummary, error) {
	summary := newEmptySummary()

	// Get and sum node capacities
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	sumNodeCapacities(summary, nodes.Items)

	// Get and sum pod requests/limits
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}
	sumPodResources(summary, pods.Items)

	return summary, nil
}

// newEmptySummary creates a ClusterCapacitySummary with all quantities initialized to zero.
func newEmptySummary() *ClusterCapacitySummary {
	return &ClusterCapacitySummary{
		TotalCPUCapacity:    *resource.NewQuantity(0, resource.DecimalSI),
		TotalCPUAllocatable: *resource.NewQuantity(0, resource.DecimalSI),
		TotalMemCapacity:    *resource.NewQuantity(0, resource.BinarySI),
		TotalMemAllocatable: *resource.NewQuantity(0, resource.BinarySI),
		TotalCPURequests:    *resource.NewQuantity(0, resource.DecimalSI),
		TotalCPULimits:      *resource.NewQuantity(0, resource.DecimalSI),
		TotalMemRequests:    *resource.NewQuantity(0, resource.BinarySI),
		TotalMemLimits:      *resource.NewQuantity(0, resource.BinarySI),
	}
}

// sumNodeCapacities aggregates capacity from all nodes into the summary.
func sumNodeCapacities(summary *ClusterCapacitySummary, nodes []corev1.Node) {
	for _, node := range nodes {
		summary.TotalCPUCapacity.Add(*node.Status.Capacity.Cpu())
		summary.TotalCPUAllocatable.Add(*node.Status.Allocatable.Cpu())
		summary.TotalMemCapacity.Add(*node.Status.Capacity.Memory())
		summary.TotalMemAllocatable.Add(*node.Status.Allocatable.Memory())
	}
}

// sumPodResources aggregates requests and limits from all containers in all pods.
func sumPodResources(summary *ClusterCapacitySummary, pods []corev1.Pod) {
	for _, pod := range pods {
		sumContainerResources(summary, pod.Spec.Containers)
		sumContainerResources(summary, pod.Spec.InitContainers)
	}
}

// sumContainerResources aggregates requests and limits from a slice of containers.
func sumContainerResources(summary *ClusterCapacitySummary, containers []corev1.Container) {
	for _, c := range containers {
		if c.Resources.Requests != nil {
			if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				summary.TotalCPURequests.Add(cpuReq)
			}
			if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
				summary.TotalMemRequests.Add(memReq)
			}
		}
		if c.Resources.Limits != nil {
			if cpuLim, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
				summary.TotalCPULimits.Add(cpuLim)
			}
			if memLim, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
				summary.TotalMemLimits.Add(memLim)
			}
		}
	}
}
