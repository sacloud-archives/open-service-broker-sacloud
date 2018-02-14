package osb

type ServiceInstanceProvisionRequest struct {
	ServiceID string `json:"service_id"`

	PlanID string `json:"plan_id"`

	Context *Context `json:"context,omitempty"`

	OrganizationGUID string `json:"organization_guid"`

	SpaceGUID string `json:"space_guid"`

	Parameters interface{} `json:"parameters,omitempty"`
}
