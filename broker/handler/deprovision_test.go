package handler

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sacloud/open-service-broker-sacloud/service"
	"github.com/stretchr/testify/assert"
)

func TestDeprovisioningHandler(t *testing.T) {

	instanceID := testInstanceID
	serviceID := service.MariaDBServiceID
	planID := service.MariaDBPlan10GID

	target := fmt.Sprintf("/v2/service_instance/%s", instanceID)

	t.Run("Empty service_id", func(t *testing.T) {
		url := fmt.Sprintf("%s?service_id=", target)
		req := httptest.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
		w := httptest.NewRecorder()

		deprovisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateServiceIDRequiredResponse(), w.Body.Bytes())
	})

	t.Run("Empty plan_id", func(t *testing.T) {
		url := fmt.Sprintf("%s?service_id=%s&plan_id=", target, serviceID)
		req := httptest.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
		w := httptest.NewRecorder()

		deprovisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generatePlanIDRequiredResponse(), w.Body.Bytes())
	})

	t.Run("Invalid ServiceID", func(t *testing.T) {
		url := fmt.Sprintf("%s?service_id=%s&plan_id=%s", target, "invalid", planID)
		req := httptest.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
		w := httptest.NewRecorder()

		deprovisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInvalidServiceIDResponse(), w.Body.Bytes())
	})

	t.Run("Invalid PlanID", func(t *testing.T) {
		url := fmt.Sprintf("%s?service_id=%s&plan_id=%s", target, serviceID, "invalid")
		req := httptest.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
		w := httptest.NewRecorder()

		deprovisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInvalidPlanIDResponse(), w.Body.Bytes())
	})
}

func TestDeprovisioning(t *testing.T) {

	instanceID := testInstanceID
	req := httptest.NewRequest(http.MethodDelete, "/", bytes.NewBuffer([]byte{}))

	t.Run("Handler returns error", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceStateErr: errors.New("dummy"),
		}

		deprovisioning(w, req, instanceID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Instance not found", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{}

		deprovisioning(w, req, instanceID, dummyHandler)

		assert.Equal(t, http.StatusGone, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Still migrating", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceState: &dummyInstanceState{
				isMigrating: true,
			},
		}
		deprovisioning(w, req, instanceID, dummyHandler)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateStateMigratingResponse(), w.Body.Bytes())
	})

	t.Run("Still deprovisioning", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceState: &dummyInstanceState{},
		}
		deprovisioning(w, req, instanceID, dummyHandler)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, generateDeprovisionAcceptedResponse(), w.Body.Bytes())
	})

	t.Run("Delete instance is failed", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceState: &dummyInstanceState{
				isUp: true,
			},
			deleteInstanceErr: errors.New("dummy"),
		}

		deprovisioning(w, req, instanceID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Accepted", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceState: &dummyInstanceState{
				isUp: true,
			},
		}

		deprovisioning(w, req, instanceID, dummyHandler)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, generateDeprovisionAcceptedResponse(), w.Body.Bytes())
	})
}
