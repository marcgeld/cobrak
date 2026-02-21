package nodeinfo

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeNode(t *testing.T) {
	// Create a fake Kubernetes client with a test node
	client := fake.NewSimpleClientset()

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				"kubernetes.io/os": "linux",
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OperatingSystem:         "linux",
				KernelVersion:           "5.10.0",
				Architecture:            "amd64",
				KubeletVersion:          "v1.24.0",
				ContainerRuntimeVersion: "docker://20.10.12",
			},
			Capacity: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(8*1024*1024*1024, resource.BinarySI),
			},
			Allocatable: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(8*1024*1024*1024, resource.BinarySI),
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

	client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	// Test AnalyzeNode
	info, err := AnalyzeNode(context.Background(), client, "test-node")
	if err != nil {
		t.Fatalf("AnalyzeNode failed: %v", err)
	}

	if info.NodeName != "test-node" {
		t.Errorf("Expected node name 'test-node', got '%s'", info.NodeName)
	}

	if info.OS != "linux" {
		t.Errorf("Expected OS 'linux', got '%s'", info.OS)
	}

	if info.Kernel != "5.10.0" {
		t.Errorf("Expected kernel '5.10.0', got '%s'", info.Kernel)
	}

	if info.CPU.Count != 4 {
		t.Errorf("Expected 4 CPU cores, got %d", info.CPU.Count)
	}

	if info.ContainerRuntime.Name != "docker" {
		t.Errorf("Expected container runtime 'docker', got '%s'", info.ContainerRuntime.Name)
	}

	if info.MemoryPressure.Pressure != "LOW" {
		t.Errorf("Expected memory pressure 'LOW', got '%s'", info.MemoryPressure.Pressure)
	}
}

func TestGetNodeHealthStatus(t *testing.T) {
	client := fake.NewSimpleClientset()

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "healthy-node",
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OperatingSystem: "linux",
				KernelVersion:   "5.10.0",
				Architecture:    "amd64",
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

	client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	status, err := GetNodeHealthStatus(context.Background(), client, "healthy-node")
	if err != nil {
		t.Fatalf("GetNodeHealthStatus failed: %v", err)
	}

	if status.NodeName != "healthy-node" {
		t.Errorf("Expected node name 'healthy-node', got '%s'", status.NodeName)
	}

	if status.Status != "HEALTHY" {
		t.Errorf("Expected status 'HEALTHY', got '%s'", status.Status)
	}
}

func TestRenderNodeInfoCompact(t *testing.T) {
	info := &NodeInfo{
		NodeName:     "node-1",
		OS:           "linux",
		Architecture: "amd64",
		CPU: CPUInfo{
			Count: 4,
		},
		GPU: GPUInfo{
			Available: false,
		},
		MemoryPressure: MemoryPressure{
			Pressure: "LOW",
		},
		ContainerRuntime: ContainerRuntime{
			Name: "docker",
		},
		VirtualizationType: "AWS EC2",
	}

	output := RenderNodeInfoCompact(info)
	if output == "" {
		t.Errorf("RenderNodeInfoCompact returned empty output")
	}

	if !contains(output, "node-1") {
		t.Errorf("Output doesn't contain node name")
	}

	if !contains(output, "docker") {
		t.Errorf("Output doesn't contain container runtime")
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
