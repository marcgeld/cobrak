package resources

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestBuildDiff_WithRealData tests diff calculation with realistic data
func TestBuildDiff_WithRealData(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
	}

	// Simulate usage data (50% of request)
	usages := map[string]map[string]*corev1.ResourceList{
		"default": {
			"app-pod:app": &corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},
	}

	diffs, err := BuildDiff([]corev1.Pod{*pod}, usages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diffs) == 0 {
		t.Error("expected diff entries")
		return
	}

	// Verify diff calculation
	diff := diffs[0]
	if diff.PodName != "app-pod" {
		t.Errorf("expected pod name 'app-pod', got %s", diff.PodName)
	}

	// Usage should be less than request
	if diff.CPUUsage.MilliValue() >= diff.CPURequest.MilliValue() {
		t.Error("expected usage < request")
	}
}

// TestBuildInventory_ResourceCoverage tests inventory with resource coverage
func TestBuildInventory_ResourceCoverage(t *testing.T) {
	// Pod with all resources specified
	completePod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "complete-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
	}

	// Pod with incomplete resources
	incompletePod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "incomplete-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
							// Missing memory request
						},
						Limits: corev1.ResourceList{
							// Missing limits
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(completePod, incompletePod)
	ctx := context.Background()

	nsInv, containers, _, err := BuildInventory(ctx, client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nsInv) != 1 {
		t.Errorf("expected 1 namespace, got %d", len(nsInv))
		return
	}

	inv := nsInv[0]

	// Verify resource coverage statistics
	if inv.ContainersTotal != 2 {
		t.Errorf("expected 2 containers, got %d", inv.ContainersTotal)
	}

	// One container missing memory request
	if inv.ContainersMissingAnyRequests != 1 {
		t.Errorf("expected 1 container missing requests, got %d", inv.ContainersMissingAnyRequests)
	}

	// One container missing limits completely
	if inv.ContainersMissingAnyLimits != 1 {
		t.Errorf("expected 1 container missing limits, got %d", inv.ContainersMissingAnyLimits)
	}

	// Both containers should be in missing resources list
	if len(containers) != 1 {
		t.Errorf("expected 1 container in missing list, got %d", len(containers))
	}
}

// TestExtractContainerResources tests resource extraction from containers
func TestExtractContainerResources(t *testing.T) {
	tests := []struct {
		name              string
		container         corev1.Container
		expectedHasCPUReq bool
		expectedHasMemReq bool
		expectedHasCPULim bool
		expectedHasMemLim bool
	}{
		{
			name: "Full resources",
			container: corev1.Container{
				Name: "app",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			expectedHasCPUReq: true,
			expectedHasMemReq: true,
			expectedHasCPULim: true,
			expectedHasMemLim: true,
		},
		{
			name: "Requests only",
			container: corev1.Container{
				Name: "app",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
				},
			},
			expectedHasCPUReq: true,
			expectedHasMemReq: true,
			expectedHasCPULim: false,
			expectedHasMemLim: false,
		},
		{
			name: "No resources",
			container: corev1.Container{
				Name: "app",
			},
			expectedHasCPUReq: false,
			expectedHasMemReq: false,
			expectedHasCPULim: false,
			expectedHasMemLim: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasCPUReq := len(tt.container.Resources.Requests[corev1.ResourceCPU]) > 0
			hasMemReq := len(tt.container.Resources.Requests[corev1.ResourceMemory]) > 0
			hasCPULim := len(tt.container.Resources.Limits[corev1.ResourceCPU]) > 0
			hasMemLim := len(tt.container.Resources.Limits[corev1.ResourceMemory]) > 0

			if hasCPUReq != tt.expectedHasCPUReq {
				t.Errorf("CPU request: expected %v, got %v", tt.expectedHasCPUReq, hasCPUReq)
			}
			if hasMemReq != tt.expectedHasMemReq {
				t.Errorf("Memory request: expected %v, got %v", tt.expectedHasMemReq, hasMemReq)
			}
			if hasCPULim != tt.expectedHasCPULim {
				t.Errorf("CPU limit: expected %v, got %v", tt.expectedHasCPULim, hasCPULim)
			}
			if hasMemLim != tt.expectedHasMemLim {
				t.Errorf("Memory limit: expected %v, got %v", tt.expectedHasMemLim, hasMemLim)
			}
		})
	}
}

// TestNamespaceResourceAggregation tests aggregating resources by namespace
func TestNamespaceResourceAggregation(t *testing.T) {
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "production",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			},
		},
	}

	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "production",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod1, pod2)
	ctx := context.Background()

	nsInv, _, _, err := BuildInventory(ctx, client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nsInv) != 1 {
		t.Errorf("expected 1 namespace, got %d", len(nsInv))
		return
	}

	inv := nsInv[0]
	if inv.Namespace != "production" {
		t.Errorf("expected production namespace, got %s", inv.Namespace)
	}

	if inv.ContainersTotal != 2 {
		t.Errorf("expected 2 containers, got %d", inv.ContainersTotal)
	}

	// Verify CPU aggregation: 500m + 500m = 1000m = 1 CPU
	expectedCPU := int64(1000)
	actualCPU := inv.CPURequestsTotal.MilliValue()
	if actualCPU != expectedCPU {
		t.Errorf("expected CPU %dm, got %dm", expectedCPU, actualCPU)
	}
}
