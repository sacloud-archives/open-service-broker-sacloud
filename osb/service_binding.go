package osb

type ServiceBinding struct {
	Credentials interface{} `json:"credentials,omitempty"`

	SyslogDrainURL string `json:"syslog_drain_url,omitempty"`

	RouteServiceURL string `json:"route_service_url,omitempty"`

	VolumeMounts *[]ServiceBindingVolumeMount `json:"volume_mounts,omitempty"`
}

type BindingAlreadyExistsError struct{}

func (e *BindingAlreadyExistsError) Error() string {
	return "Binding already exists"
}
