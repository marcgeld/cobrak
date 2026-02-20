package resources

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerResources holds the requests/limits for a single container.
type ContainerResources struct {
	Namespace     string
	PodName       string
	ContainerName string
	IsInit        bool

	CPURequest resource.Quantity
	CPULimit   resource.Quantity
	MemRequest resource.Quantity
	MemLimit   resource.Quantity

	HasCPURequest bool
	HasCPULimit   bool
	HasMemRequest bool
	HasMemLimit   bool
}

// NamespaceInventory aggregates resource coverage for a namespace.
type NamespaceInventory struct {
	Namespace string

	ContainersTotal              int
	ContainersMissingAnyRequests int
	ContainersMissingAnyLimits   int

	CPURequestsTotal resource.Quantity
	CPULimitsTotal   resource.Quantity
	MemRequestsTotal resource.Quantity
	MemLimitsTotal   resource.Quantity
}

// PolicySummary holds LimitRange and ResourceQuota summaries for a namespace.
type PolicySummary struct {
	Namespace      string
	LimitRanges    []LimitRangeSummary
	ResourceQuotas []ResourceQuotaSummary
}

// LimitRangeSummary is a compact summary of a LimitRange.
type LimitRangeSummary struct {
	Name  string
	Items []LimitRangeItemSummary
}

// LimitRangeItemSummary is one item within a LimitRange.
type LimitRangeItemSummary struct {
	Type          string
	DefaultCPU    string
	DefaultMemory string
	MaxCPU        string
	MaxMemory     string
	MinCPU        string
	MinMemory     string
}

// ResourceQuotaSummary is a compact summary of a ResourceQuota.
type ResourceQuotaSummary struct {
	Name string
	Hard map[v1.ResourceName]resource.Quantity
	Used map[v1.ResourceName]resource.Quantity
}

// ContainerUsage holds actual observed CPU/memory usage for a container.
type ContainerUsage struct {
	Namespace     string
	PodName       string
	ContainerName string
	CPUUsage      resource.Quantity
	MemUsage      resource.Quantity
}

// ContainerDiff compares usage with requests/limits for a container.
type ContainerDiff struct {
	Namespace     string
	PodName       string
	ContainerName string

	CPUUsage      resource.Quantity
	CPURequest    resource.Quantity
	CPULimit      resource.Quantity
	HasCPURequest bool
	HasCPULimit   bool

	MemUsage      resource.Quantity
	MemRequest    resource.Quantity
	MemLimit      resource.Quantity
	HasMemRequest bool
	HasMemLimit   bool

	// Derived signals (ratios: usage / request)
	CPUUsageToRequest float64
	MemUsageToRequest float64
}
