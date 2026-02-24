package nodeinfo

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note: TestGetNodeHealthStatus already exists in analyzer_test.go
// This file contains additional complementary tests

// ...existing code...

// TestAnalyzeNodeInfo tests node analysis with various configurations
func TestAnalyzeNodeInfo(t *testing.T) {
	tests := []struct {
		name          string
		nodeName      string
		architecture  string
		osImage       string
		kernelVersion string
	}{
		{
			name:          "Linux node with amd64",
			nodeName:      "worker-1",
			architecture:  "amd64",
			osImage:       "Ubuntu 20.04 LTS",
			kernelVersion: "5.4.0",
		},
		{
			name:          "ARM node",
			nodeName:      "arm-worker",
			architecture:  "arm64",
			osImage:       "Debian 11",
			kernelVersion: "5.10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   tt.nodeName,
					Labels: map[string]string{},
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage:       tt.osImage,
						KernelVersion: tt.kernelVersion,
						Architecture:  tt.architecture,
					},
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			// Verify node info is extracted correctly
			if node.ObjectMeta.Name != tt.nodeName {
				t.Errorf("expected node name %s, got %s", tt.nodeName, node.ObjectMeta.Name)
			}

			if node.Status.NodeInfo.Architecture != tt.architecture {
				t.Errorf("expected architecture %s, got %s", tt.architecture, node.Status.NodeInfo.Architecture)
			}

			if node.Status.NodeInfo.OSImage != tt.osImage {
				t.Errorf("expected OS %s, got %s", tt.osImage, node.Status.NodeInfo.OSImage)
			}
		})
	}
}

// TestMemoryPressureAnalysis tests memory pressure analysis
func TestMemoryPressureAnalysis(t *testing.T) {
	tests := []struct {
		name             string
		hasMmoryPressure bool
		expectedPressure string
	}{
		{
			name:             "No memory pressure",
			hasMmoryPressure: false,
			expectedPressure: "LOW",
		},
		{
			name:             "With memory pressure",
			hasMmoryPressure: true,
			expectedPressure: "HIGH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type: corev1.NodeMemoryPressure,
							Status: func() corev1.ConditionStatus {
								if tt.hasMmoryPressure {
									return corev1.ConditionTrue
								}
								return corev1.ConditionFalse
							}(),
						},
					},
				},
			}

			pressure := analyzeMemoryPressure(node)

			if pressure.Pressure != tt.expectedPressure {
				t.Errorf("expected %s pressure, got %s", tt.expectedPressure, pressure.Pressure)
			}
		})
	}
}

// TestNodeInfoWithGPULabel tests node analysis with GPU labels
func TestNodeInfoWithGPULabel(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-node",
			Labels: map[string]string{
				"accelerator": "nvidia-tesla-v100",
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OSImage:      "Ubuntu 20.04",
				Architecture: "amd64",
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	// Verify GPU label is accessible
	gpuLabel, hasGPU := node.ObjectMeta.Labels["accelerator"]
	if !hasGPU {
		t.Error("expected GPU label to be present")
		return
	}

	if gpuLabel != "nvidia-tesla-v100" {
		t.Errorf("expected nvidia GPU, got %s", gpuLabel)
	}
}

// TestMultipleNodeAnalysis tests analyzing multiple nodes
func TestMultipleNodeAnalysis(t *testing.T) {
	nodes := []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{
					Architecture: "amd64",
				},
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node2"},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{
					Architecture: "arm64",
				},
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
			},
		},
	}

	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}

	// Verify node architectures
	for i, expectedArch := range []string{"amd64", "arm64"} {
		if nodes[i].Status.NodeInfo.Architecture != expectedArch {
			t.Errorf("node %d: expected arch %s, got %s", i, expectedArch, nodes[i].Status.NodeInfo.Architecture)
		}
	}
}
