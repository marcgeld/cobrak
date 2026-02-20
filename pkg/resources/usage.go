package resources

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricsReader is the interface for fetching pod metrics.
type MetricsReader interface {
	PodMetrics(ctx context.Context, namespace string) ([]ContainerUsage, error)
	IsAvailable(ctx context.Context) (bool, error)
}

// metricsReaderImpl is the production implementation of MetricsReader.
type metricsReaderImpl struct {
	client metricsclient.Interface
}

// NewMetricsReaderFromConfig creates a MetricsReader from a rest.Config.
func NewMetricsReaderFromConfig(cfg *rest.Config) (MetricsReader, error) {
	mc, err := metricsclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating metrics client: %w", err)
	}
	return &metricsReaderImpl{client: mc}, nil
}

// IsAvailable checks if the metrics.k8s.io API is available.
func (m *metricsReaderImpl) IsAvailable(ctx context.Context) (bool, error) {
	_, err := m.client.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// PodMetrics fetches actual CPU/memory usage for pods.
func (m *metricsReaderImpl) PodMetrics(ctx context.Context, namespace string) ([]ContainerUsage, error) {
	podMetrics, err := m.client.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pod metrics: %w", err)
	}

	return extractContainerUsages(podMetrics.Items), nil
}

func extractContainerUsages(items []metricsv1beta1.PodMetrics) []ContainerUsage {
	var usages []ContainerUsage
	for i := range items {
		pm := &items[i]
		for _, c := range pm.Containers {
			cu := ContainerUsage{
				Namespace:     pm.Namespace,
				PodName:       pm.Name,
				ContainerName: c.Name,
			}
			if cpuQ, ok := c.Usage[v1.ResourceCPU]; ok {
				cu.CPUUsage = cpuQ.DeepCopy()
			}
			if memQ, ok := c.Usage[v1.ResourceMemory]; ok {
				cu.MemUsage = memQ.DeepCopy()
			}
			usages = append(usages, cu)
		}
	}

	sort.Slice(usages, func(i, j int) bool {
		a, b := usages[i], usages[j]
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.PodName != b.PodName {
			return a.PodName < b.PodName
		}
		return a.ContainerName < b.ContainerName
	})

	return usages
}
