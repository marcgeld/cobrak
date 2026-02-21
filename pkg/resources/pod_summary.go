package resources

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// BuildPodSummaries aggregates CPU/memory requests and limits per pod.
func BuildPodSummaries(ctx context.Context, client kubernetes.Interface, namespace string) ([]PodResourceSummary, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	var summaries []PodResourceSummary
	podMap := make(map[string]*PodResourceSummary)

	for i := range pods.Items {
		pod := &pods.Items[i]
		key := pod.Namespace + "/" + pod.Name

		if _, exists := podMap[key]; !exists {
			podMap[key] = &PodResourceSummary{
				Namespace:  pod.Namespace,
				PodName:    pod.Name,
				CPUUsage:   *resource.NewQuantity(0, resource.DecimalSI),
				CPURequest: *resource.NewQuantity(0, resource.DecimalSI),
				CPULimit:   *resource.NewQuantity(0, resource.DecimalSI),
				MemUsage:   *resource.NewQuantity(0, resource.BinarySI),
				MemRequest: *resource.NewQuantity(0, resource.BinarySI),
				MemLimit:   *resource.NewQuantity(0, resource.BinarySI),
			}
		}

		summary := podMap[key]

		// Sum all container requests/limits
		for _, c := range pod.Spec.Containers {
			if c.Resources.Requests != nil {
				if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
					summary.CPURequest.Add(cpuReq)
				}
				if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
					summary.MemRequest.Add(memReq)
				}
			}
			if c.Resources.Limits != nil {
				if cpuLim, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
					summary.CPULimit.Add(cpuLim)
				}
				if memLim, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
					summary.MemLimit.Add(memLim)
				}
			}
		}

		// Sum all init container requests/limits
		for _, c := range pod.Spec.InitContainers {
			if c.Resources.Requests != nil {
				if cpuReq, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
					summary.CPURequest.Add(cpuReq)
				}
				if memReq, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
					summary.MemRequest.Add(memReq)
				}
			}
			if c.Resources.Limits != nil {
				if cpuLim, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
					summary.CPULimit.Add(cpuLim)
				}
				if memLim, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
					summary.MemLimit.Add(memLim)
				}
			}
		}
	}

	// Convert map to slice and sort by namespace/pod name
	for _, summary := range podMap {
		summaries = append(summaries, *summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Namespace == summaries[j].Namespace {
			return summaries[i].PodName < summaries[j].PodName
		}
		return summaries[i].Namespace < summaries[j].Namespace
	})

	return summaries, nil
}

// BuildPodSummariesWithUsage aggregates CPU/memory including actual usage from metrics.
func BuildPodSummariesWithUsage(ctx context.Context, client kubernetes.Interface, metricsReader MetricsReader, namespace string) ([]PodResourceSummary, error) {
	// Get base summaries (requests/limits)
	summaries, err := BuildPodSummaries(ctx, client, namespace)
	if err != nil {
		return nil, err
	}

	// Try to get usage metrics
	usages, err := metricsReader.PodMetrics(ctx, namespace)
	if err != nil {
		// Metrics not available, just return request/limit summaries
		return summaries, nil
	}

	// Create usage map for quick lookup
	usageMap := make(map[string]*ContainerUsage)
	for _, u := range usages {
		key := u.Namespace + "/" + u.PodName + "/" + u.ContainerName
		usageMap[key] = &u
	}

	// Update summaries with usage data
	for i := range summaries {
		pods, err := client.CoreV1().Pods(summaries[i].Namespace).List(ctx, metav1.ListOptions{
			FieldSelector: "metadata.name=" + summaries[i].PodName,
		})
		if err == nil && len(pods.Items) > 0 {
			pod := &pods.Items[0]
			for _, c := range pod.Spec.Containers {
				key := summaries[i].Namespace + "/" + summaries[i].PodName + "/" + c.Name
				if usage, ok := usageMap[key]; ok {
					summaries[i].CPUUsage.Add(usage.CPUUsage)
					summaries[i].MemUsage.Add(usage.MemUsage)
				}
			}
		}
	}

	return summaries, nil
}
