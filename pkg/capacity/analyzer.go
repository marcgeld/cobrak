package capacity

import (
	"context"
	"fmt"
	"sort"

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
	// Get node capacities
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	summary := &ClusterCapacitySummary{
		TotalCPUCapacity:    *resource.NewQuantity(0, resource.DecimalSI),
		TotalCPUAllocatable: *resource.NewQuantity(0, resource.DecimalSI),
		TotalMemCapacity:    *resource.NewQuantity(0, resource.BinarySI),
		TotalMemAllocatable: *resource.NewQuantity(0, resource.BinarySI),
		TotalCPURequests:    *resource.NewQuantity(0, resource.DecimalSI),
		TotalCPULimits:      *resource.NewQuantity(0, resource.DecimalSI),
		TotalMemRequests:    *resource.NewQuantity(0, resource.BinarySI),
		TotalMemLimits:      *resource.NewQuantity(0, resource.BinarySI),
	}

	// Sum node capacities
	for _, node := range nodes.Items {
		summary.TotalCPUCapacity.Add(*node.Status.Capacity.Cpu())
		summary.TotalCPUAllocatable.Add(*node.Status.Allocatable.Cpu())
		summary.TotalMemCapacity.Add(*node.Status.Capacity.Memory())
		summary.TotalMemAllocatable.Add(*node.Status.Allocatable.Memory())
	}

	// Get pod requests/limits
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
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
		for _, c := range pod.Spec.InitContainers {
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

	return summary, nil
}
