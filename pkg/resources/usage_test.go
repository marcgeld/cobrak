package resources

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

type fakeMetricsReader struct {
	available bool
	usages    []ContainerUsage
	err       error
}

func (f *fakeMetricsReader) IsAvailable(_ context.Context) (bool, error) {
	return f.available, f.err
}

func (f *fakeMetricsReader) PodMetrics(_ context.Context, _ string) ([]ContainerUsage, error) {
	return f.usages, f.err
}

func TestFakeMetricsReader_NotAvailable(t *testing.T) {
	reader := &fakeMetricsReader{available: false}
	available, err := reader.IsAvailable(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected not available")
	}
}

func TestFakeMetricsReader_PodMetrics(t *testing.T) {
	reader := &fakeMetricsReader{
		available: true,
		usages: []ContainerUsage{
			{Namespace: "ns1", PodName: "pod1", ContainerName: "c1", CPUUsage: resource.MustParse("100m"), MemUsage: resource.MustParse("128Mi")},
			{Namespace: "ns1", PodName: "pod2", ContainerName: "c2", CPUUsage: resource.MustParse("200m"), MemUsage: resource.MustParse("256Mi")},
		},
	}
	usages, err := reader.PodMetrics(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(usages) != 2 {
		t.Errorf("expected 2 usages, got %d", len(usages))
	}
}
