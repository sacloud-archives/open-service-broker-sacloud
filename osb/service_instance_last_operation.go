package osb

// ServiceInstanceLastOperation represents object of OpenServiceBroker API
type ServiceInstanceLastOperation struct {
	State       string `json:"state"`
	Description string `json:"description,omitempty"`
}
