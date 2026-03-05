package resources

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// MockMetricsReader is a mock implementation for testing metrics reading
type MockMetricsReader struct {
	usages    []ContainerUsage
	available bool
	err       error
}

func (m *MockMetricsReader) IsAvailable(ctx context.Context) (bool, error) {
	return m.available, nil
}

func (m *MockMetricsReader) PodMetrics(ctx context.Context, namespace string) ([]ContainerUsage, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.available {
		return nil, fmt.Errorf("metrics unavailable")
	}
	if namespace == "" {
		return m.usages, nil
	}
	var filtered []ContainerUsage
	for _, u := range m.usages {
		if u.Namespace == namespace {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

// TestBuildPodSummariesWithUsage_Integration tests pod summary building with metrics
func TestBuildPodSummariesWithUsage_Integration(t *testing.T) {
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

	// Mock metrics: 50% of request
	mockMetrics := &MockMetricsReader{
		available: true,
		usages: []ContainerUsage{
			{
				Namespace:     "default",
				PodName:       "app-pod",
				ContainerName: "app",
				CPUUsage:      resource.MustParse("250m"),
				MemUsage:      resource.MustParse("256Mi"),
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	ctx := context.Background()

	summaries, err := BuildPodSummariesWithUsage(ctx, client, mockMetrics, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
		return
	}

	// Verify summary has usage data
	summary := summaries[0]
	if summary.PodName != "app-pod" {
		t.Errorf("expected pod name 'app-pod', got %s", summary.PodName)
	}

	// Verify usage is populated
	if summary.CPUUsage.IsZero() {
		t.Error("expected non-zero CPU usage")
	}

	if summary.MemUsage.IsZero() {
		t.Error("expected non-zero memory usage")
	}
}

// TestBuildPodSummariesWithUsage_NoMetrics tests behavior when metrics unavailable
func TestBuildPodSummariesWithUsage_NoMetrics(t *testing.T) {
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
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
	}

	// Metrics unavailable
	mockMetrics := &MockMetricsReader{
		available: false,
	}

	client := fake.NewSimpleClientset(pod)
	ctx := context.Background()

	summaries, err := BuildPodSummariesWithUsage(ctx, client, mockMetrics, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
		return
	}

	// Should still have request data
	if summaries[0].CPURequest.IsZero() {
		t.Error("expected non-zero CPU request even without metrics")
	}
}

// TestBuildPodSummariesWithUsage_MultipleContainers tests with multiple containers
func TestBuildPodSummariesWithUsage_MultipleContainers(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "multi-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				{
					Name: "container2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
	}

	// Metrics for both containers
	mockMetrics := &MockMetricsReader{
		available: true,
		usages: []ContainerUsage{
			{
				Namespace:     "default",
				PodName:       "multi-pod",
				ContainerName: "container1",
				CPUUsage:      resource.MustParse("80m"),
				MemUsage:      resource.MustParse("100Mi"),
			},
			{
				Namespace:     "default",
				PodName:       "multi-pod",
				ContainerName: "container2",
				CPUUsage:      resource.MustParse("70m"),
				MemUsage:      resource.MustParse("92Mi"),
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	ctx := context.Background()

	summaries, err := BuildPodSummariesWithUsage(ctx, client, mockMetrics, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
		return
	}

	// Should aggregate resources from both containers
	expectedCPU := int64(300) // 100m + 200m
	actualCPU := summaries[0].CPURequest.MilliValue()
	if actualCPU != expectedCPU {
		t.Errorf("expected aggregated CPU %dm, got %dm", expectedCPU, actualCPU)
	}
}

// TestBuildPodSummariesWithUsage_NamespaceFilter tests filtering by namespace
func TestBuildPodSummariesWithUsage_NamespaceFilter(t *testing.T) {
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
				},
			},
		},
	}

	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-pod",
			Namespace: "kube-system",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "system",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
				},
			},
		},
	}

	mockMetrics := &MockMetricsReader{
		available: true,
		usages: []ContainerUsage{
			{Namespace: "default", PodName: "default-pod", ContainerName: "app", CPUUsage: resource.MustParse("50m")},
			{Namespace: "kube-system", PodName: "kube-pod", ContainerName: "system", CPUUsage: resource.MustParse("50m")},
		},
	}

	client := fake.NewSimpleClientset(pod1, pod2)
	ctx := context.Background()

	// Filter to default namespace
	summaries, err := BuildPodSummariesWithUsage(ctx, client, mockMetrics, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary (filtered), got %d", len(summaries))
		return
	}

	if summaries[0].Namespace != "default" {
		t.Errorf("expected 'default' namespace, got %s", summaries[0].Namespace)
	}
}

// TestBuildPodSummariesWithUsage_UsageVsRequest tests usage comparison with requests
func TestBuildPodSummariesWithUsage_UsageVsRequest(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),   // 1000m
							corev1.ResourceMemory: resource.MustParse("1Gi"), // 1024Mi
						},
					},
				},
			},
		},
	}

	// Usage is much lower than request (efficiency test)
	mockMetrics := &MockMetricsReader{
		available: true,
		usages: []ContainerUsage{
			{
				Namespace:     "default",
				PodName:       "test-pod",
				ContainerName: "app",
				CPUUsage:      resource.MustParse("100m"),  // 10% utilization
				MemUsage:      resource.MustParse("100Mi"), // ~10% utilization
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	ctx := context.Background()

	summaries, err := BuildPodSummariesWithUsage(ctx, client, mockMetrics, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
		return
	}

	// Verify usage is less than request
	if summaries[0].CPUUsage.MilliValue() >= summaries[0].CPURequest.MilliValue() {
		t.Error("expected usage < request for this test case")
	}

	if summaries[0].MemUsage.Value() >= summaries[0].MemRequest.Value() {
		t.Error("expected memory usage < request for this test case")
	}
}
