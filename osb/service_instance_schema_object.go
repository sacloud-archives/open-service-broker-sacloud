package osb

type ServiceInstanceSchemaObject struct {
	Create *SchemaParameters `json:"create,omitempty"`

	Update *SchemaParameters `json:"update,omitempty"`
}
