package osb

// ServiceInstanceSchemaObject represents object of OpenServiceBroker API
type ServiceInstanceSchemaObject struct {
	Create *SchemaParameters `json:"create,omitempty"`
	Update *SchemaParameters `json:"update,omitempty"`
}
