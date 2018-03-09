package osb

// Error represents object of OpenServiceBroker API
//
// See [Service Broker Errors](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#service-broker-errors) for more details.
type Error struct {
	Error       string `json:"error,omitempty"`
	Description string `json:"description,omitempty"`
}
