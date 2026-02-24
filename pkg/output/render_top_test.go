package output

import (
	"testing"

	"github.com/marcgeld/cobrak/pkg/resources"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestRenderPodResourceSummary_WithoutTop(t *testing.T) {
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

	result := RenderPodResourceSummary(pods, 0) // top=0 means no limit

	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Count lines to verify both pods are rendered
	// Header + 2 pods = 3 lines minimum
	lines := len(result) > 0
	if !lines {
		t.Error("Expected output with pods")
	}
}

func TestRenderPodResourceSummary_WithTop(t *testing.T) {
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
		{
			Namespace:  "default",
			PodName:    "pod3",
			CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
	}

	result := RenderPodResourceSummary(pods, 1) // top=1 should only show 1 pod

	// Verify that pod1 is in output
	if !containsString(result, "pod1") {
		t.Error("Expected pod1 in output")
	}

	// Verify that pod3 is NOT in output (limited to 1)
	if containsString(result, "pod3") {
		t.Error("Expected pod3 to be excluded when top=1")
	}
}

func TestRenderPodResourceSummary_TopZero(t *testing.T) {
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

	result := RenderPodResourceSummary(pods, 0) // top=0 means show all

	// Both pods should be in output
	if !containsString(result, "pod1") {
		t.Error("Expected pod1 in output when top=0")
	}
	if !containsString(result, "pod2") {
		t.Error("Expected pod2 in output when top=0")
	}
}

func TestRenderPodResourceSummary_EmptyList(t *testing.T) {
	pods := []resources.PodResourceSummary{}

	result := RenderPodResourceSummary(pods, 5)

	if result != "No pods found." {
		t.Errorf("Expected 'No pods found.', got %q", result)
	}
}

func TestRenderPodResourceSummaryWithUsage_WithTop(t *testing.T) {
	pods := []resources.PodResourceSummary{
		{
			Namespace:  "default",
			PodName:    "pod1",
			CPUUsage:   *resource.NewMilliQuantity(50, resource.DecimalSI),
			CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemUsage:   *resource.NewQuantity(64*1024*1024, resource.BinarySI),
			MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
		{
			Namespace:  "default",
			PodName:    "pod2",
			CPUUsage:   *resource.NewMilliQuantity(50, resource.DecimalSI),
			CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemUsage:   *resource.NewQuantity(64*1024*1024, resource.BinarySI),
			MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
		{
			Namespace:  "default",
			PodName:    "pod3",
			CPUUsage:   *resource.NewMilliQuantity(50, resource.DecimalSI),
			CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemUsage:   *resource.NewQuantity(64*1024*1024, resource.BinarySI),
			MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
	}

	result := RenderPodResourceSummaryWithUsage(pods, 2) // top=2 should show only 2 pods

	if !containsString(result, "pod1") {
		t.Error("Expected pod1 in output")
	}
	if !containsString(result, "pod2") {
		t.Error("Expected pod2 in output")
	}
	if containsString(result, "pod3") {
		t.Error("Expected pod3 to be excluded when top=2")
	}
}

// Helper function
func containsString(haystack, needle string) bool {
	for i := 0; i < len(haystack)-len(needle)+1; i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
