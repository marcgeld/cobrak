package output

import (
	"strings"
	"testing"

	"github.com/marcgeld/cobrak/pkg/resources"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TestRenderTableFunctions tests all table rendering functions
func TestRenderTableFunctions(t *testing.T) {
	// Test namespace inventory table
	inventories := []resources.NamespaceInventory{
		{
			Namespace:                    "default",
			ContainersTotal:              10,
			ContainersMissingAnyRequests: 2,
			ContainersMissingAnyLimits:   1,
			CPURequestsTotal:             resource.MustParse("2"),
			CPULimitsTotal:               resource.MustParse("4"),
			MemRequestsTotal:             resource.MustParse("2Gi"),
			MemLimitsTotal:               resource.MustParse("4Gi"),
		},
		{
			Namespace:                    "kube-system",
			ContainersTotal:              5,
			ContainersMissingAnyRequests: 0,
			ContainersMissingAnyLimits:   0,
			CPURequestsTotal:             resource.MustParse("1"),
			CPULimitsTotal:               resource.MustParse("2"),
			MemRequestsTotal:             resource.MustParse("1Gi"),
			MemLimitsTotal:               resource.MustParse("2Gi"),
		},
	}

	result := RenderNamespaceInventoryTable(inventories)

	if result == "" {
		t.Error("expected non-empty namespace inventory table")
	}

	if !strings.Contains(result, "default") {
		t.Error("expected default namespace in output")
	}

	if !strings.Contains(result, "kube-system") {
		t.Error("expected kube-system namespace in output")
	}

	if !strings.Contains(result, "10") {
		t.Error("expected container count in output")
	}
}

// TestRenderMissingResourcesTable tests missing resources rendering
func TestRenderMissingResourcesTable(t *testing.T) {
	containers := []resources.ContainerResources{
		{
			Namespace:     "default",
			PodName:       "pod-no-cpu",
			ContainerName: "app",
			HasCPURequest: false,
			HasMemRequest: true,
			HasCPULimit:   false,
			HasMemLimit:   true,
		},
		{
			Namespace:     "default",
			PodName:       "pod-no-resources",
			ContainerName: "app",
			HasCPURequest: false,
			HasMemRequest: false,
			HasCPULimit:   false,
			HasMemLimit:   false,
		},
		{
			Namespace:     "kube-system",
			PodName:       "complete-pod",
			ContainerName: "app",
			HasCPURequest: true,
			HasMemRequest: true,
			HasCPULimit:   true,
			HasMemLimit:   true,
		},
	}

	result := RenderMissingResourcesTable(containers, 0)

	if result == "" {
		t.Error("expected non-empty missing resources table")
	}

	if !strings.Contains(result, "pod-no-cpu") {
		t.Error("expected pod with missing CPU in output")
	}

	if !strings.Contains(result, "pod-no-resources") {
		t.Error("expected pod with no resources in output")
	}
}

// TestRenderPolicySummary tests policy summary rendering
func TestRenderPolicySummary(t *testing.T) {
	policies := []resources.PolicySummary{
		{
			Namespace: "default",
			LimitRanges: []resources.LimitRangeSummary{
				{
					Name: "default-limits",
				},
			},
			ResourceQuotas: []resources.ResourceQuotaSummary{
				{
					Name: "quota-default",
				},
			},
		},
		{
			Namespace: "kube-system",
			LimitRanges: []resources.LimitRangeSummary{
				{
					Name: "kube-limits",
				},
			},
		},
	}

	result := RenderPolicySummary(policies)

	if result == "" {
		t.Error("expected non-empty policy summary")
	}

	if !strings.Contains(result, "default") {
		t.Error("expected default namespace in output")
	}

	if !strings.Contains(result, "kube-system") {
		t.Error("expected kube-system namespace in output")
	}
}

// TestRenderUsageTable tests usage table rendering with various data
func TestRenderUsageTable(t *testing.T) {
	usages := []resources.ContainerUsage{
		{
			Namespace:     "default",
			PodName:       "web-pod",
			ContainerName: "web",
			CPUUsage:      *resource.NewMilliQuantity(250, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
		{
			Namespace:     "default",
			PodName:       "api-pod",
			ContainerName: "api",
			CPUUsage:      *resource.NewMilliQuantity(500, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(512*1024*1024, resource.BinarySI),
		},
		{
			Namespace:     "kube-system",
			PodName:       "etcd",
			ContainerName: "etcd",
			CPUUsage:      *resource.NewMilliQuantity(100, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(128*1024*1024, resource.BinarySI),
		},
	}

	result := RenderUsageTable(usages, 0)

	if result == "" {
		t.Error("expected non-empty usage table")
	}

	if !strings.Contains(result, "web-pod") {
		t.Error("expected web-pod in output")
	}

	if !strings.Contains(result, "api-pod") {
		t.Error("expected api-pod in output")
	}

	if !strings.Contains(result, "etcd") {
		t.Error("expected etcd pod in output")
	}
}

// TestRenderUsageTable_TopLimit tests usage table with top limit
func TestRenderUsageTable_TopLimit(t *testing.T) {
	usages := []resources.ContainerUsage{
		{
			Namespace:     "default",
			PodName:       "pod1",
			ContainerName: "container",
			CPUUsage:      *resource.NewMilliQuantity(100, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(128*1024*1024, resource.BinarySI),
		},
		{
			Namespace:     "default",
			PodName:       "pod2",
			ContainerName: "container",
			CPUUsage:      *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
		{
			Namespace:     "default",
			PodName:       "pod3",
			ContainerName: "container",
			CPUUsage:      *resource.NewMilliQuantity(300, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(512*1024*1024, resource.BinarySI),
		},
	}

	result := RenderUsageTable(usages, 2)

	if result == "" {
		t.Error("expected non-empty output with top limit")
	}

	// Should contain first two pods
	if !strings.Contains(result, "pod1") {
		t.Error("expected pod1 in top 2 output")
	}

	if !strings.Contains(result, "pod2") {
		t.Error("expected pod2 in top 2 output")
	}

	// pod3 should be excluded
	if strings.Contains(result, "pod3") {
		t.Error("expected pod3 to be excluded with top=2")
	}
}

// TestRenderDiffTable tests diff table rendering
func TestRenderDiffTable(t *testing.T) {
	diffs := []resources.ContainerDiff{
		{
			Namespace:     "default",
			PodName:       "web-pod",
			ContainerName: "web",
			CPUUsage:      *resource.NewMilliQuantity(200, resource.DecimalSI),
			CPURequest:    *resource.NewMilliQuantity(500, resource.DecimalSI),
			CPULimit:      *resource.NewMilliQuantity(1000, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(256*1024*1024, resource.BinarySI),
			MemRequest:    *resource.NewQuantity(512*1024*1024, resource.BinarySI),
			MemLimit:      *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
		},
		{
			Namespace:     "default",
			PodName:       "api-pod",
			ContainerName: "api",
			CPUUsage:      *resource.NewMilliQuantity(800, resource.DecimalSI),
			CPURequest:    *resource.NewMilliQuantity(500, resource.DecimalSI),
			CPULimit:      *resource.NewMilliQuantity(1000, resource.DecimalSI),
			MemUsage:      *resource.NewQuantity(900*1024*1024, resource.BinarySI),
			MemRequest:    *resource.NewQuantity(512*1024*1024, resource.BinarySI),
			MemLimit:      *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
		},
	}

	result := RenderDiffTable(diffs, 0)

	if result == "" {
		t.Error("expected non-empty diff table")
	}

	if !strings.Contains(result, "web-pod") {
		t.Error("expected web-pod in diff output")
	}

	if !strings.Contains(result, "api-pod") {
		t.Error("expected api-pod in diff output")
	}
}

// TestRenderPodResourceSummary tests pod resource summary rendering with top
func TestRenderPodResourceSummary_Comprehensive(t *testing.T) {
	pods := []resources.PodResourceSummary{
		{
			Namespace:  "default",
			PodName:    "prod-pod-1",
			CPURequest: *resource.NewMilliQuantity(500, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(1000, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(512*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
		},
		{
			Namespace:  "default",
			PodName:    "prod-pod-2",
			CPURequest: *resource.NewMilliQuantity(1000, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(2000, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(2048*1024*1024, resource.BinarySI),
		},
		{
			Namespace:  "kube-system",
			PodName:    "system-pod",
			CPURequest: *resource.NewMilliQuantity(100, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(200, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(128*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
		},
	}

	// Test without top limit
	result := RenderPodResourceSummary(pods, 0)
	if result == "" {
		t.Error("expected non-empty output")
	}

	if !strings.Contains(result, "prod-pod-1") {
		t.Error("expected prod-pod-1 in output")
	}

	if !strings.Contains(result, "system-pod") {
		t.Error("expected system-pod in output")
	}

	// Test with top limit
	limitedResult := RenderPodResourceSummary(pods, 2)
	if !strings.Contains(limitedResult, "prod-pod-1") {
		t.Error("expected prod-pod-1 in limited output")
	}

	if strings.Contains(limitedResult, "system-pod") {
		t.Error("expected system-pod to be excluded with top=2")
	}
}

// TestRenderPodResourceSummaryTotals tests totals rendering
func TestRenderPodResourceSummaryTotals_Comprehensive(t *testing.T) {
	pods := []resources.PodResourceSummary{
		{
			Namespace:  "default",
			PodName:    "pod1",
			CPURequest: *resource.NewMilliQuantity(500, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(1000, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(512*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
		},
		{
			Namespace:  "default",
			PodName:    "pod2",
			CPURequest: *resource.NewMilliQuantity(500, resource.DecimalSI),
			CPULimit:   *resource.NewMilliQuantity(1000, resource.DecimalSI),
			MemRequest: *resource.NewQuantity(512*1024*1024, resource.BinarySI),
			MemLimit:   *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
		},
	}

	result := RenderPodResourceSummaryTotals(pods)

	if result == "" {
		t.Error("expected non-empty totals output")
	}

	// Should have totals (1000m CPU, 1Gi memory)
	if !strings.Contains(result, "1") {
		t.Error("expected aggregated values in totals")
	}
}
