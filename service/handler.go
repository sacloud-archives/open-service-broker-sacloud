package service

import "github.com/sacloud/open-service-broker-sacloud/osb"

// Handler is interface that represents Broker operations
type Handler interface {
	InstanceState(instanceID string) (InstanceState, error)
	BindingState(instanceID, bindingID string) (BindingState, error)

	CreateInstance(instanceID string) error
	UpdateInstance(instanceID string) error
	DeleteInstance(instanceID string) error

	CreateBinding(instanceID, bindingID string) (*osb.ServiceBinding, error)
	DeleteBinding(instanceID, bindingID string) error

	IsValid() (bool, error)
}
