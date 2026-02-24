package output

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/marcgeld/cobrak/pkg/capacity"
	"github.com/marcgeld/cobrak/pkg/resources"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Pressure is an alias for capacity.ClusterPressure
type Pressure = capacity.ClusterPressure

// RenderNamespaceInventoryTable formats a table of namespace inventories.
func RenderNamespaceInventoryTable(inventories []resources.NamespaceInventory) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tCONTAINERS\tMISSING REQUESTS\tMISSING LIMITS\tCPU REQ\tCPU LIM\tMEM REQ\tMEM LIM")
	for _, ns := range inventories {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\t%s\t%s\t%s\n",
			ns.Namespace,
			ns.ContainersTotal,
			ns.ContainersMissingAnyRequests,
			ns.ContainersMissingAnyLimits,
			ns.CPURequestsTotal.String(),
			ns.CPULimitsTotal.String(),
			ns.MemRequestsTotal.String(),
			ns.MemLimitsTotal.String(),
		)
	}
	_ = w.Flush() //nolint:errcheck
	return strings.TrimRight(buf.String(), "\n")
}

// RenderMissingResourcesTable formats a table of containers missing requests/limits.
func RenderMissingResourcesTable(containers []resources.ContainerResources, top int) string {
	var missing []resources.ContainerResources
	for _, c := range containers {
		if !c.HasCPURequest || !c.HasMemRequest || !c.HasCPULimit || !c.HasMemLimit {
			missing = append(missing, c)
		}
	}

	if len(missing) == 0 {
		return "No containers with missing requests/limits."
	}

	if top > 0 && len(missing) > top {
		missing = missing[:top]
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tPOD\tCONTAINER\tINIT\tCPU REQ\tCPU LIM\tMEM REQ\tMEM LIM")
	for _, c := range missing {
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%v\t%v\t%v\t%v\n",
			c.Namespace, c.PodName, c.ContainerName, c.IsInit,
			c.HasCPURequest, c.HasCPULimit, c.HasMemRequest, c.HasMemLimit,
		)
	}
	_ = w.Flush() //nolint:errcheck
	return strings.TrimRight(buf.String(), "\n")
}

