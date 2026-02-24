package output

import (
	"strings"
	"testing"

	"github.com/marcgeld/cobrak/pkg/capacity"
	"github.com/marcgeld/cobrak/pkg/resources"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestRenderPressureSimple(t *testing.T) {
	pressure := &capacity.ClusterPressure{
		Overall:        capacity.PressureLow,
		CPUUtilization: 45.0,
		MemUtilization: 60.0,
		NodePressures: []capacity.NodePressure{
			{
				NodeName:       "node-1",
				CPUPressure:    capacity.PressureMedium,
				CPUUtilization: 70.0,
				MemPressure:    capacity.PressureLow,
				MemUtilization: 40.0,
			},
		},
		NamespacePressures: []capacity.NamespacePressure{
			{
				Namespace:  "production",
				CPUPercent: 85.0,
				MemPercent: 90.0,
			},
		},
	}

	result := RenderPressureSimple(pressure)

	if result == "" {
		t.Error("Expected non-empty output")
	}

	// Should contain cluster pressure
	if !strings.Contains(result, "Cluster Pressure") {
		t.Error("Expected 'Cluster Pressure' in output")
	}

	// Should contain node pressure for high utilization
	if !strings.Contains(result, "node-1") {
		t.Error("Expected node-1 in output")
	}

	// Should contain namespace pressure for high utilization
	if !strings.Contains(result, "production") {
		t.Error("Expected production namespace in output")
	}
}

func TestRenderPressureSimple_AllLow(t *testing.T) {
	pressure := &capacity.ClusterPressure{
		Overall: capacity.PressureLow,
		NodePressures: []capacity.NodePressure{
			{
				NodeName:    "node-1",
				CPUPressure: capacity.PressureLow,
				MemPressure: capacity.PressureLow,
			},
		},
		NamespacePressures: []capacity.NamespacePressure{},
	}

	result := RenderPressureSimple(pressure)

	if result == "" {
		t.Error("Expected non-empty output")
	}

	// Should not show nodes with LOW pressure
	if strings.Contains(result, "node-1") {
		t.Error("Expected node-1 to be excluded (LOW pressure)")
	}
}

func TestRenderPodResourceSummaryTotals(t *testing.T) {
	pods := []resources.PodResourceSummary{
		{
			Namespace:  "default",
			PodName:    "pod1",
			CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
		{
			Namespace:  "default",
			PodName:    "pod2",
			CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
	}

	result := RenderPodResourceSummaryTotals(pods)

	if result == "" {
		t.Error("Expected non-empty totals output")
	}

	// Should contain "TOTAL" or similar indicator
	if !strings.Contains(strings.ToUpper(result), "TOTAL") {
		t.Error("Expected totals label in output")
	}
}

func TestRenderPodResourceSummaryTotals_Empty(t *testing.T) {
	var pods []resources.PodResourceSummary

	result := RenderPodResourceSummaryTotals(pods)

	// Empty list returns empty string (which is ok)
	if result != "" {
		t.Errorf("Expected empty string for empty pods, got %q", result)
	}
}

func TestColorizePressureLevel(t *testing.T) {
	tests := []struct {
		level capacity.PressureLevel
		name  string
	}{
		{capacity.PressureLow, "LOW"},
		{capacity.PressureMedium, "MEDIUM"},
		{capacity.PressureHigh, "HIGH"},
		{capacity.PressureSaturated, "SATURATED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			// This is a private function, but we test it through RenderPressureSimple
			pressure := &capacity.ClusterPressure{
				Overall:            tt.level,
				NodePressures:      []capacity.NodePressure{},
				NamespacePressures: []capacity.NamespacePressure{},
			}

			result := RenderPressureSimple(pressure)

			if result == "" {
				t.Error("Expected non-empty output")
			}

			if !strings.Contains(result, "Cluster Pressure") {
				t.Error("Expected cluster pressure in output")
			}
		})
	}
}
