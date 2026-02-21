package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// OutputFormat specifies the output format
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatYAML OutputFormat = "yaml"
)

// ParseOutputFormat parses a string to OutputFormat
func ParseOutputFormat(format string) (OutputFormat, error) {
	switch format {
	case "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	default:
		return FormatText, fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", format)
	}
}

// Renderer is an interface for rendering data in different formats
type Renderer interface {
	RenderText() string
	RenderJSON() (string, error)
	RenderYAML() (string, error)
}

// RenderOutput renders data in the specified format
func RenderOutput(data interface{}, format OutputFormat) (string, error) {
	switch format {
	case FormatJSON:
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("JSON marshaling error: %w", err)
		}
		return string(jsonBytes), nil
	case FormatYAML:
		yamlBytes, err := yaml.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("YAML marshaling error: %w", err)
		}
		return string(yamlBytes), nil
	case FormatText:
		// For text format, data should implement Renderer interface
		if renderer, ok := data.(Renderer); ok {
			return renderer.RenderText(), nil
		}
		// Fallback to JSON if text rendering not implemented
		return RenderOutput(data, FormatJSON)
	default:
		return "", fmt.Errorf("unsupported format: %v", format)
	}
}

// MetaOutput provides a unified output structure for different formats
type MetaOutput struct {
	Format OutputFormat
	Data   interface{}
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Render renders the MetaOutput in the specified format
func (m *MetaOutput) Render() (string, error) {
	if m.Format == FormatText {
		if renderer, ok := m.Data.(Renderer); ok {
			return renderer.RenderText(), nil
		}
		// Fallback to JSON text representation
		return RenderOutput(m.Data, FormatJSON)
	}
	return RenderOutput(m.Data, m.Format)
}

// ResourcesSummary represents the complete resources output structure
type ResourcesSummary struct {
	ClusterCapacity    *ClusterCapacitySummary `json:"cluster_capacity" yaml:"clusterCapacity"`
	PodDetails         []PodDetail             `json:"pod_details" yaml:"podDetails"`
	Totals             *ResourceTotals         `json:"totals" yaml:"totals"`
	NamespaceInventory []NamespaceSummary      `json:"namespace_inventory" yaml:"namespaceInventory"`
	MetricsAvailable   bool                    `json:"metrics_available" yaml:"metricsAvailable"`
}

// ClusterCapacitySummary represents cluster capacity data
type ClusterCapacitySummary struct {
	CPUCapacity    string `json:"cpu_capacity" yaml:"cpuCapacity"`
	CPUAllocatable string `json:"cpu_allocatable" yaml:"cpuAllocatable"`
	CPURequests    string `json:"cpu_requests" yaml:"cpuRequests"`
	CPULimits      string `json:"cpu_limits" yaml:"cpuLimits"`
	MemCapacity    string `json:"mem_capacity" yaml:"memCapacity"`
	MemAllocatable string `json:"mem_allocatable" yaml:"memAllocatable"`
	MemRequests    string `json:"mem_requests" yaml:"memRequests"`
	MemLimits      string `json:"mem_limits" yaml:"memLimits"`
}

// PodDetail represents a single pod's resource details
type PodDetail struct {
	Namespace  string `json:"namespace" yaml:"namespace"`
	Pod        string `json:"pod" yaml:"pod"`
	CPURequest string `json:"cpu_request" yaml:"cpuRequest"`
	CPULimit   string `json:"cpu_limit" yaml:"cpuLimit"`
	MemRequest string `json:"mem_request" yaml:"memRequest"`
	MemLimit   string `json:"mem_limit" yaml:"memLimit"`
}

// ResourceTotals represents total resources
type ResourceTotals struct {
	TotalCPURequests string `json:"total_cpu_requests" yaml:"totalCpuRequests"`
	TotalCPULimits   string `json:"total_cpu_limits" yaml:"totalCpuLimits"`
	TotalMemRequests string `json:"total_mem_requests" yaml:"totalMemRequests"`
	TotalMemLimits   string `json:"total_mem_limits" yaml:"totalMemLimits"`
}

// NamespaceSummary represents namespace resource summary
type NamespaceSummary struct {
	Namespace       string `json:"namespace" yaml:"namespace"`
	ContainersTotal int    `json:"containers_total" yaml:"containersTotal"`
	MissingRequests int    `json:"missing_requests" yaml:"missingRequests"`
	MissingLimits   int    `json:"missing_limits" yaml:"missingLimits"`
	CPURequests     string `json:"cpu_requests" yaml:"cpuRequests"`
	CPULimits       string `json:"cpu_limits" yaml:"cpuLimits"`
	MemRequests     string `json:"mem_requests" yaml:"memRequests"`
	MemLimits       string `json:"mem_limits" yaml:"memLimits"`
}

// PressureSummary represents cluster pressure data
type PressureSummary struct {
	ClusterPressure    string         `json:"cluster_pressure" yaml:"clusterPressure"`
	CPUUtilization     float64        `json:"cpu_utilization" yaml:"cpuUtilization"`
	MemUtilization     float64        `json:"mem_utilization" yaml:"memUtilization"`
	NodePressures      []NodePressure `json:"node_pressures" yaml:"nodePressures"`
	NamespacePressures []NSPressure   `json:"namespace_pressures" yaml:"namespacePressures"`
}

// NodePressure represents pressure for a single node
type NodePressure struct {
	NodeName       string  `json:"node_name" yaml:"nodeName"`
	CPUPressure    string  `json:"cpu_pressure" yaml:"cpuPressure"`
	CPUUtilization float64 `json:"cpu_utilization" yaml:"cpuUtilization"`
	MemPressure    string  `json:"mem_pressure" yaml:"memPressure"`
	MemUtilization float64 `json:"mem_utilization" yaml:"memUtilization"`
}

// NSPressure represents namespace pressure
type NSPressure struct {
	Namespace  string  `json:"namespace" yaml:"namespace"`
	CPUPercent float64 `json:"cpu_percent" yaml:"cpuPercent"`
	MemPercent float64 `json:"mem_percent" yaml:"memPercent"`
}

// NodeInfoSummary represents node information in structured format
type NodeInfoSummary struct {
	NodeName         string      `json:"node_name" yaml:"nodeName"`
	OS               string      `json:"os" yaml:"os"`
	Kernel           string      `json:"kernel" yaml:"kernel"`
	Architecture     string      `json:"architecture" yaml:"architecture"`
	KubeletVersion   string      `json:"kubelet_version" yaml:"kubeletVersion"`
	CPU              CPUData     `json:"cpu" yaml:"cpu"`
	GPU              GPUData     `json:"gpu" yaml:"gpu"`
	Memory           MemoryData  `json:"memory" yaml:"memory"`
	Filesystem       FSData      `json:"filesystem" yaml:"filesystem"`
	ContainerRuntime RuntimeData `json:"container_runtime" yaml:"containerRuntime"`
	Virtualization   string      `json:"virtualization" yaml:"virtualization"`
}

// CPUData represents CPU information
type CPUData struct {
	Model    string `json:"model" yaml:"model"`
	Cores    int    `json:"cores" yaml:"cores"`
	Capacity int64  `json:"capacity_m" yaml:"capacityM"`
}

// GPUData represents GPU information
type GPUData struct {
	Available bool     `json:"available" yaml:"available"`
	Count     int      `json:"count" yaml:"count"`
	Models    []string `json:"models" yaml:"models"`
}

// MemoryData represents memory information
type MemoryData struct {
	Total       string  `json:"total" yaml:"total"`
	Utilization float64 `json:"utilization" yaml:"utilization"`
	Pressure    string  `json:"pressure" yaml:"pressure"`
}

// FSData represents filesystem information
type FSData struct {
	RootFSLatency      int     `json:"root_fs_latency_ms" yaml:"rootFsLatencyMs"`
	RootFSInodesUsed   float64 `json:"root_fs_inodes_used_percent" yaml:"rootFsInodesUsedPercent"`
	RootFSCapacityUsed float64 `json:"root_fs_capacity_used_percent" yaml:"rootFsCapacityUsedPercent"`
}

// RuntimeData represents container runtime information
type RuntimeData struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
}

// NodeHealthSummary represents node health in structured format
type NodeHealthSummary struct {
	NodeName string   `json:"node_name" yaml:"nodeName"`
	Status   string   `json:"status" yaml:"status"`
	Issues   []string `json:"issues" yaml:"issues"`
}
