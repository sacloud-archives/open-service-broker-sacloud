package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/sacloud/open-service-broker-sacloud/osb"
	"github.com/sacloud/open-service-broker-sacloud/service"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	ret := m.Run()
	os.Exit(ret)
}

const fooValue = "bar"

var (
	testInstanceID   = "5C7CDA6B-5BBC-4600-9CCA-28F6DF36D943"
	testArbitraryMap = map[string]interface{}{
		"foo": "bar",
	}
	testArbitraryMapJSON = []byte(fmt.Sprintf(`{"foo":"%s"}`, fooValue))
)

var (
	dummyHandler = &dummyServiceHandler{}
)

type dummyServiceHandler struct {
	instanceState       service.InstanceState
	bindingState        service.BindingState
	instanceStateErr    error
	bindingStateErr     error
	createInstanceErr   error
	updateInstanceErr   error
	deleteInstanceErr   error
	createBindingErr    error
	deleteBindingErr    error
	createBindingResult *osb.ServiceBinding
	validateResult      error
}

func (s *dummyServiceHandler) InstanceState(instanceID string) (service.InstanceState, error) {
	if s.instanceState == nil {
		return nil, s.instanceStateErr
	}
	return s.instanceState, s.instanceStateErr
}

func (s *dummyServiceHandler) BindingState(instanceID, bindingID string) (service.BindingState, error) {
	if s.bindingState == nil {
		return nil, s.bindingStateErr
	}
	return s.bindingState, s.bindingStateErr
}

func (s *dummyServiceHandler) CreateInstance(instanceID string) error {
	return s.createInstanceErr
}

func (s *dummyServiceHandler) UpdateInstance(instanceID string) error {
	return s.updateInstanceErr
}

func (s *dummyServiceHandler) DeleteInstance(instanceID string) error {
	return s.deleteInstanceErr
}

func (s *dummyServiceHandler) CreateBinding(instanceID, bindingID string) (*osb.ServiceBinding, error) {
	if s.createBindingResult == nil {
		return nil, s.createBindingErr
	}
	return s.createBindingResult, s.createBindingErr
}

func (s *dummyServiceHandler) DeleteBinding(instanceID, bindingID string) error {
	return s.deleteBindingErr
}

func (s *dummyServiceHandler) IsValid() (bool, error) {
	return s.validateResult == nil, s.validateResult
}

type dummyInstanceState struct {
	isFailed    bool
	isUp        bool
	isMigrating bool
	hasDiff     bool
}

func (s *dummyInstanceState) IsFailed() bool {
	return s.isFailed
}

func (s *dummyInstanceState) IsUp() bool {
	return s.isUp
}

func (s *dummyInstanceState) IsMigrating() bool {
	return s.isMigrating
}

func (s *dummyInstanceState) HasDiff() bool {
	return s.hasDiff
}

type dummyBindingState struct {
	hasDiff bool
	binding *osb.ServiceBinding
}

func (s *dummyBindingState) HasDiff() bool {
	return s.hasDiff
}

func (s *dummyBindingState) Binding() *osb.ServiceBinding {
	return s.binding
}

type dummyReader struct{}

func (r *dummyReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("dummy")
}
