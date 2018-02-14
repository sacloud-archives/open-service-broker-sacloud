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

func TestProvisioningHandler(t *testing.T) {

	instanceID := testInstanceID
	serviceID := testInstanceID
	planID := testInstanceID

	target := fmt.Sprintf("/v2/service_instances/%s", instanceID)

	t.Run("Unreadable body", func(t *testing.T) {

		body := &dummyReader{}
		req := httptest.NewRequest(http.MethodPut, target, body)
		w := httptest.NewRecorder()

		provisionHandler(w, req)
		// should return 500
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Empty body", func(t *testing.T) {
		body := bytes.NewReader([]byte{})
		req := httptest.NewRequest(http.MethodPut, target, body)
		w := httptest.NewRecorder()

		provisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateMalformedRequestResponse(), w.Body.Bytes())
	})

	t.Run("Empty JSON", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{}`))
		req := httptest.NewRequest(http.MethodPut, target, body)
		w := httptest.NewRecorder()

		provisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateServiceIDRequiredResponse(), w.Body.Bytes())
	})

	t.Run("Empty plan_id", func(t *testing.T) {
		body := bytes.NewReader([]byte(fmt.Sprintf(`{"service_id": "%s"}`, serviceID)))
		req := httptest.NewRequest(http.MethodPut, target, body)
		w := httptest.NewRecorder()

		provisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generatePlanIDRequiredResponse(), w.Body.Bytes())
	})

	t.Run("Invalid ServiceID", func(t *testing.T) {
		strBody := fmt.Sprintf(`{"service_id":"%s","plan_id":"%s"}`, serviceID, planID)
		body := bytes.NewReader([]byte(strBody))
		req := httptest.NewRequest(http.MethodPut, target, body)
		w := httptest.NewRecorder()

		provisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInvalidServiceIDResponse(), w.Body.Bytes())
	})

	t.Run("Invalid PlanID", func(t *testing.T) {
		// use exists service ID
		strBody := fmt.Sprintf(`{"service_id":"%s","plan_id":"%s"}`, service.MariaDBServiceID, planID)
		body := bytes.NewReader([]byte(strBody))
		req := httptest.NewRequest(http.MethodPut, target, body)
		w := httptest.NewRecorder()

		provisionHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInvalidPlanIDResponse(), w.Body.Bytes())
	})
}

func TestProvisioning(t *testing.T) {

	instanceID := testInstanceID
	serviceID := service.MariaDBServiceID
	planID := service.MariaDBPlan10GID

	parameterJSONFormat := fmt.Sprintf(`{
		"service_id": "%s",
		"plan_id": "%s",
		"parameters": %%s
	}`, serviceID, planID)

	target := fmt.Sprintf("/v2/service_instance/%s", instanceID)
	body := bytes.NewBuffer([]byte(fmt.Sprintf(parameterJSONFormat, `{}`)))
	req := httptest.NewRequest(http.MethodPut, target, body)

	t.Run("Validation failed", func(t *testing.T) {
		w := httptest.NewRecorder()

		expectErr := errors.New("dummy")
		dummyHandler = &dummyServiceHandler{
			validateResult: expectErr,
		}

		provisioning(w, req, instanceID, dummyHandler)

		// should return 400
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateMalformedParameterResponse(expectErr.Error()), w.Body.Bytes())
	})

	t.Run("Handler returns error", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceStateErr: errors.New("dummy"),
		}

		provisioning(w, req, instanceID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Instance not exists", func(t *testing.T) {
		t.Run("create failed", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				createInstanceErr: errors.New("dummy"),
			}
			provisioning(w, req, instanceID, dummyHandler)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})

		t.Run("creation accepted", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{}
			provisioning(w, req, instanceID, dummyHandler)

			assert.Equal(t, http.StatusAccepted, w.Code)
			assert.Equal(t, generateProvisionAcceptedResponse(), w.Body.Bytes())
		})
	})

	t.Run("Instance Exists", func(t *testing.T) {
		t.Run("conflict attrs", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{
					hasDiff: true,
				},
			}

			provisioning(w, req, instanceID, dummyHandler)

			assert.Equal(t, http.StatusConflict, w.Code)
			assert.Equal(t, generateConflictResponse(), w.Body.Bytes())
		})

		t.Run("with Failed instanceState", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{
					isFailed: true,
				},
			}

			provisioning(w, req, instanceID, dummyHandler)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})

		t.Run("with Up instanceState", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{
					isUp: true,
				},
			}

			provisioning(w, req, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})

		t.Run("still provisioning", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{},
			}

			provisioning(w, req, instanceID, dummyHandler)

			assert.Equal(t, http.StatusAccepted, w.Code)
			assert.Equal(t, generateProvisionAcceptedResponse(), w.Body.Bytes())
		})

	})
}
