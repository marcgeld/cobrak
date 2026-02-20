package capacity

import (
	"context"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
