package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"errors"

	"github.com/sacloud/open-service-broker-sacloud/osb"
	"github.com/sacloud/open-service-broker-sacloud/service"
	"github.com/stretchr/testify/assert"
)

func TestBindingHandler(t *testing.T) {

	instanceID := testInstanceID
	bindingID := testInstanceID
	serviceID := testInstanceID
	planID := testInstanceID

	target := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)

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

func TestBinding(t *testing.T) {

	instanceID := testInstanceID
	bindingID := testInstanceID
	serviceID := service.MariaDBServiceID
	planID := service.MariaDBPlan10GID

	parameterJSONFormat := fmt.Sprintf(`{
		"service_id": "%s",
		"plan_id": "%s",
		"parameters": %%s
	}`, serviceID, planID)

	target := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)
	body := bytes.NewBuffer([]byte(fmt.Sprintf(parameterJSONFormat, `{}`)))
	req := httptest.NewRequest(http.MethodPut, target, body)

	t.Run("Validation failed", func(t *testing.T) {
		w := httptest.NewRecorder()

		expectErr := errors.New("dummy")
		dummyHandler = &dummyServiceHandler{
			validateResult: expectErr,
		}

		binding(w, req, instanceID, bindingID, dummyHandler)

		// should return 400
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateMalformedParameterResponse(expectErr.Error()), w.Body.Bytes())
	})

	t.Run("Invalid instance state", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceStateErr: errors.New("dummy"),
		}

		binding(w, req, instanceID, bindingID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Instance not exists", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{}

		binding(w, req, instanceID, bindingID, dummyHandler)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInstanceNotFoundResponse(), w.Body.Bytes())
	})

	t.Run("Handler returns error", func(t *testing.T) {
		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceState:   &dummyInstanceState{},
			bindingStateErr: errors.New("dummy"),
		}

		binding(w, req, instanceID, bindingID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Binding not exists", func(t *testing.T) {
		t.Run("create failed", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState:    &dummyInstanceState{},
				createBindingErr: errors.New("dummy"),
			}
			binding(w, req, instanceID, bindingID, dummyHandler)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})

		t.Run("result is nil", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{},
			}
			binding(w, req, instanceID, bindingID, dummyHandler)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})

		t.Run("binding created", func(t *testing.T) {
			w := httptest.NewRecorder()

			result := &osb.ServiceBinding{
				Credentials: map[string]interface{}{
					"foo": "bar",
				},
			}
			resultResponse, _ := json.Marshal(result)

			dummyHandler = &dummyServiceHandler{
				instanceState:       &dummyInstanceState{},
				createBindingResult: result,
			}
			binding(w, req, instanceID, bindingID, dummyHandler)

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.Equal(t, resultResponse, w.Body.Bytes())
		})
	})
	t.Run("Binding exists", func(t *testing.T) {
		t.Run("binding has diff", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{},
				bindingState: &dummyBindingState{
					hasDiff: true,
				},
			}
			binding(w, req, instanceID, bindingID, dummyHandler)

			assert.Equal(t, http.StatusConflict, w.Code)
			assert.Equal(t, generateBindingConflictResponse(), w.Body.Bytes())
		})
		t.Run("binding already created", func(t *testing.T) {
			w := httptest.NewRecorder()
			result := &osb.ServiceBinding{
				Credentials: map[string]interface{}{
					"foo": "bar",
				},
			}
			resultResponse, _ := json.Marshal(result)

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{},
				bindingState: &dummyBindingState{
					binding: result,
				},
			}
			binding(w, req, instanceID, bindingID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, resultResponse, w.Body.Bytes())
		})
	})
}
