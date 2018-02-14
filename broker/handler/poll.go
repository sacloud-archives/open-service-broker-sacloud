package handler

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/service"
)

func pollHandler(w http.ResponseWriter, req *http.Request) (handled bool) {

	instanceID := mux.Vars(req)[reqInstanceID]

	logFields := log.Fields{
		"instanceID": instanceID,
	}

	log.WithFields(logFields).Debug("received polling request")

	operation := req.URL.Query().Get("operation")
	if operation == "" {
		logFields["parameter"] = "operation"
		log.WithFields(logFields).Debug(
			"bad polling request: request is missing required query parameter",
		)
		writeResponse(w, http.StatusBadRequest, generateOperationRequiredResponse())
		return
	}
	if operation != operations.Provisioning &&
		operation != operations.Deprovisioning &&
		operation != operations.Updating {
		logFields["operation"] = operation
		log.WithFields(logFields).Debug(
			fmt.Sprintf(
				`bad polling request: query parameter has invalid value; only "%s",`+
					` %s, and "%s" are accepted`,
				operations.Provisioning,
				operations.Deprovisioning,
				operations.Updating,
			),
		)
		writeResponse(w, http.StatusBadRequest, generateOperationInvalidResponse())
		return
	}

	logFields["operation"] = operation

	// collect optional query parameters from URL
	serviceID := req.URL.Query().Get(reqServiceID)
	planID := ""

	if serviceID != "" {
		svc, ok := service.CurrentCatalog.FindService(serviceID)
		if !ok {
			logFields["serviceID"] = serviceID
			log.WithFields(logFields).Debug(
				"bad polling request: invalid serviceID",
			)
			writeResponse(w, http.StatusBadRequest, generateInvalidServiceIDResponse())
			return
		}

		planID = req.URL.Query().Get(reqPlanID)
		if planID != "" {
			_, ok = svc.FindPlan(planID)
			if !ok {
				logFields["serviceID"] = serviceID
				logFields["planID"] = planID
				log.WithFields(logFields).Debug(
					"bad polling request: invalid planID for service",
				)
				writeResponse(w, http.StatusBadRequest, generateInvalidPlanIDResponse())
				return
			}
		}
	}

	if serviceID == "" || planID == "" {

		switch operation {
		case operations.Provisioning:
			log.WithFields(logFields).Info(
				"bad polling request: service_id and plan_id are empty",
			)
			writeResponse(w, http.StatusBadRequest, generateServiceIDRequiredResponse())
			return
		case operations.Deprovisioning:
			// can't find instance, so we respond 'done'
			log.WithFields(logFields).Info(
				"polling succeeded: instance is gone(service_id and plan_id are empty)",
			)
			writeResponse(w, http.StatusGone, generateEmptyResponse())
			return
		}
	}

	handler := service.Factory(operation, serviceID, planID, []byte{})
	if handler == nil {
		logFields["field"] = "provisioner"
		log.WithFields(logFields).Warn(
			"bad provisioning request: invalid provisioner",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	polling(w, req, operation, instanceID, handler)
	handled = true
	return

}

func polling(w http.ResponseWriter, req *http.Request, operation string, instanceID string, handler service.Handler) {
	logFields := log.Fields{
		"instanceID": instanceID,
		"operation":  operation,
	}

	state, err := handler.InstanceState(instanceID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"polling failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	if state == nil {
		if operation == operations.Deprovisioning {
			log.WithFields(logFields).Info(
				"polling succeeded: instance is gone",
			)
			writeResponse(w, http.StatusGone, generateEmptyResponse())
			return
		}

		log.WithFields(logFields).Info(
			"polling failed: instance not found",
		)
		writeResponse(w, http.StatusOK, generateOperationFailedResponse())
		return

	}

	if operation == operations.Provisioning {
		if state.IsFailed() {
			log.WithFields(logFields).Info(
				"polling failed: instance not found",
			)
			writeResponse(w, http.StatusOK, generateOperationFailedResponse())
			return
		}

		if state.IsUp() {
			log.WithFields(logFields).Info(
				"polling succeeded: instance fully provisioned",
			)
			writeResponse(w, http.StatusOK, generateOperationSucceededResponse())
			return
		}
	}

	log.WithFields(logFields).Info(
		"polling in progress",
	)
	writeResponse(w, http.StatusOK, generateOperationInProgressResponse())
}
