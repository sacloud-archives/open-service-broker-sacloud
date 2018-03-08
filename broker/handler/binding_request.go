package handler

import (
	"encoding/json"
)

// BindingRequest represents a request to binding a service instance
type BindingRequest struct {
	ServiceID  string                 `json:"service_id"`
	PlanID     string                 `json:"plan_id"`
	Parameters map[string]interface{} `json:"parameters"`
}

// NewBindingRequestFromJSON returns a new BindingRequest unmarshaled
// from the provided JSON []byte
func NewBindingRequestFromJSON(
	jsonBytes []byte,
) (*BindingRequest, error) {
	provisioningRequest := &BindingRequest{}
	err := json.Unmarshal(jsonBytes, provisioningRequest)
	if err != nil {
		return nil, err
	}
	return provisioningRequest, nil
}

// ToJSON returns a []byte containing a JSON representation of the binding
// request
func (p *BindingRequest) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// RawParameter returns a []byte containing a JSON representation of Parameters field
func (p *BindingRequest) RawParameter() ([]byte, error) {
	return json.Marshal(p.Parameters)
}
