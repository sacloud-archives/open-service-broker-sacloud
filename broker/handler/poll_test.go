package handler

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/service"
	"github.com/stretchr/testify/assert"
)

func TestPollHandler(t *testing.T) {

	instanceID := testInstanceID

	target := fmt.Sprintf("/v2/service_instance/%s/last_operation", instanceID)

	t.Run("Empty operation", func(t *testing.T) {
		body := bytes.NewReader([]byte{})
		url := target + "?operation="

		req := httptest.NewRequest(http.MethodGet, url, body)
		w := httptest.NewRecorder()

		pollHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateOperationRequiredResponse(), w.Body.Bytes())
	})

	t.Run("Invalid operation", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{}`))
		url := fmt.Sprintf("%s?operation=%s", target, "foobar")

		req := httptest.NewRequest(http.MethodGet, url, body)
		w := httptest.NewRecorder()

		pollHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateOperationInvalidResponse(), w.Body.Bytes())
	})

	t.Run("Invalid service_id", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{}`))
		url := fmt.Sprintf("%s?operation=%s&service_id=%s",
			target, operations.Provisioning, "foobar")

		req := httptest.NewRequest(http.MethodGet, url, body)
		w := httptest.NewRecorder()

		pollHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInvalidServiceIDResponse(), w.Body.Bytes())
	})

	t.Run("Invalid plan_id", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{}`))
		url := fmt.Sprintf("%s?operation=%s&service_id=%s&plan_id=%s",
			target, operations.Provisioning, service.MariaDBServiceID, "foobar")

		req := httptest.NewRequest(http.MethodGet, url, body)
		w := httptest.NewRecorder()

		pollHandler(w, req)

		// should return 400(bad request)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, generateInvalidPlanIDResponse(), w.Body.Bytes())
	})

}

func TestPolling(t *testing.T) {

	instanceID := testInstanceID
	serviceID := service.MariaDBServiceID
	planID := service.MariaDBPlan10GID
	targetFormat := fmt.Sprintf("/v2/service_instance/%s/last_operation?operation=%%s&service_id=%s&plan_id=%s",
		instanceID, serviceID, planID)

	t.Run("Instance instanceState fetch error", func(t *testing.T) {
		url := fmt.Sprintf(targetFormat, operations.Provisioning)
		req := httptest.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))

		w := httptest.NewRecorder()

		dummyHandler = &dummyServiceHandler{
			instanceStateErr: errors.New("dummy"),
		}

		polling(w, req, operations.Provisioning, instanceID, dummyHandler)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
	})

	t.Run("Provision/Update", func(t *testing.T) {
		url := fmt.Sprintf(targetFormat, operations.Provisioning)
		req := httptest.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))

		t.Run("not found", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{} // instanceState is nil

			polling(w, req, operations.Provisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateOperationFailedResponse(), w.Body.Bytes())
		})

		t.Run("succeeded", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{isUp: true},
			}

			polling(w, req, operations.Provisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateOperationSucceededResponse(), w.Body.Bytes())
		})

		t.Run("failed", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{isFailed: true},
			}

			polling(w, req, operations.Provisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateOperationFailedResponse(), w.Body.Bytes())
		})

		t.Run("in progress", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{},
			}

			polling(w, req, operations.Provisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateOperationInProgressResponse(), w.Body.Bytes())
		})

	})

	t.Run("Deprovision", func(t *testing.T) {
		url := fmt.Sprintf(targetFormat, operations.Deprovisioning)
		req := httptest.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))

		t.Run("not found", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{} // instanceState is nil

			polling(w, req, operations.Deprovisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusGone, w.Code)
			assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
		})

		t.Run("still available", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{isUp: true},
			}

			polling(w, req, operations.Deprovisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateOperationInProgressResponse(), w.Body.Bytes())
		})

		t.Run("failed", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{isFailed: true},
			}

			polling(w, req, operations.Deprovisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateOperationInProgressResponse(), w.Body.Bytes())
		})

		t.Run("in progress", func(t *testing.T) {
			w := httptest.NewRecorder()

			dummyHandler = &dummyServiceHandler{
				instanceState: &dummyInstanceState{},
			}

			polling(w, req, operations.Deprovisioning, instanceID, dummyHandler)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, generateOperationInProgressResponse(), w.Body.Bytes())
		})

	})
}
