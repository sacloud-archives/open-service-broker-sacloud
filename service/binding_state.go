package service

import "github.com/sacloud/open-service-broker-sacloud/osb"

// BindingState is represents current binding state
type BindingState interface {
	HasDiff() bool
	Binding() *osb.ServiceBinding
}
