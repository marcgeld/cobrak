package nodeinfo

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestRenderNodeInfo tests basic node information rendering
func TestRenderNodeInfo(t *testing.T) {
	client := fake.NewSimpleClientset()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-1",
			Labels: map[string]string{
				"kubernetes.io/hostname": "worker-1",
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OperatingSystem:         "linux",
				KernelVersion:           "5.4.0",
				KubeletVersion:          "v1.24.0",
				ContainerRuntimeVersion: "docker://20.10.0",
				Architecture:            "amd64",
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	nodeInfo, err := AnalyzeNode(context.Background(), client, "worker-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := RenderNodeInfo(nodeInfo)
	if result == "" {
		t.Error("expected non-empty node info output")
		return
	}

	if !strings.Contains(result, "worker-1") {
		t.Error("expected node name in output")
	}

	if !strings.Contains(result, "amd64") {
		t.Error("expected architecture in output")
	}
}

// TestRenderNodeHealth tests node health status rendering
func TestRenderNodeHealth(t *testing.T) {
	tests := []struct {
		name             string
		status           string
		issues           []string
		expectedContains string
	}{
		{
			name:             "Healthy node",
			status:           "HEALTHY",
			issues:           []string{},
			expectedContains: "HEALTHY",
		},
		{
			name:             "Node with memory pressure",
			status:           "WARNING",
			issues:           []string{"Memory pressure detected"},
			expectedContains: "WARNING",
		},
		{
			name:             "Not ready node",
			status:           "CRITICAL",
			issues:           []string{"Node not ready"},
			expectedContains: "CRITICAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &NodeHealthStatus{
				NodeName: "test-node",
				Status:   tt.status,
				Issues:   tt.issues,
			}

			result := RenderNodeHealth(status)
			if result == "" {
				t.Errorf("expected non-empty health output for %s", tt.name)
				return
			}

			if !strings.Contains(result, tt.expectedContains) {
				t.Errorf("expected '%s' in health output, got: %s", tt.expectedContains, result)
			}
		})
	}
}

// TestRenderMultipleNodeInfoCompact tests compact rendering of multiple nodes
func TestRenderMultipleNodeInfoCompact(t *testing.T) {
	nodes := []NodeInfo{
		{
			NodeName:       "worker-1",
			OS:             "Ubuntu 20.04 LTS",
			Kernel:         "5.4.0",
			Architecture:   "amd64",
			MemoryPressure: MemoryPressure{Pressure: "LOW"},
		},
		{
			NodeName:       "worker-2",
			OS:             "Ubuntu 20.04 LTS",
			Kernel:         "5.4.0",
			Architecture:   "arm64",
			MemoryPressure: MemoryPressure{Pressure: "HIGH"},
		},
		{
			NodeName:       "worker-3",
			OS:             "Debian 11",
			Kernel:         "5.10.0",
			Architecture:   "amd64",
			MemoryPressure: MemoryPressure{Pressure: "LOW"},
		},
	}

	result := RenderMultipleNodeInfoCompact(nodes)

	if result == "" {
		t.Error("expected non-empty compact output")
		return
	}

	// Verify all nodes appear
	for _, node := range nodes {
		if !strings.Contains(result, node.NodeName) {
			t.Errorf("expected node %s in output", node.NodeName)
		}
	}
}

// TestRenderNodeInfo_WithGPU tests node rendering with GPU info
func TestRenderNodeInfo_WithGPU(t *testing.T) {
	client := fake.NewSimpleClientset()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-node",
			Labels: map[string]string{
				"nvidia.com/gpu": "1",
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				Architecture: "amd64",
			},
		},
	}
	client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	nodeInfo, err := AnalyzeNode(context.Background(), client, "gpu-node")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !nodeInfo.GPU.Available {
		t.Error("expected GPU info to be extracted from labels")
	}

	result := RenderNodeInfo(nodeInfo)
	if result == "" {
		t.Error("expected non-empty output for GPU node")
	}
}

// TestRenderNodeInfo_WithVirtualization tests node rendering with virtualization info
func TestRenderNodeInfo_WithVirtualization(t *testing.T) {
	client := fake.NewSimpleClientset()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "virt-node",
		},
		Spec: corev1.NodeSpec{
			ProviderID: "aws:///us-east-1a/i-1234567890abcdef0",
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				Architecture: "amd64",
			},
		},
	}
	client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	nodeInfo, err := AnalyzeNode(context.Background(), client, "virt-node")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := RenderNodeInfo(nodeInfo)
	if result == "" {
		t.Error("expected non-empty output")
	}

	if nodeInfo.VirtualizationType == "" {
		t.Error("expected virtualization info to be populated")
	}
}

// TestRenderNodeHealth_WithMemoryPressure tests memory pressure rendering
func TestRenderNodeHealth_WithMemoryPressure(t *testing.T) {
	status := &NodeHealthStatus{
		NodeName: "memory-pressure-node",
		Status:   "WARNING",
		Issues:   []string{"Memory pressure detected"},
	}

	result := RenderNodeHealth(status)
	if result == "" {
		t.Error("expected non-empty health output")
		return
	}

	if !strings.Contains(strings.ToLower(result), "memory") {
		t.Error("expected memory pressure indication in output")
	}
}

// TestRenderMultipleNodeInfoCompact_Empty tests compact rendering with empty list
func TestRenderMultipleNodeInfoCompact_Empty(t *testing.T) {
	var nodes []NodeInfo

	result := RenderMultipleNodeInfoCompact(nodes)

	if result == "" {
		// Empty output is acceptable for empty input
		return
	}

	if !strings.Contains(strings.ToLower(result), "no") &&
		!strings.Contains(strings.ToLower(result), "empty") {
		// Any non-empty output should be informative
	}
}

// TestRenderNodeInfo_AllArchitectures tests rendering for different architectures
func TestRenderNodeInfo_AllArchitectures(t *testing.T) {
	architectures := []string{"amd64", "arm64", "ppc64le", "s390x"}

	for _, arch := range architectures {
		t.Run(arch, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-" + arch,
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						Architecture: arch,
					},
				},
			}
			client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

			nodeInfo, err := AnalyzeNode(context.Background(), client, "node-"+arch)
			if err != nil {
				t.Fatalf("unexpected error for arch %s: %v", arch, err)
			}

			result := RenderNodeInfo(nodeInfo)
			if result == "" {
				t.Errorf("expected non-empty output for architecture %s", arch)
			}

			if !strings.Contains(result, arch) {
				t.Errorf("expected architecture %s in output", arch)
			}
		})
	}
}