// RenderPolicySummary formats LimitRange and ResourceQuota summaries.
func RenderPolicySummary(policies []resources.PolicySummary) string {
	if len(policies) == 0 {
		return "No policy objects (LimitRange, ResourceQuota) found."
	}

	var sb strings.Builder
	for _, ps := range policies {
		fmt.Fprintf(&sb, "Namespace: %s\n", ps.Namespace)

		if len(ps.LimitRanges) > 0 {
			sb.WriteString("  LimitRanges:\n")
			for _, lr := range ps.LimitRanges {
				fmt.Fprintf(&sb, "    - %s\n", lr.Name)
				for _, item := range lr.Items {
					fmt.Fprintf(&sb, "      Type: %s", item.Type)
					if item.DefaultCPU != "" {
						fmt.Fprintf(&sb, "  DefaultCPU: %s", item.DefaultCPU)
					}
					if item.DefaultMemory != "" {
						fmt.Fprintf(&sb, "  DefaultMemory: %s", item.DefaultMemory)
					}
					if item.MaxCPU != "" {
						fmt.Fprintf(&sb, "  MaxCPU: %s", item.MaxCPU)
					}
					if item.MaxMemory != "" {
						fmt.Fprintf(&sb, "  MaxMemory: %s", item.MaxMemory)
					}
					sb.WriteString("\n")
				}
			}
		}

		if len(ps.ResourceQuotas) > 0 {
			sb.WriteString("  ResourceQuotas:\n")
			for _, rq := range ps.ResourceQuotas {
				fmt.Fprintf(&sb, "    - %s\n", rq.Name)
				var hardKeys []string
				for k := range rq.Hard {
					hardKeys = append(hardKeys, string(k))
				}
				sort.Strings(hardKeys)
				for _, k := range hardKeys {
					hard := rq.Hard[v1.ResourceName(k)]
					used := rq.Used[v1.ResourceName(k)]
					fmt.Fprintf(&sb, "      %s: used=%s hard=%s\n", k, used.String(), hard.String())
				}
			}
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// RenderPressureSimple renders a simple pressure summary with colors.
func RenderPressureSimple(pressure *Pressure) string {
	var sb strings.Builder

	// Cluster overall pressure with color
	pressureText := colorizePressureLevel(string(pressure.Overall), pressure.Overall)
	fmt.Fprintf(&sb, "Cluster Pressure: %s\n", pressureText)

	// Node pressures
	for _, np := range pressure.NodePressures {
		if np.CPUPressure != "LOW" {
			cpuPressure := colorizePressureLevel(string(np.CPUPressure), np.CPUPressure)
			nodeName := Header(np.NodeName)
			fmt.Fprintf(&sb, "Node %s: CPU %s (%.0f%%)\n", nodeName, cpuPressure, np.CPUUtilization)
		}
		if np.MemPressure != "LOW" {
			memPressure := colorizePressureLevel(string(np.MemPressure), np.MemPressure)
			nodeName := Header(np.NodeName)
			fmt.Fprintf(&sb, "Node %s: Memory %s (%.0f%%)\n", nodeName, memPressure, np.MemUtilization)
		}
	}

	// Namespace pressures - only show if >= 80%
	for _, nsp := range pressure.NamespacePressures {
		if nsp.CPUPercent >= 80 {
			nsName := Info(nsp.Namespace)
			fmt.Fprintf(&sb, "Namespace %s: CPU %.0f%% requested\n", nsName, nsp.CPUPercent)
		}
		if nsp.MemPercent >= 80 {
			nsName := Info(nsp.Namespace)
			fmt.Fprintf(&sb, "Namespace %s: Memory %.0f%% requested\n", nsName, nsp.MemPercent)
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// colorizePressureLevel applies appropriate color to pressure level text
func colorizePressureLevel(text string, level capacity.PressureLevel) string {
	switch level {
	case capacity.PressureLow:
		return PressureLowColor(text)
	case capacity.PressureMedium:
		return PressureMediumColor(text)
	case capacity.PressureHigh:
		return PressureHighColor(text)
	case capacity.PressureSaturated:
		return PressureSaturatedColor(text)
	default:
		return text
	}
}

// RenderUsageTable formats a table of container usages.
func RenderUsageTable(usages []resources.ContainerUsage, top int) string {
	if len(usages) == 0 {
		return "No usage data available."
	}

	if top > 0 && len(usages) > top {
		usages = usages[:top]
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tPOD\tCONTAINER\tCPU\tMEMORY")
	for _, u := range usages {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			u.Namespace, u.PodName, u.ContainerName,
			u.CPUUsage.String(), u.MemUsage.String(),
		)
	}
	_ = w.Flush() //nolint:errcheck
	return strings.TrimRight(buf.String(), "\n")
}

// RenderDiffTable formats a table of container diffs.
func RenderDiffTable(diffs []resources.ContainerDiff, top int) string {
	if len(diffs) == 0 {
		return "No diff data available."
	}

	if top > 0 && len(diffs) > top {
		diffs = diffs[:top]
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tPOD\tCONTAINER\tCPU USAGE\tCPU REQ\tCPU RATIO\tMEM USAGE\tMEM REQ\tMEM RATIO")
	for _, d := range diffs {
		cpuRatio := "-"
		if d.HasCPURequest {
			cpuRatio = fmt.Sprintf("%.2f", d.CPUUsageToRequest)
		}
		memRatio := "-"
		if d.HasMemRequest {
			memRatio = fmt.Sprintf("%.2f", d.MemUsageToRequest)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			d.Namespace, d.PodName, d.ContainerName,
			d.CPUUsage.String(), d.CPURequest.String(), cpuRatio,
			d.MemUsage.String(), d.MemRequest.String(), memRatio,
		)
	}
	_ = w.Flush() //nolint:errcheck
	return strings.TrimRight(buf.String(), "\n")
}

// RenderPodResourceSummary formats a table of pod resource summaries (requests/limits).
func RenderPodResourceSummary(pods []resources.PodResourceSummary, top int) string {
	if len(pods) == 0 {
		return "No pods found."
	}

	// Limit to top N if top > 0
	if top > 0 && len(pods) > top {
		pods = pods[:top]
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tPOD\tCPU REQUEST\tCPU LIMIT\tMEM REQUEST\tMEM LIMIT")
	for _, pod := range pods {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			pod.Namespace, pod.PodName,
			pod.CPURequest.String(), pod.CPULimit.String(),
			pod.MemRequest.String(), pod.MemLimit.String(),
		)
	}
	_ = w.Flush() //nolint:errcheck
	return strings.TrimRight(buf.String(), "\n")
}

// RenderPodResourceSummaryWithUsage formats a table of pod resource summaries including usage data.
func RenderPodResourceSummaryWithUsage(pods []resources.PodResourceSummary, top int) string {
	if len(pods) == 0 {
		return "No pods found."
	}

	// Limit to top N if top > 0
	if top > 0 && len(pods) > top {
		pods = pods[:top]
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tPOD\tCPU USAGE\tCPU REQUEST\tCPU LIMIT\tMEM USAGE\tMEM REQUEST\tMEM LIMIT")
	for _, pod := range pods {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			pod.Namespace, pod.PodName,
			pod.CPUUsage.String(), pod.CPURequest.String(), pod.CPULimit.String(),
			pod.MemUsage.String(), pod.MemRequest.String(), pod.MemLimit.String(),
		)
	}
	_ = w.Flush() //nolint:errcheck
	return strings.TrimRight(buf.String(), "\n")
}

// RenderPodResourceSummaryTotals renders totals for pod resource summaries.
func RenderPodResourceSummaryTotals(pods []resources.PodResourceSummary) string {
	if len(pods) == 0 {
		return ""
	}

	var totalCPUUsage, totalCPURequest, totalCPULimit *resource.Quantity
	var totalMemUsage, totalMemRequest, totalMemLimit *resource.Quantity

	// Initialize with zero values
	totalCPUUsage = resource.NewQuantity(0, resource.DecimalSI)
	totalCPURequest = resource.NewQuantity(0, resource.DecimalSI)
	totalCPULimit = resource.NewQuantity(0, resource.DecimalSI)
	totalMemUsage = resource.NewQuantity(0, resource.BinarySI)
	totalMemRequest = resource.NewQuantity(0, resource.BinarySI)
	totalMemLimit = resource.NewQuantity(0, resource.BinarySI)

	// Sum all values
	for _, pod := range pods {
		totalCPUUsage.Add(pod.CPUUsage)
		totalCPURequest.Add(pod.CPURequest)
		totalCPULimit.Add(pod.CPULimit)
		totalMemUsage.Add(pod.MemUsage)
		totalMemRequest.Add(pod.MemRequest)
		totalMemLimit.Add(pod.MemLimit)
	}

	var sb strings.Builder
	sb.WriteString("=== TOTALS ===\n")
	fmt.Fprintf(&sb, "Total CPU Usage:       %s\n", totalCPUUsage.String())
	fmt.Fprintf(&sb, "Total CPU Requests:    %s\n", totalCPURequest.String())
	fmt.Fprintf(&sb, "Total CPU Limits:      %s\n", totalCPULimit.String())
	fmt.Fprintf(&sb, "\nTotal Memory Usage:    %s\n", totalMemUsage.String())
	fmt.Fprintf(&sb, "Total Memory Requests: %s\n", totalMemRequest.String())
	fmt.Fprintf(&sb, "Total Memory Limits:   %s\n", totalMemLimit.String())

	return strings.TrimRight(sb.String(), "\n")
}
