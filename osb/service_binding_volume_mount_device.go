package osb

// ServiceBindingVolumeMountDevice represents object of OpenServiceBroker API
type ServiceBindingVolumeMountDevice struct {
	VolumeID    string      `json:"volume_id"`
	MountConfig interface{} `json:"mount_config,omitempty"`
}
