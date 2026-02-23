package output

import (
	"strings"
	"testing"

	"github.com/marcgeld/cobrak/pkg/resources"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestRenderNamespaceInventoryTable_Empty(t *testing.T) {
	out := RenderNamespaceInventoryTable(nil)
	if !strings.Contains(out, "NAMESPACE") {
		t.Errorf("expected header in output, got: %s", out)
	}
}

func TestRenderNamespaceInventoryTable_WithData(t *testing.T) {
	inv := []resources.NamespaceInventory{
		{
			Namespace:                    "default",
			ContainersTotal:              5,
			ContainersMissingAnyRequests: 2,
			ContainersMissingAnyLimits:   1,
			CPURequestsTotal:             resource.MustParse("500m"),
			CPULimitsTotal:               resource.MustParse("1"),
			MemRequestsTotal:             resource.MustParse("512Mi"),
			MemLimitsTotal:               resource.MustParse("1Gi"),
		},
	}
	out := RenderNamespaceInventoryTable(inv)
	if !strings.Contains(out, "default") {
		t.Errorf("expected 'default' in output, got: %s", out)
	}
	if !strings.Contains(out, "500m") {
		t.Errorf("expected CPU request in output, got: %s", out)
	}
}

func TestRenderMissingResourcesTable_NoMissing(t *testing.T) {
	containers := []resources.ContainerResources{
		{HasCPURequest: true, HasMemRequest: true, HasCPULimit: true, HasMemLimit: true},
	}
	out := RenderMissingResourcesTable(containers, 10)
	if !strings.Contains(out, "No containers") {
		t.Errorf("expected 'No containers' in output, got: %s", out)
	}
}

func TestRenderUsageTable_Empty(t *testing.T) {
	out := RenderUsageTable(nil, 10)
	if !strings.Contains(out, "No usage") {
		t.Errorf("expected 'No usage' in output, got: %s", out)
	}
}

func TestRenderDiffTable_Empty(t *testing.T) {
	out := RenderDiffTable(nil, 10)
	if !strings.Contains(out, "No diff") {
		t.Errorf("expected 'No diff' in output, got: %s", out)
	}
}

func TestRenderPolicySummary_Empty(t *testing.T) {
	out := RenderPolicySummary(nil)
	if !strings.Contains(out, "No policy") {
		t.Errorf("expected 'No policy' in output, got: %s", out)
	}
}
