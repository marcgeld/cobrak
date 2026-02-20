package resources

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestBuildDiff_Empty(t *testing.T) {
	diffs := BuildDiff(nil, nil)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestBuildDiff_WithData(t *testing.T) {
	inventory := []ContainerResources{
		{
			Namespace:     "default",
			PodName:       "pod1",
			ContainerName: "c1",
			CPURequest:    resource.MustParse("200m"),
			HasCPURequest: true,
			MemRequest:    resource.MustParse("256Mi"),
			HasMemRequest: true,
		},
	}

	usage := []ContainerUsage{
		{
			Namespace:     "default",
			PodName:       "pod1",
			ContainerName: "c1",
			CPUUsage:      resource.MustParse("100m"),
			MemUsage:      resource.MustParse("128Mi"),
		},
	}

	diffs := BuildDiff(inventory, usage)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}

	d := diffs[0]
	if d.CPUUsageToRequest < 0.49 || d.CPUUsageToRequest > 0.51 {
		t.Errorf("expected CPUUsageToRequest ~0.5, got %f", d.CPUUsageToRequest)
	}
	if d.MemUsageToRequest < 0.49 || d.MemUsageToRequest > 0.51 {
		t.Errorf("expected MemUsageToRequest ~0.5, got %f", d.MemUsageToRequest)
	}
}
