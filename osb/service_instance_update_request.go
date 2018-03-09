package osb

// ServiceInstanceUpdateRequest represents object of OpenServiceBroker API
type ServiceInstanceUpdateRequest struct {
	Context        *Context                       `json:"context,omitempty"`
	ServiceID      string                         `json:"service_id"`
	PlanID         string                         `json:"plan_id,omitempty"`
	Parameters     interface{}                    `json:"parameters,omitempty"`
	PreviousValues *ServiceInstancePreviousValues `json:"previous_values,omitempty"`
}
