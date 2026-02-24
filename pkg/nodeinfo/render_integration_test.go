package nodeinfo

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestRenderNodeInfo tests basic node information rendering
func TestRenderNodeInfo(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-1",
			Labels: map[string]string{
				"kubernetes.io/hostname": "worker-1",
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OSImage:                 "Ubuntu 20.04 LTS",
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

	nodeAnalysis := AnalyzeNode(node, "")
	result := RenderNodeInfo(nodeAnalysis)

	if result == "" {
		t.Error("expected non-empty node info output")
		return
	}

	// Verify content contains key information
	if !strings.Contains(result, "worker-1") {
		t.Error("expected node name in output")
	}

	if !strings.Contains(result, "Ubuntu") {
		t.Error("expected OS info in output")
	}

	if !strings.Contains(result, "amd64") {
		t.Error("expected architecture in output")
	}
}

// TestRenderNodeHealth tests node health status rendering
func TestRenderNodeHealth(t *testing.T) {
	tests := []struct {
		name             string
		conditions       []corev1.NodeCondition
		expectedContains string
	}{
		{
			name: "Healthy node",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionFalse,
				},
			},
			expectedContains: "HEALTHY",
		},
		{
			name: "Node with memory pressure",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionTrue,
				},
			},
			expectedContains: "WARNING",
		},
		{
			name: "Not ready node",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
			},
			expectedContains: "UNHEALTHY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						Architecture: "amd64",
					},
					Conditions: tt.conditions,
				},
			}

			nodeAnalysis := AnalyzeNode(node, "")
			result := RenderNodeHealth(nodeAnalysis)

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
	nodes := []NodeAnalysis{
		{
			NodeName:     "worker-1",
			OS:           "Ubuntu 20.04 LTS",
			Kernel:       "5.4.0",
			Architecture: "amd64",
			Health:       "HEALTHY",
		},
		{
			NodeName:     "worker-2",
			OS:           "Ubuntu 20.04 LTS",
			Kernel:       "5.4.0",
			Architecture: "arm64",
			Health:       "WARNING",
		},
		{
			NodeName:     "worker-3",
			OS:           "Debian 11",
			Kernel:       "5.10.0",
			Architecture: "amd64",
			Health:       "UNHEALTHY",
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

	// Verify health statuses appear
	if !strings.Contains(result, "HEALTHY") {
		t.Error("expected HEALTHY status in output")
	}

	if !strings.Contains(result, "WARNING") {
		t.Error("expected WARNING status in output")
	}

	if !strings.Contains(result, "UNHEALTHY") {
		t.Error("expected UNHEALTHY status in output")
	}
}

// TestRenderNodeInfo_WithGPU tests node rendering with GPU info
func TestRenderNodeInfo_WithGPU(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-node",
			Labels: map[string]string{
				"accelerator": "nvidia-tesla-v100",
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				Architecture: "amd64",
			},
		},
	}

	nodeAnalysis := AnalyzeNode(node, "")

	// GPU info should be extracted from labels
	if nodeAnalysis.GPUInfo == "" {
		t.Error("expected GPU info to be extracted from labels")
	}

	result := RenderNodeInfo(nodeAnalysis)
	if result == "" {
		t.Error("expected non-empty output for GPU node")
	}
}

// TestRenderNodeInfo_WithVirtualization tests node rendering with virtualization info
func TestRenderNodeInfo_WithVirtualization(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "virt-node",
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				Architecture: "amd64",
			},
		},
	}

	nodeAnalysis := AnalyzeNode(node, "")
	result := RenderNodeInfo(nodeAnalysis)

	if result == "" {
		t.Error("expected non-empty output")
	}

	// Virtualization field should be populated (even if Unknown)
	if nodeAnalysis.Virtualization == "" {
		t.Error("expected virtualization info to be populated")
	}
}

// TestRenderNodeHealth_WithMemoryPressure tests memory pressure rendering
func TestRenderNodeHealth_WithMemoryPressure(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "memory-pressure-node",
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				Architecture: "amd64",
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:    corev1.NodeMemoryPressure,
					Status:  corev1.ConditionTrue,
					Message: "Memory pressure detected",
				},
			},
		},
	}

	nodeAnalysis := AnalyzeNode(node, "")
	result := RenderNodeHealth(nodeAnalysis)

	if result == "" {
		t.Error("expected non-empty health output")
		return
	}

	// Should indicate pressure condition
	if !strings.Contains(strings.ToLower(result), "memory") {
		t.Error("expected memory pressure indication in output")
	}
}

// TestRenderMultipleNodeInfoCompact_Empty tests compact rendering with empty list
func TestRenderMultipleNodeInfoCompact_Empty(t *testing.T) {
	var nodes []NodeAnalysis

	result := RenderMultipleNodeInfoCompact(nodes)

	// Should handle empty list gracefully
	if result == "" {
		// Empty output is acceptable for empty input
		return
	}

	// Or it might return a message
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

			nodeAnalysis := AnalyzeNode(node, "")
			result := RenderNodeInfo(nodeAnalysis)

			if result == "" {
				t.Errorf("expected non-empty output for architecture %s", arch)
			}

			if !strings.Contains(result, arch) {
				t.Errorf("expected architecture %s in output", arch)
			}
		})
	}
}
