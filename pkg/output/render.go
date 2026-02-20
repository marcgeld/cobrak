package output

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	v1 "k8s.io/api/core/v1"
	"github.com/marcgeld/cobrak/pkg/resources"
)

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
	w.Flush()
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
	w.Flush()
	return strings.TrimRight(buf.String(), "\n")
}

// RenderPolicySummary formats LimitRange and ResourceQuota summaries.
func RenderPolicySummary(policies []resources.PolicySummary) string {
	if len(policies) == 0 {
		return "No policy objects (LimitRange, ResourceQuota) found."
	}

	var sb strings.Builder
	for _, ps := range policies {
		sb.WriteString(fmt.Sprintf("Namespace: %s\n", ps.Namespace))

		if len(ps.LimitRanges) > 0 {
			sb.WriteString("  LimitRanges:\n")
			for _, lr := range ps.LimitRanges {
				sb.WriteString(fmt.Sprintf("    - %s\n", lr.Name))
				for _, item := range lr.Items {
					sb.WriteString(fmt.Sprintf("      Type: %s", item.Type))
					if item.DefaultCPU != "" {
						sb.WriteString(fmt.Sprintf("  DefaultCPU: %s", item.DefaultCPU))
					}
					if item.DefaultMemory != "" {
						sb.WriteString(fmt.Sprintf("  DefaultMemory: %s", item.DefaultMemory))
					}
					if item.MaxCPU != "" {
						sb.WriteString(fmt.Sprintf("  MaxCPU: %s", item.MaxCPU))
					}
					if item.MaxMemory != "" {
						sb.WriteString(fmt.Sprintf("  MaxMemory: %s", item.MaxMemory))
					}
					sb.WriteString("\n")
				}
			}
		}

		if len(ps.ResourceQuotas) > 0 {
			sb.WriteString("  ResourceQuotas:\n")
			for _, rq := range ps.ResourceQuotas {
				sb.WriteString(fmt.Sprintf("    - %s\n", rq.Name))
				var hardKeys []string
				for k := range rq.Hard {
					hardKeys = append(hardKeys, string(k))
				}
				sort.Strings(hardKeys)
				for _, k := range hardKeys {
					hard := rq.Hard[v1.ResourceName(k)]
					used := rq.Used[v1.ResourceName(k)]
					sb.WriteString(fmt.Sprintf("      %s: used=%s hard=%s\n", k, used.String(), hard.String()))
				}
			}
		}
	}
	return strings.TrimRight(sb.String(), "\n")
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
	w.Flush()
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
	w.Flush()
	return strings.TrimRight(buf.String(), "\n")
}
