package handler

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/service"
)

func unbindHandler(w http.ResponseWriter, req *http.Request) (handled bool) {

	//collect parameters
	instanceID := mux.Vars(req)[reqInstanceID]
	bindingID := mux.Vars(req)[reqBindingID]

	logFields := log.Fields{
		"instanceID": instanceID,
		"bindingID":  bindingID,
	}
	log.WithFields(logFields).Debug("received unbinding request")

	// collect optional query parameters from URL
	serviceID := req.URL.Query().Get(reqServiceID)
	if serviceID == "" {
		logFields["field"] = "service_id" //nolint
		log.WithFields(logFields).Debug(
			"bad unbinding request: service_id is required",
		)
		writeResponse(w, http.StatusBadRequest, generateServiceIDRequiredResponse())
		return
	}

	planID := req.URL.Query().Get(reqPlanID)
	if planID == "" {
		logFields["field"] = "plan_id" // nolint
		log.WithFields(logFields).Debug(
			"bad unbinding request: plan_id is required",
		)
		writeResponse(w, http.StatusBadRequest, generatePlanIDRequiredResponse())
		return
	}

	svc, ok := service.CurrentCatalog.FindService(serviceID)
	if !ok {
		logFields["serviceID"] = serviceID
		log.WithFields(logFields).Debug(
			"bad unbinding request: invalid serviceID",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidServiceIDResponse())
		return
	}

	_, ok = svc.FindPlan(planID)
	if !ok {
		logFields["serviceID"] = serviceID
		logFields["planID"] = planID
		log.WithFields(logFields).Debug(
			"bad unbinding request: invalid planID for service",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidPlanIDResponse())
		return
	}

	handler := service.Factory(operations.Unbinding, serviceID, planID, []byte{})
	if handler == nil {
		logFields["field"] = "unbound handler"
		log.WithFields(logFields).Warn(
			"bad unbinding request: invalid unbound handler",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	unbinding(w, req, instanceID, bindingID, handler)
	handled = true
	return
}

func unbinding(w http.ResponseWriter, req *http.Request, instanceID, bindingID string, handler service.Handler) {
	logFields := log.Fields{
		"instanceID": instanceID,
		"bindingID":  bindingID,
	}

	instanceState, err := handler.InstanceState(instanceID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"unbinding failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	if instanceState == nil {
		log.WithFields(logFields).Error(
			"unbinding failed: instance not found",
		)
		writeResponse(w, http.StatusBadRequest, generateInstanceNotFoundResponse())
		return
	}

	if !instanceState.IsUp() || instanceState.IsFailed() {
		log.WithFields(logFields).Error(
			"unbinding failed: instance state is invalid",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	err = handler.DeleteBinding(instanceID, bindingID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"unbinding failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}
	log.WithFields(logFields).Info(
		"unbinding complete: Binding deleted",
	)
	writeResponse(w, http.StatusOK, generateEmptyResponse())
}
