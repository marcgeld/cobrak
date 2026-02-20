package resources

import (
	"sort"
)

// BuildDiff joins inventory and usage data to compute per-container diffs.
func BuildDiff(inventory []ContainerResources, usage []ContainerUsage) []ContainerDiff {
	type key struct{ ns, pod, container string }
	usageMap := make(map[key]ContainerUsage, len(usage))
	for _, u := range usage {
		usageMap[key{u.Namespace, u.PodName, u.ContainerName}] = u
	}

	diffs := make([]ContainerDiff, 0, len(inventory))
	for _, cr := range inventory {
		k := key{cr.Namespace, cr.PodName, cr.ContainerName}
		u := usageMap[k]

		diff := ContainerDiff{
			Namespace:     cr.Namespace,
			PodName:       cr.PodName,
			ContainerName: cr.ContainerName,

			CPUUsage:      u.CPUUsage.DeepCopy(),
			CPURequest:    cr.CPURequest.DeepCopy(),
			CPULimit:      cr.CPULimit.DeepCopy(),
			HasCPURequest: cr.HasCPURequest,
			HasCPULimit:   cr.HasCPULimit,

			MemUsage:      u.MemUsage.DeepCopy(),
			MemRequest:    cr.MemRequest.DeepCopy(),
			MemLimit:      cr.MemLimit.DeepCopy(),
			HasMemRequest: cr.HasMemRequest,
			HasMemLimit:   cr.HasMemLimit,
		}

		if cr.HasCPURequest && !cr.CPURequest.IsZero() {
			usageMilli := float64(u.CPUUsage.MilliValue())
			requestMilli := float64(cr.CPURequest.MilliValue())
			diff.CPUUsageToRequest = usageMilli / requestMilli
		}

		if cr.HasMemRequest && !cr.MemRequest.IsZero() {
			usageBytes := float64(u.MemUsage.Value())
			requestBytes := float64(cr.MemRequest.Value())
			diff.MemUsageToRequest = usageBytes / requestBytes
		}

		diffs = append(diffs, diff)
	}

	sort.Slice(diffs, func(i, j int) bool {
		a, b := diffs[i], diffs[j]
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.PodName != b.PodName {
			return a.PodName < b.PodName
		}
		return a.ContainerName < b.ContainerName
	})

	return diffs
}
