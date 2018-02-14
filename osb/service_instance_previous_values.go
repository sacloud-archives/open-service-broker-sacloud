package osb

type ServiceInstancePreviousValues struct {
	ServiceID string `json:"service_id,omitempty"`

	PlanID string `json:"plan_id,omitempty"`

	OrganizationID string `json:"organization_id,omitempty"`

	SpaceID string `json:"space_id,omitempty"`
}
