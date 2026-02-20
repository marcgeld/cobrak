package resources

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// BuildInventory queries the cluster for pods, limitranges, and resourcequotas
// and returns per-namespace inventories, per-container resources, and policy summaries.
func BuildInventory(ctx context.Context, client kubernetes.Interface, namespace string) (
	[]NamespaceInventory,
	[]ContainerResources,
	[]PolicySummary,
	error,
) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("listing pods: %w", err)
	}

	limitRanges, err := client.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("listing limitranges: %w", err)
	}

	resourceQuotas, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("listing resourcequotas: %w", err)
	}

	var allContainers []ContainerResources
	nsMap := make(map[string]*NamespaceInventory)

	for i := range pods.Items {
		pod := &pods.Items[i]
		ns := pod.Namespace
		if _, ok := nsMap[ns]; !ok {
			nsMap[ns] = &NamespaceInventory{Namespace: ns}
		}

		for _, c := range pod.Spec.InitContainers {
			cr := extractContainerResources(ns, pod.Name, c, true)
			allContainers = append(allContainers, cr)
			addToNamespaceInventory(nsMap[ns], cr)
		}

		for _, c := range pod.Spec.Containers {
			cr := extractContainerResources(ns, pod.Name, c, false)
			allContainers = append(allContainers, cr)
			addToNamespaceInventory(nsMap[ns], cr)
		}
	}

	policyMap := make(map[string]*PolicySummary)

	for i := range limitRanges.Items {
		lr := &limitRanges.Items[i]
		ns := lr.Namespace
		if _, ok := policyMap[ns]; !ok {
			policyMap[ns] = &PolicySummary{Namespace: ns}
		}
		policyMap[ns].LimitRanges = append(policyMap[ns].LimitRanges, summarizeLimitRange(lr))
	}

	for i := range resourceQuotas.Items {
		rq := &resourceQuotas.Items[i]
		ns := rq.Namespace
		if _, ok := policyMap[ns]; !ok {
			policyMap[ns] = &PolicySummary{Namespace: ns}
		}
		policyMap[ns].ResourceQuotas = append(policyMap[ns].ResourceQuotas, summarizeResourceQuota(rq))
	}

	nsKeys := make([]string, 0, len(nsMap))
	for k := range nsMap {
		nsKeys = append(nsKeys, k)
	}
	sort.Strings(nsKeys)

	nsInventories := make([]NamespaceInventory, 0, len(nsKeys))
	for _, ns := range nsKeys {
		nsInventories = append(nsInventories, *nsMap[ns])
	}

	sort.Slice(allContainers, func(i, j int) bool {
		a, b := allContainers[i], allContainers[j]
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.PodName != b.PodName {
			return a.PodName < b.PodName
		}
		if a.ContainerName != b.ContainerName {
			return a.ContainerName < b.ContainerName
		}
		return a.IsInit && !b.IsInit
	})

	policyKeys := make([]string, 0, len(policyMap))
	for k := range policyMap {
		policyKeys = append(policyKeys, k)
	}
	sort.Strings(policyKeys)

	policies := make([]PolicySummary, 0, len(policyKeys))
	for _, ns := range policyKeys {
		policies = append(policies, *policyMap[ns])
	}

	return nsInventories, allContainers, policies, nil
}

func extractContainerResources(ns, podName string, c v1.Container, isInit bool) ContainerResources {
	cr := ContainerResources{
		Namespace:     ns,
		PodName:       podName,
		ContainerName: c.Name,
		IsInit:        isInit,
	}

	if req, ok := c.Resources.Requests[v1.ResourceCPU]; ok {
		cr.CPURequest = req.DeepCopy()
		cr.HasCPURequest = true
	}
	if lim, ok := c.Resources.Limits[v1.ResourceCPU]; ok {
		cr.CPULimit = lim.DeepCopy()
		cr.HasCPULimit = true
	}
	if req, ok := c.Resources.Requests[v1.ResourceMemory]; ok {
		cr.MemRequest = req.DeepCopy()
		cr.HasMemRequest = true
	}
	if lim, ok := c.Resources.Limits[v1.ResourceMemory]; ok {
		cr.MemLimit = lim.DeepCopy()
		cr.HasMemLimit = true
	}

	return cr
}

func addToNamespaceInventory(inv *NamespaceInventory, cr ContainerResources) {
	inv.ContainersTotal++

	if !cr.HasCPURequest && !cr.HasMemRequest {
		inv.ContainersMissingAnyRequests++
	}
	if !cr.HasCPULimit && !cr.HasMemLimit {
		inv.ContainersMissingAnyLimits++
	}

	if cr.HasCPURequest {
		inv.CPURequestsTotal.Add(cr.CPURequest)
	}
	if cr.HasCPULimit {
		inv.CPULimitsTotal.Add(cr.CPULimit)
	}
	if cr.HasMemRequest {
		inv.MemRequestsTotal.Add(cr.MemRequest)
	}
	if cr.HasMemLimit {
		inv.MemLimitsTotal.Add(cr.MemLimit)
	}
}

func summarizeLimitRange(lr *v1.LimitRange) LimitRangeSummary {
	s := LimitRangeSummary{Name: lr.Name}
	for _, item := range lr.Spec.Limits {
		is := LimitRangeItemSummary{Type: string(item.Type)}
		if v, ok := item.Default[v1.ResourceCPU]; ok {
			is.DefaultCPU = v.String()
		}
		if v, ok := item.Default[v1.ResourceMemory]; ok {
			is.DefaultMemory = v.String()
		}
		if v, ok := item.Max[v1.ResourceCPU]; ok {
			is.MaxCPU = v.String()
		}
		if v, ok := item.Max[v1.ResourceMemory]; ok {
			is.MaxMemory = v.String()
		}
		if v, ok := item.Min[v1.ResourceCPU]; ok {
			is.MinCPU = v.String()
		}
		if v, ok := item.Min[v1.ResourceMemory]; ok {
			is.MinMemory = v.String()
		}
		s.Items = append(s.Items, is)
	}
	return s
}

func summarizeResourceQuota(rq *v1.ResourceQuota) ResourceQuotaSummary {
	s := ResourceQuotaSummary{
		Name: rq.Name,
		Hard: make(map[v1.ResourceName]resource.Quantity),
		Used: make(map[v1.ResourceName]resource.Quantity),
	}
	for k, v := range rq.Status.Hard {
		s.Hard[k] = v.DeepCopy()
	}
	for k, v := range rq.Status.Used {
		s.Used[k] = v.DeepCopy()
	}
	return s
}
