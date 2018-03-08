package handler

import (
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/service"
)

func provisionHandler(w http.ResponseWriter, req *http.Request) (handled bool) {

	//collect parameters
	instanceID := mux.Vars(req)[reqInstanceID]

	logFields := log.Fields{
		"instanceID": instanceID,
	}
	log.WithFields(logFields).Debug("received provisioning request")

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logFields["error"] = err
		log.WithFields(logFields).Error(
			"pre-provisioning error: error reading request body",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}
	defer req.Body.Close() // nolint

	provisioningRequest, err := NewProvisioningRequestFromJSON(bodyBytes)
	if err != nil {
		logFields["error"] = err
		log.WithFields(logFields).Debug(
			"bad provisioning request: error unmarshaling request body",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	serviceID := provisioningRequest.ServiceID
	if serviceID == "" {
		logFields["field"] = "service_id" // nolint
		log.WithFields(logFields).Debug(
			"bad provisioning request: required request body field is missing",
		)
		writeResponse(w, http.StatusBadRequest, generateServiceIDRequiredResponse())
		return
	}

	planID := provisioningRequest.PlanID
	if planID == "" {
		logFields["field"] = "plan_id" // nolint
		log.WithFields(logFields).Debug(
			"bad provisioning request: required request body field is missing",
		)
		writeResponse(w, http.StatusBadRequest, generatePlanIDRequiredResponse())
		return
	}

	svc, ok := service.CurrentCatalog.FindService(serviceID)
	if !ok {
		logFields["serviceID"] = serviceID
		log.WithFields(logFields).Debug(
			"bad provisioning request: invalid serviceID",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidServiceIDResponse())
		return
	}

	_, ok = svc.FindPlan(planID)
	if !ok {
		logFields["serviceID"] = serviceID
		logFields["planID"] = planID
		log.WithFields(logFields).Debug(
			"bad provisioning request: invalid planID for service",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidPlanIDResponse())
		return
	}

	rawParameter, err := provisioningRequest.RawParameter()
	if err != nil {
		logFields["field"] = "parameters"
		log.WithFields(logFields).Debug(
			"bad provisioning request: error marshaling request body(parameters field)",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	handler := service.Factory(operations.Provisioning, serviceID, planID, rawParameter)
	if handler == nil {
		logFields["field"] = "provisioner"
		log.WithFields(logFields).Warn(
			"bad provisioning request: invalid provisioner",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	provisioning(w, req, instanceID, handler)
	handled = true
	return
}

func provisioning(w http.ResponseWriter, req *http.Request, instanceID string, handler service.Handler) {
	logFields := log.Fields{
		"instanceID": instanceID,
	}

	_, err := handler.IsValid()
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Debug(
			`bad provisioning request: invalid JSON parameter`)
		writeResponse(w, http.StatusBadRequest, generateMalformedParameterResponse(err.Error()))
		return
	}

	state, err := handler.InstanceState(instanceID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"provisioning failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	if state == nil {
		// create new Database Appliance resource
		err = handler.CreateInstance(instanceID)
		if err != nil {
			logFields["error"] = err
			log.WithFields(logFields).Error(
				"provisioning error: error creating SakuraCloud resource",
			)
			writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
			return
		}

		// Now, creation succeeded
		log.WithFields(logFields).Info(
			"provisioning succeeded: Instance creation accepted",
		)
		writeResponse(w, http.StatusAccepted, generateProvisionAcceptedResponse())
		return
	}

	if state.HasDiff() {
		log.WithFields(logFields).Error(
			"provisioning error: Instance is already exists with different attrs",
		)
		writeResponse(w, http.StatusConflict, generateConflictResponse())
		return
	}

	switch {
	case state.IsFailed(): // failed on sakura cloud
		log.WithFields(logFields).Error(
			"provisioning error: Instance.State is failed",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
	case state.IsUp(): // fully provisioned
		log.WithFields(logFields).Info(
			"provisioning succeeded: Instance already fully provisioned",
		)
		writeResponse(w, http.StatusOK, generateEmptyResponse())
	default: // still provisioning
		log.WithFields(logFields).Info(
			"provisioning succeeded: Instance already started provisioning",
		)
		writeResponse(w, http.StatusAccepted, generateProvisionAcceptedResponse())
	}
}
