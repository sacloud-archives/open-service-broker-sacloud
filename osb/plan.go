package osb

type Plan struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Metadata    *Metadata      `json:"metadata,omitempty"`
	Free        bool           `json:"free"`
	Bindable    bool           `json:"bindable"`
	Schemas     *SchemasObject `json:"schemas,omitempty"`
}
