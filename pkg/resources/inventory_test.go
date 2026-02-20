package resources

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBuildInventory_Empty(t *testing.T) {
	client := fake.NewSimpleClientset()
	nsInv, containers, policies, err := BuildInventory(context.Background(), client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nsInv) != 0 {
		t.Errorf("expected 0 namespaces, got %d", len(nsInv))
	}
	if len(containers) != 0 {
		t.Errorf("expected 0 containers, got %d", len(containers))
	}
	if len(policies) != 0 {
		t.Errorf("expected 0 policies, got %d", len(policies))
	}
}

func TestBuildInventory_WithPods(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "container1",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("200m"),
							v1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
				{
					Name: "container2",
					// No requests or limits
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	nsInv, containers, _, err := BuildInventory(context.Background(), client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nsInv) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(nsInv))
	}
	if nsInv[0].ContainersTotal != 2 {
		t.Errorf("expected 2 containers, got %d", nsInv[0].ContainersTotal)
	}
	if nsInv[0].ContainersMissingAnyRequests != 1 {
		t.Errorf("expected 1 container missing requests, got %d", nsInv[0].ContainersMissingAnyRequests)
	}
	if nsInv[0].ContainersMissingAnyLimits != 1 {
		t.Errorf("expected 1 container missing limits, got %d", nsInv[0].ContainersMissingAnyLimits)
	}
	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}
}
