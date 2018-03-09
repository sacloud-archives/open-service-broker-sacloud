package osb

// Plan represents object of OpenServiceBroker API
type Plan struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Metadata    *Metadata      `json:"metadata,omitempty"`
	Free        bool           `json:"free"`
	Bindable    bool           `json:"bindable"` // nolint
	Schemas     *SchemasObject `json:"schemas,omitempty"`
}
