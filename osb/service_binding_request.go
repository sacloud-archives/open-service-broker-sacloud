package osb

// ServiceBindingRequest represents object of OpenServiceBroker API
type ServiceBindingRequest struct {
	Context      *Context                     `json:"context,omitempty"`
	ServiceID    string                       `json:"service_id"`
	PlanID       string                       `json:"plan_id"`
	AppGUID      string                       `json:"app_guid,omitempty"`
	BindResource *ServiceBindingResouceObject `json:"bind_resource,omitempty"`
	Parameters   interface{}                  `json:"parameters,omitempty"`
}
