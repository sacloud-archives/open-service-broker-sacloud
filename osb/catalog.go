package osb

type Catalog struct {
	Services []*Service `json:"services"`
}

// FindService returns a service with the specified ID
func (c *Catalog) FindService(id string) (*Service, bool) {
	for _, s := range c.Services {
		if s.ID == id {
			return s, true
		}
	}
	return nil, false
}
