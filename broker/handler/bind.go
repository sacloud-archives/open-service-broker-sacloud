package handler

import (
	"io/ioutil"
	"net/http"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/osb"
	"github.com/sacloud/open-service-broker-sacloud/service"
)

func bindingHandler(w http.ResponseWriter, req *http.Request) (handled bool) {

	//collect parameters
	instanceID := mux.Vars(req)[reqInstanceID]
	bindingID := mux.Vars(req)[reqBindingID]

	logFields := log.Fields{
		"instanceID": instanceID,
		"bindingID":  bindingID,
	}
	log.WithFields(logFields).Debug("received binding request")

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logFields["error"] = err
		log.WithFields(logFields).Error(
			"pre-binding error: error reading request body",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}
	defer req.Body.Close() // nolint

	bindingRequest, err := NewProvisioningRequestFromJSON(bodyBytes)
	if err != nil {
		logFields["error"] = err
		log.WithFields(logFields).Debug(
			"bad binding request: error unmarshaling request body",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	serviceID := bindingRequest.ServiceID
	if serviceID == "" {
		logFields["field"] = "service_id" // nolint
		log.WithFields(logFields).Debug(
			"bad binding request: required request body field is missing",
		)
		writeResponse(w, http.StatusBadRequest, generateServiceIDRequiredResponse())
		return
	}

	planID := bindingRequest.PlanID
	if planID == "" {
		logFields["field"] = "plan_id" // nolint
		log.WithFields(logFields).Debug(
			"bad binding request: required request body field is missing",
		)
		writeResponse(w, http.StatusBadRequest, generatePlanIDRequiredResponse())
		return
	}

	svc, ok := service.CurrentCatalog.FindService(serviceID)
	if !ok {
		logFields["serviceID"] = serviceID
		log.WithFields(logFields).Debug(
			"bad binding request: invalid serviceID",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidServiceIDResponse())
		return
	}

	_, ok = svc.FindPlan(planID)
	if !ok {
		logFields["serviceID"] = serviceID
		logFields["planID"] = planID
		log.WithFields(logFields).Debug(
			"bad binding request: invalid planID for service",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidPlanIDResponse())
		return
	}

	rawParameter, err := bindingRequest.RawParameter()
	if err != nil {
		logFields["field"] = "parameters"
		log.WithFields(logFields).Debug(
			"bad binding request: error marshaling request body(parameters field)",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	handler := service.Factory(operations.Binding, serviceID, planID, rawParameter)
	if handler == nil {
		logFields["field"] = "binding handler"
		log.WithFields(logFields).Warn(
			"bad binding request: invalid binding handler",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	binding(w, req, instanceID, bindingID, handler)
	handled = true
	return
}

func binding(w http.ResponseWriter, req *http.Request, instanceID, bindingID string, handler service.Handler) {
	logFields := log.Fields{
		"instanceID": instanceID,
	}

	_, err := handler.IsValid()
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Debug(
			`bad binding request: invalid JSON parameter`)
		writeResponse(w, http.StatusBadRequest, generateMalformedParameterResponse(err.Error()))
		return
	}

	instanceState, err := handler.InstanceState(instanceID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"binding failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	if instanceState == nil {
		log.WithFields(logFields).Error(
			`bad binding request: Instance not found`)
		writeResponse(w, http.StatusBadRequest, generateInstanceNotFoundResponse())
		return
	}

	bindingState, err := handler.BindingState(instanceID, bindingID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"binding failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	if bindingState == nil {
		var result *osb.ServiceBinding
		result, err = handler.CreateBinding(instanceID, bindingID)
		if err != nil {
			logFields["error"] = err
			log.WithFields(logFields).Error(
				"binding error: error creating SakuraCloud resource binding",
			)
			writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
			return
		}

		if result == nil {
			log.WithFields(logFields).Error(
				"binding error: binding result is empty",
			)
			writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
			return
		}

		response, err := json.Marshal(result)
		if err != nil {
			logFields["error"] = err
			log.WithFields(logFields).Error(
				"binding error: error marshaling binding-result to JSON",
			)
			writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
			return
		}

		// Now, creation succeeded
		log.WithFields(logFields).Info(
			"binding succeeded: Instance binding created",
		)
		writeResponse(w, http.StatusCreated, response)
		return

	}

	if bindingState.HasDiff() {
		log.WithFields(logFields).Error(
			"binding error: binding already exists but conflicted with requested parameters",
		)
		writeResponse(w, http.StatusConflict, generateBindingConflictResponse())
		return
	}

	response, err := json.Marshal(bindingState.Binding())
	if err != nil {
		logFields["error"] = err
		log.WithFields(logFields).Error(
			"binding error: error marshaling binding-result to JSON",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	log.WithFields(logFields).Info(
		"binding succeeded: binding already exists",
	)
	writeResponse(w, http.StatusOK, response)
}
