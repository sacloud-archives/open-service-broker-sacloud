package osb

// ServiceInstanceAsyncProvision represents object of OpenServiceBroker API
type ServiceInstanceAsyncProvision struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}
