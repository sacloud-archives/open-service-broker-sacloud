package handler

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/service"
)

func deprovisionHandler(w http.ResponseWriter, req *http.Request) (handled bool) {

	//collect parameters
	instanceID := mux.Vars(req)[reqInstanceID]

	logFields := log.Fields{
		"instanceID": instanceID,
	}
	log.WithFields(logFields).Debug("received deprovisioning request")

	// collect optional query parameters from URL
	serviceID := req.URL.Query().Get(reqServiceID)
	if serviceID == "" {
		logFields["field"] = "service_id" // nolint
		log.WithFields(logFields).Debug(
			"bad deprovisioning request: service_id is required",
		)
		writeResponse(w, http.StatusBadRequest, generateServiceIDRequiredResponse())
		return
	}

	planID := req.URL.Query().Get(reqPlanID)
	if planID == "" {
		logFields["field"] = "plan_id" //nolint
		log.WithFields(logFields).Debug(
			"bad deprovisioning request: plan_id is required",
		)
		writeResponse(w, http.StatusBadRequest, generatePlanIDRequiredResponse())
		return
	}

	svc, ok := service.CurrentCatalog.FindService(serviceID)
	if !ok {
		logFields["serviceID"] = serviceID
		log.WithFields(logFields).Debug(
			"bad deprovisioning request: invalid serviceID",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidServiceIDResponse())
		return
	}

	_, ok = svc.FindPlan(planID)
	if !ok {
		logFields["serviceID"] = serviceID
		logFields["planID"] = planID
		log.WithFields(logFields).Debug(
			"bad deprovisioning request: invalid planID for service",
		)
		writeResponse(w, http.StatusBadRequest, generateInvalidPlanIDResponse())
		return
	}

	handler := service.Factory(operations.Deprovisioning, serviceID, planID, []byte{})
	if handler == nil {
		logFields["field"] = "deprovisioner"
		log.WithFields(logFields).Warn(
			"bad deprovisioning request: invalid deprovisioner",
		)
		writeResponse(w, http.StatusBadRequest, generateMalformedRequestResponse())
		return
	}

	deprovisioning(w, req, instanceID, handler)
	handled = true
	return
}

func deprovisioning(w http.ResponseWriter, req *http.Request, instanceID string, handler service.Handler) {
	logFields := log.Fields{
		"instanceID": instanceID,
	}

	state, err := handler.InstanceState(instanceID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"deprovisioning failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}

	if state == nil {
		log.WithFields(logFields).Info(
			"deprovisioning succeeded: instance is gone",
		)
		writeResponse(w, http.StatusGone, generateEmptyResponse())
		return
	}

	if state.IsMigrating() {
		log.WithFields(logFields).Warn(
			"deprovisioning unprocessable: instance is migrating, please try again after",
		)
		writeResponse(w, http.StatusBadRequest, generateStateMigratingResponse())
		return
	}

	if !(state.IsUp() || state.IsFailed()) {
		log.WithFields(logFields).Info(
			"deprovisioning succeeded: Instance already started deprovisioning",
		)
		writeResponse(w, http.StatusAccepted, generateDeprovisionAcceptedResponse())
		return
	}

	err = handler.DeleteInstance(instanceID)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			"deprovisioning failed: service handler returned error",
		)
		writeResponse(w, http.StatusInternalServerError, generateEmptyResponse())
		return
	}
	log.WithFields(logFields).Info(
		"deprovisioning accepted: Instance deletion accepted",
	)
	writeResponse(w, http.StatusAccepted, generateDeprovisionAcceptedResponse())
}
