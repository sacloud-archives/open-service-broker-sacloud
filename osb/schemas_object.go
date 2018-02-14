package osb

type SchemasObject struct {
	ServiceInstance *ServiceInstanceSchemaObject `json:"service_instance,omitempty"`

	ServiceBinding *ServiceBindingSchemaObject `json:"service_binding,omitempty"`
}
