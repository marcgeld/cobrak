package output

import (
	"strings"
	"testing"
)

// TestParseOutputFormat tests output format parsing
func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  OutputFormat
		shouldErr bool
	}{
		{
			name:      "text format",
			input:     "text",
			expected:  FormatText,
			shouldErr: false,
		},
		{
			name:      "json format",
			input:     "json",
			expected:  FormatJSON,
			shouldErr: false,
		},
		{
			name:      "yaml format",
			input:     "yaml",
			expected:  FormatYAML,
			shouldErr: false,
		},
		{
			name:      "uppercase text",
			input:     "TEXT",
			expected:  FormatText,
			shouldErr: true, // Should be lowercase
		},
		{
			name:      "invalid format",
			input:     "invalid",
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "empty format",
			input:     "",
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "whitespace",
			input:     "  json  ",
			expected:  "",
			shouldErr: true, // Should fail on whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseOutputFormat(tt.input)

			if (err != nil) != tt.shouldErr {
				if tt.shouldErr {
					t.Errorf("expected error, got none")
				} else {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if !tt.shouldErr && result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestRenderOutput_AllFormats tests RenderOutput with all formats
func TestRenderOutput_AllFormats(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
	}{
		{
			name:   "text format",
			format: FormatText,
		},
		{
			name:   "json format",
			format: FormatJSON,
		},
		{
			name:   "yaml format",
			format: FormatYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal valid ResourcesSummary
			summary := &ResourcesSummary{
				Cluster: &ClusterCapacitySummary{
					CPUCapacity:    "4",
					MemCapacity:    "8Gi",
					CPURequests:    "2",
					MemRequests:    "4Gi",
					CPULimits:      "3",
					MemLimits:      "6Gi",
					CPUAllocatable: "3.8",
					MemAllocatable: "7.5Gi",
				},
				PodDetails:       []PodCapacitySummary{},
				Inventory:        []NamespaceCapacitySummary{},
				MetricsAvailable: false,
			}

			result, err := RenderOutput(summary, tt.format)
			if err != nil {
				t.Fatalf("format %s failed: %v", tt.format, err)
			}

			if result == "" {
				t.Errorf("format %s produced empty output", tt.format)
			}

			// Verify format-specific content
			switch tt.format {
			case FormatText:
				if !strings.Contains(result, "Capacity") {
					t.Error("text format missing Capacity")
				}
			case FormatJSON:
				if !strings.Contains(result, "{") && !strings.Contains(result, "}") {
					t.Error("JSON format invalid")
				}
			case FormatYAML:
				if !strings.Contains(result, ":") {
					t.Error("YAML format missing colons")
				}
			}
		})
	}
}

// TestRenderOutput_WithPodDetails tests RenderOutput with pod data
func TestRenderOutput_WithPodDetails(t *testing.T) {
	summary := &ResourcesSummary{
		Cluster: &ClusterCapacitySummary{
			CPUCapacity: "4",
			MemCapacity: "8Gi",
		},
		PodDetails: []PodCapacitySummary{
			{
				Namespace:  "default",
				Pod:        "web-pod",
				CPURequest: "500m",
				CPULimit:   "1000m",
				MemRequest: "512Mi",
				MemLimit:   "1Gi",
			},
			{
				Namespace:  "default",
				Pod:        "api-pod",
				CPURequest: "500m",
				CPULimit:   "1000m",
				MemRequest: "512Mi",
				MemLimit:   "1Gi",
			},
		},
		Inventory:        []NamespaceCapacitySummary{},
		MetricsAvailable: false,
	}

	result, err := RenderOutput(summary, FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty output with pod details")
		return
	}

	// Verify pod names appear
	if !strings.Contains(result, "web-pod") {
		t.Error("expected web-pod in output")
	}

	if !strings.Contains(result, "api-pod") {
		t.Error("expected api-pod in output")
	}
}

// TestRenderOutput_WithInventory tests RenderOutput with namespace inventory
func TestRenderOutput_WithInventory(t *testing.T) {
	summary := &ResourcesSummary{
		Cluster: &ClusterCapacitySummary{
			CPUCapacity: "4",
		},
		PodDetails: []PodCapacitySummary{},
		Inventory: []NamespaceCapacitySummary{
			{
				Namespace:        "default",
				ContainersTotal:  5,
				CPURequestsTotal: "1",
				MemRequestsTotal: "1Gi",
			},
			{
				Namespace:        "kube-system",
				ContainersTotal:  3,
				CPURequestsTotal: "500m",
				MemRequestsTotal: "512Mi",
			},
		},
		MetricsAvailable: false,
	}

	result, err := RenderOutput(summary, FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty output with inventory")
		return
	}

	// Verify namespaces appear
	if !strings.Contains(result, "default") {
		t.Error("expected default namespace in output")
	}

	if !strings.Contains(result, "kube-system") {
		t.Error("expected kube-system namespace in output")
	}
}

// TestRenderOutput_EmptyData tests RenderOutput with minimal data
func TestRenderOutput_EmptyData(t *testing.T) {
	summary := &ResourcesSummary{
		Cluster:          &ClusterCapacitySummary{},
		PodDetails:       []PodCapacitySummary{},
		Inventory:        []NamespaceCapacitySummary{},
		MetricsAvailable: false,
	}

	result, err := RenderOutput(summary, FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still produce output even with empty data
	if result == "" {
		t.Error("expected non-empty output even with minimal data")
	}
}

// TestFormatConsistency tests that all formats produce valid output
func TestFormatConsistency(t *testing.T) {
	summary := &ResourcesSummary{
		Cluster: &ClusterCapacitySummary{
			CPUCapacity:    "4",
			MemCapacity:    "8Gi",
			CPURequests:    "2",
			MemRequests:    "4Gi",
			CPUAllocatable: "3.8",
			MemAllocatable: "7.5Gi",
		},
		PodDetails: []PodCapacitySummary{
			{
				Namespace:  "default",
				Pod:        "test-pod",
				CPURequest: "100m",
				MemRequest: "128Mi",
			},
		},
		Inventory:        []NamespaceCapacitySummary{},
		MetricsAvailable: false,
	}

	formats := []OutputFormat{FormatText, FormatJSON, FormatYAML}
	results := make(map[OutputFormat]string)

	// Generate output for all formats
	for _, format := range formats {
		result, err := RenderOutput(summary, format)
		if err != nil {
			t.Fatalf("format %s failed: %v", format, err)
		}
		if result == "" {
			t.Fatalf("format %s produced empty output", format)
		}
		results[format] = result
	}

	// Verify all formats contain key data (though formatted differently)
	for format, result := range results {
		// All formats should contain pod name (though formatted differently)
		if !strings.Contains(strings.ToLower(result), "pod") {
			t.Errorf("format %s missing pod reference", format)
		}

		// All formats should contain capacity info
		if !strings.Contains(strings.ToLower(result), "capacity") &&
			!strings.Contains(strings.ToLower(result), "cpu") {
			t.Errorf("format %s missing capacity info", format)
		}
	}
}
