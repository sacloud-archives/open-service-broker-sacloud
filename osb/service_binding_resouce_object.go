package osb

// ServiceBindingResouceObject represents object of OpenServiceBroker API
type ServiceBindingResouceObject struct {
	AppGUID string `json:"app_guid,omitempty"`
	Route   string `json:"route,omitempty"`
}
