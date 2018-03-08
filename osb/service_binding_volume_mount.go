package osb

type ServiceBindingVolumeMount struct {
	Driver string `json:"driver"`

	ContainerDir string `json:"container_dir"`

	Mode string `json:"mode"`

	DeviceType string `json:"device_type"`

	Device *ServiceBindingVolumeMountDevice `json:"device"`
}
