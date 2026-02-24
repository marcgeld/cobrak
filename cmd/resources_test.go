package cmd

import (
	"bytes"
	"testing"

	"github.com/marcgeld/cobrak/pkg/capacity"
	"github.com/marcgeld/cobrak/pkg/output"
	"github.com/marcgeld/cobrak/pkg/resources"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestBuildResourcesSummary_TopLimit(t *testing.T) {
	// Create mock data with 5 pods
	podSummaries := []resources.PodResourceSummary{
		createMockPod("pod1"),
		createMockPod("pod2"),
		createMockPod("pod3"),
		createMockPod("pod4"),
		createMockPod("pod5"),
	}

	nsInventories := []resources.NamespaceInventory{
		{
			Namespace:       "default",
			ContainersTotal: 5,
		},
	}

	summary := &capacity.ClusterCapacitySummary{
		TotalCPUCapacity:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
		TotalCPUAllocatable: *resource.NewMilliQuantity(900, resource.DecimalSI),
		TotalCPURequests:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		TotalCPULimits:      *resource.NewMilliQuantity(800, resource.DecimalSI),
		TotalMemCapacity:    *resource.NewQuantity(4*1024*1024*1024, resource.BinarySI),
		TotalMemAllocatable: *resource.NewQuantity(3*1024*1024*1024, resource.BinarySI),
		TotalMemRequests:    *resource.NewQuantity(1*1024*1024*1024, resource.BinarySI),
		TotalMemLimits:      *resource.NewQuantity(2*1024*1024*1024, resource.BinarySI),
	}

	// Test with top=2
	top := 2
	result := buildResourcesSummary(summary, podSummaries, nsInventories, false, top)

	if len(result.PodDetails) != 2 {
		t.Errorf("Expected 2 pods in result with top=2, got %d", len(result.PodDetails))
	}

	if result.PodDetails[0].Pod != "pod1" {
		t.Errorf("Expected first pod to be pod1, got %s", result.PodDetails[0].Pod)
	}

	if result.PodDetails[1].Pod != "pod2" {
		t.Errorf("Expected second pod to be pod2, got %s", result.PodDetails[1].Pod)
	}
}

func TestBuildResourcesSummary_TopZero(t *testing.T) {
	// Create mock data with 3 pods
	podSummaries := []resources.PodResourceSummary{
		createMockPod("pod1"),
		createMockPod("pod2"),
		createMockPod("pod3"),
	}

	nsInventories := []resources.NamespaceInventory{
		{
			Namespace:       "default",
			ContainersTotal: 3,
		},
	}

	summary := &capacity.ClusterCapacitySummary{
		TotalCPUCapacity:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
		TotalCPUAllocatable: *resource.NewMilliQuantity(900, resource.DecimalSI),
		TotalCPURequests:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		TotalCPULimits:      *resource.NewMilliQuantity(800, resource.DecimalSI),
		TotalMemCapacity:    *resource.NewQuantity(4*1024*1024*1024, resource.BinarySI),
		TotalMemAllocatable: *resource.NewQuantity(3*1024*1024*1024, resource.BinarySI),
		TotalMemRequests:    *resource.NewQuantity(1*1024*1024*1024, resource.BinarySI),
		TotalMemLimits:      *resource.NewQuantity(2*1024*1024*1024, resource.BinarySI),
	}

	// Test with top=0 (no limit)
	top := 0
	result := buildResourcesSummary(summary, podSummaries, nsInventories, false, top)

	if len(result.PodDetails) != 3 {
		t.Errorf("Expected 3 pods in result with top=0, got %d", len(result.PodDetails))
	}
}

func TestBuildResourcesSummary_TopLargerThanPods(t *testing.T) {
	// Create mock data with 2 pods
	podSummaries := []resources.PodResourceSummary{
		createMockPod("pod1"),
		createMockPod("pod2"),
	}

	nsInventories := []resources.NamespaceInventory{
		{
			Namespace:       "default",
			ContainersTotal: 2,
		},
	}

	summary := &capacity.ClusterCapacitySummary{
		TotalCPUCapacity:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
		TotalCPUAllocatable: *resource.NewMilliQuantity(900, resource.DecimalSI),
		TotalCPURequests:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		TotalCPULimits:      *resource.NewMilliQuantity(800, resource.DecimalSI),
		TotalMemCapacity:    *resource.NewQuantity(4*1024*1024*1024, resource.BinarySI),
		TotalMemAllocatable: *resource.NewQuantity(3*1024*1024*1024, resource.BinarySI),
		TotalMemRequests:    *resource.NewQuantity(1*1024*1024*1024, resource.BinarySI),
		TotalMemLimits:      *resource.NewQuantity(2*1024*1024*1024, resource.BinarySI),
	}

	// Test with top=10 (larger than number of pods)
	top := 10
	result := buildResourcesSummary(summary, podSummaries, nsInventories, false, top)

	if len(result.PodDetails) != 2 {
		t.Errorf("Expected 2 pods in result when top=10, got %d", len(result.PodDetails))
	}
}

// Helper function to create mock pod
func createMockPod(name string) resources.PodResourceSummary {
	return resources.PodResourceSummary{
		Namespace:  "default",
		PodName:    name,
		CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
		CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
		MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
		MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
	}
}

// Test that output formatting respects top limit
func TestRenderOutput_TopLimit(t *testing.T) {
	// Create test data
	pods := []resources.PodResourceSummary{
		createMockPod("pod1"),
		createMockPod("pod2"),
		createMockPod("pod3"),
	}

	result := output.RenderPodResourceSummary(pods, 2)

	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Verify pod1 and pod2 are in output
	if !containsSubstring(result, "pod1") {
		t.Error("Expected pod1 in output")
	}
	if !containsSubstring(result, "pod2") {
		t.Error("Expected pod2 in output")
	}

	// Verify pod3 is NOT in output
	if containsSubstring(result, "pod3") {
		t.Error("Expected pod3 to be excluded")
	}
}

// Helper function
func containsSubstring(text, substring string) bool {
	return bytes.Contains([]byte(text), []byte(substring))
}
