package osb

// Service represents object of OpenServiceBroker API
type Service struct {
	Name            string           `json:"name"`
	ID              string           `json:"id"`
	Description     string           `json:"description"`
	Tags            []string         `json:"tags"`
	Requires        []string         `json:"requires"`
	Bindable        bool             `json:"bindable"` // nolint
	Metadata        *Metadata        `json:"metadata,omitempty"`
	DashboardClient *DashboardClient `json:"dashboard_client,omitempty"`
	PlanUpdateable  bool             `json:"plan_updateable,omitempty"` // nolint
	Plans           []*Plan          `json:"plans"`
}

// FindPlan returns a plan with the specified ID
func (s *Service) FindPlan(id string) (*Plan, bool) {
	for _, p := range s.Plans {
		if p.ID == id {
			return p, true
		}
	}
	return nil, false
}
