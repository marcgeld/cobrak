package capacity

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestCalculateClusterPressure_Integration tests full cluster pressure calculation
func TestCalculateClusterPressure_Integration(t *testing.T) {
	// Create test node with allocatable resources
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-1",
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}

	// Create pod with moderate resource requests
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			NodeName: "worker-1",
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(node, pod)
	ctx := context.Background()

	thresholds := PressureThresholds{
		Low:       50,
		Medium:    70,
		High:      85,
		Saturated: 95,
	}

	pressure, err := CalculatePressureWithThresholds(ctx, client, "", thresholds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pressure == nil {
		t.Error("expected non-nil pressure")
		return
	}

	// Verify cluster pressure is LOW (25% CPU, 25% memory)
	if pressure.Overall != PressureLow {
		t.Errorf("expected LOW pressure, got %s", pressure.Overall)
	}

	// Verify CPU and memory utilization
	if pressure.CPUUtilization < 20 || pressure.CPUUtilization > 30 {
		t.Errorf("expected ~25%% CPU utilization, got %.1f%%", pressure.CPUUtilization)
	}

	if pressure.MemUtilization < 20 || pressure.MemUtilization > 30 {
		t.Errorf("expected ~25%% memory utilization, got %.1f%%", pressure.MemUtilization)
	}
}

// TestNodePressure_Tracking tests individual node pressure tracking
func TestNodePressure_Tracking(t *testing.T) {
	// Create two nodes with different pressure levels
	node1 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "low-pressure",
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}

	node2 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "high-pressure",
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"), // Small node
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionTrue, // Under pressure!
				},
			},
		},
	}

	client := fake.NewSimpleClientset(node1, node2)
	ctx := context.Background()

	thresholds := PressureThresholds{
		Low:       50,
		Medium:    70,
		High:      85,
		Saturated: 95,
	}

	pressure, err := CalculatePressureWithThresholds(ctx, client, "", thresholds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify node pressures are tracked
	if len(pressure.NodePressures) != 2 {
		t.Errorf("expected 2 node pressures, got %d", len(pressure.NodePressures))
		return
	}

	// At least one node should show HIGH or SATURATED due to conditions
	foundHighPressure := false
	for _, np := range pressure.NodePressures {
		if np.MemPressure == PressureHigh || np.MemPressure == PressureSaturated {
			foundHighPressure = true
			break
		}
	}

	if !foundHighPressure {
		t.Error("expected at least one node with high pressure")
	}
}

// TestNamespacePressure_Calculation tests namespace-level pressure
func TestNamespacePressure_Calculation(t *testing.T) {
	// Create pods in different namespaces
	prodPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prod-pod",
			Namespace: "production",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
					},
				},
			},
		},
	}

	devPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dev-pod",
			Namespace: "development",
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

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	}

	client := fake.NewSimpleClientset(node, prodPod, devPod)
	ctx := context.Background()

	thresholds := PressureThresholds{
		Low:       50,
		Medium:    70,
		High:      85,
		Saturated: 95,
	}

	pressure, err := CalculatePressureWithThresholds(ctx, client, "", thresholds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify namespace pressures are calculated
	if len(pressure.NamespacePressures) != 2 {
		t.Errorf("expected 2 namespace pressures, got %d", len(pressure.NamespacePressures))
		return
	}

	// Find production namespace
	var prodNS *NamespacePressure
	for i := range pressure.NamespacePressures {
		if pressure.NamespacePressures[i].Namespace == "production" {
			prodNS = &pressure.NamespacePressures[i]
			break
		}
	}

	if prodNS == nil {
		t.Error("production namespace not found")
		return
	}

	// Production should have higher pressure (2 CPUs out of 4)
	if prodNS.CPUPercent < 40 {
		t.Errorf("expected production CPU >40%%, got %.1f%%", prodNS.CPUPercent)
	}
}

// TestPressureThresholds_Validation tests threshold validation
func TestPressureThresholds_Validation(t *testing.T) {
	tests := []struct {
		name       string
		thresholds PressureThresholds
		isValid    bool
	}{
		{
			name: "Valid thresholds",
			thresholds: PressureThresholds{
				Low:       50,
				Medium:    70,
				High:      85,
				Saturated: 95,
			},
			isValid: true,
		},
		{
			name: "Invalid - Medium >= High",
			thresholds: PressureThresholds{
				Low:       50,
				Medium:    90, // Too high!
				High:      85,
				Saturated: 95,
			},
			isValid: false,
		},
		{
			name: "Invalid - Low >= Medium",
			thresholds: PressureThresholds{
				Low:       70, // Too high!
				Medium:    70,
				High:      85,
				Saturated: 95,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify threshold ordering
			isValid := tt.thresholds.Low < tt.thresholds.Medium &&
				tt.thresholds.Medium < tt.thresholds.High &&
				tt.thresholds.High < tt.thresholds.Saturated

			if isValid != tt.isValid {
				t.Errorf("validation mismatch: expected %v, got %v", tt.isValid, isValid)
			}
		})
	}
}

// TestHighPressureDetection tests detection of high pressure situations
func TestHighPressureDetection(t *testing.T) {
	// Create a small node that's heavily loaded
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "saturated-node",
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   corev1.NodeDiskPressure,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	// Heavy pod consuming most resources
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "heavy-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			NodeName: "saturated-node",
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("450Mi"),
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(node, pod)
	ctx := context.Background()

	thresholds := PressureThresholds{
		Low:       50,
		Medium:    70,
		High:      85,
		Saturated: 95,
	}

	pressure, err := CalculatePressureWithThresholds(ctx, client, "", thresholds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should detect high pressure from both conditions and utilization
	if pressure.Overall == PressureLow || pressure.Overall == PressureMedium {
		t.Errorf("expected HIGH or SATURATED pressure, got %s", pressure.Overall)
	}
}

// TestEmptyCluster_LowPressure tests that empty cluster has low pressure
func TestEmptyCluster_LowPressure(t *testing.T) {
	client := fake.NewSimpleClientset()
	ctx := context.Background()

	thresholds := PressureThresholds{
		Low:       50,
		Medium:    70,
		High:      85,
		Saturated: 95,
	}

	pressure, err := CalculatePressureWithThresholds(ctx, client, "", thresholds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty cluster should be LOW pressure
	if pressure.Overall != PressureLow {
		t.Errorf("expected LOW pressure for empty cluster, got %s", pressure.Overall)
	}

	// No nodes or pods
	if len(pressure.NodePressures) != 0 {
		t.Errorf("expected 0 node pressures, got %d", len(pressure.NodePressures))
	}
}

// TestNamespaceFilter_Pressure tests pressure calculation with namespace filter
func TestNamespaceFilter_Pressure(t *testing.T) {
	prodPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prod-pod",
			Namespace: "production",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("2"),
						},
					},
				},
			},
		},
	}

	devPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dev-pod",
			Namespace: "development",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("4"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("4"),
			},
		},
	}

	client := fake.NewSimpleClientset(node, prodPod, devPod)
	ctx := context.Background()

	thresholds := PressureThresholds{
		Low:       50,
		Medium:    70,
		High:      85,
		Saturated: 95,
	}

	// Filter to production only
	pressure, err := CalculatePressureWithThresholds(ctx, client, "production", thresholds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have production namespace
	found := false
	for _, ns := range pressure.NamespacePressures {
		if ns.Namespace == "production" {
			found = true
		}
		if ns.Namespace == "development" {
			t.Error("development namespace should be filtered out")
		}
	}

	if !found {
		t.Error("production namespace not found after filtering")
	}
}
