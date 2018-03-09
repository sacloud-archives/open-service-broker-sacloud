package osb

// SchemaParameters represents object of OpenServiceBroker API
type SchemaParameters struct {
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}
