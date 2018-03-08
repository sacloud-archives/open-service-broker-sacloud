package osb

type ServiceInstanceLastOperation struct {
	State string `json:"state"`

	Description string `json:"description,omitempty"`
}
