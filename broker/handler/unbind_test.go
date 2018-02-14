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

func TestUnboundingHandler(t *testing.T) {

	instanceID := testInstanceID
	bindingID := testInstanceID
	serviceID := service.MariaDBServiceID
	planID := service.MariaDBPlan10GID

	target := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)

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

func TestUnbounding(t *testing.T) {

	instanceID := testInstanceID
	bindingID := testInstanceID
	req := httptest.NewRequest(http.MethodDelete, "/", bytes.NewBuffer([]byte{}))

	t.Run("InstanceState returns error", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceStateErr: errors.New("dummy"),
		}

		unbinding(w, req, instanceID, bindingID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Instance not found", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{}

		unbinding(w, req, instanceID, bindingID, dummyHandler)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInstanceNotFoundResponse(), w.Body.Bytes())
	})

	t.Run("Invalid instance state", func(t *testing.T) {
		t.Run("failed state", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{
					isUp:     false,
					isFailed: true,
				},
			}

			unbinding(w, req, instanceID, bindingID, dummyHandler)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})
		t.Run("downed state", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{
					isUp:     false,
					isFailed: false,
				},
			}

			unbinding(w, req, instanceID, bindingID, dummyHandler)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})

	})

	t.Run("Delete binding is failed", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceState: &dummyInstanceState{
				isUp: true,
			},
			deleteBindingErr: errors.New("dummy"),
		}

		unbinding(w, req, instanceID, bindingID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Done", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceState: &dummyInstanceState{
				isUp: true,
			},
		}

		unbinding(w, req, instanceID, bindingID, dummyHandler)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})
}
