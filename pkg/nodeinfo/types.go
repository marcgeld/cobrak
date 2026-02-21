package nodeinfo

// NodeInfo contains detailed system information about a node
type NodeInfo struct {
	NodeName           string
	OS                 string
	Kernel             string
	CPU                CPUInfo
	GPU                GPUInfo
	MemoryPressure     MemoryPressure
	FilesystemLatency  FilesystemLatency
	ContainerRuntime   ContainerRuntime
	VirtualizationType string
	Architecture       string
	KubeletVersion     string
}

// CPUInfo contains CPU information
type CPUInfo struct {
	Model    string
	Count    int
	Capacity int64 // in millicores
}

// GPUInfo contains GPU information
type GPUInfo struct {
	Available bool
	GPUs      []GPU
}

// GPU represents a single GPU
type GPU struct {
	Index string
	Model string
}

// MemoryPressure contains memory-related pressure metrics
type MemoryPressure struct {
	Total            int64   // Total memory in bytes
	Available        int64   // Available memory in bytes
	Used             int64   // Used memory in bytes
	UtilizationRatio float64 // 0.0-1.0
	Pressure         string  // LOW, MEDIUM, HIGH
	PageCacheRatio   float64 // Ratio of pagecache to total
}

// FilesystemLatency contains filesystem performance metrics
type FilesystemLatency struct {
	RootFSLatency      int     // milliseconds
	RootFSInodesUsed   float64 // percentage 0-100
	RootFSCapacityUsed float64 // percentage 0-100
}

// ContainerRuntime contains information about the container runtime
type ContainerRuntime struct {
	Name    string // docker, containerd, cri-o, etc.
	Version string
}

// NodeHealthStatus represents overall node health
type NodeHealthStatus struct {
	NodeName  string
	Status    string // HEALTHY, WARNING, CRITICAL
	Issues    []string
	Timestamp int64
}
