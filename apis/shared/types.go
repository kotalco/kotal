package shared

// Resources is node compute and storage resources
// +k8s:deepcopy-gen=true
type Resources struct {
	// CPU is cpu cores the node requires
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPU string `json:"cpu,omitempty"`
	// CPULimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPULimit string `json:"cpuLimit,omitempty"`
	// Memory is memmory requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Memory string `json:"memory,omitempty"`
	// MemoryLimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// Storage is disk space storage requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Storage string `json:"storage,omitempty"`
	// StorageClass is the volume storage class
	StorageClass *string `json:"storageClass,omitempty"`
}
