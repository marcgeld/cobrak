package resources

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestBuildPodSummaries_Integration tests pod summary building with fake Kubernetes client
func TestBuildPodSummaries_Integration(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "app-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	ctx := context.Background()

	summaries, err := BuildPodSummaries(ctx, client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}

	if summaries[0].PodName != "app-pod" {
		t.Errorf("expected pod name 'app-pod', got %s", summaries[0].PodName)
	}

	if summaries[0].Namespace != "default" {
		t.Errorf("expected namespace 'default', got %s", summaries[0].Namespace)
	}

	// Verify resources were extracted
	if summaries[0].CPURequest.IsZero() {
		t.Error("expected non-zero CPU request")
	}

	if summaries[0].MemRequest.IsZero() {
		t.Error("expected non-zero memory request")
	}
}

func TestBuildPodSummaries_MultipleNamespaces(t *testing.T) {
	pods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default-pod",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "container",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-pod",
				Namespace: "kube-system",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "container",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pods[0], pods[1])
	ctx := context.Background()

	summaries, err := BuildPodSummaries(ctx, client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(summaries))
	}

	namespaces := make(map[string]bool)
	for _, s := range summaries {
		namespaces[s.Namespace] = true
	}

	if !namespaces["default"] {
		t.Error("expected default namespace")
	}
	if !namespaces["kube-system"] {
		t.Error("expected kube-system namespace")
	}
}

func TestBuildPodSummaries_WithNamespaceFilter(t *testing.T) {
	pods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default-pod",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "container",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("100m"),
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-pod",
				Namespace: "kube-system",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "container",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("100m"),
							},
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pods[0], pods[1])
	ctx := context.Background()

	summaries, err := BuildPodSummaries(ctx, client, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary (filtered), got %d", len(summaries))
	}

	if summaries[0].Namespace != "default" {
		t.Errorf("expected 'default' namespace, got %s", summaries[0].Namespace)
	}
}

func TestBuildPodSummaries_EmptyCluster(t *testing.T) {
	client := fake.NewSimpleClientset()
	ctx := context.Background()

	summaries, err := BuildPodSummaries(ctx, client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries for empty cluster, got %d", len(summaries))
	}
}

func TestBuildPodSummaries_MissingResources(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "no-resources-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container",
					// No resources specified
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	ctx := context.Background()

	summaries, err := BuildPodSummaries(ctx, client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}

	// Should handle missing resources gracefully
	if !summaries[0].CPURequest.IsZero() {
		t.Errorf("expected zero CPU request, got %s", summaries[0].CPURequest.String())
	}

	if !summaries[0].MemRequest.IsZero() {
		t.Errorf("expected zero memory request, got %s", summaries[0].MemRequest.String())
	}
}

func TestBuildPodSummaries_MultipleContainers(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "multi-container-pod",
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

	client := fake.NewSimpleClientset(pod)
	ctx := context.Background()

	summaries, err := BuildPodSummaries(ctx, client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}

	// Should aggregate CPU from both containers
	expectedCPU := int64(300) // 100m + 200m
	if summaries[0].CPURequest.MilliValue() != expectedCPU {
		t.Errorf("expected aggregated CPU %dm, got %dm", expectedCPU, summaries[0].CPURequest.MilliValue())
	}

	// Should aggregate memory from both containers
	expectedMem := int64(128*1024*1024 + 256*1024*1024)
	if summaries[0].MemRequest.Value() != expectedMem {
		t.Errorf("expected aggregated memory %d, got %d", expectedMem, summaries[0].MemRequest.Value())
	}
}
