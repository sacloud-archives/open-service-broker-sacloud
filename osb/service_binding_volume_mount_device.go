package osb

type ServiceBindingVolumeMountDevice struct {
	VolumeID string `json:"volume_id"`

	MountConfig interface{} `json:"mount_config,omitempty"`
}
