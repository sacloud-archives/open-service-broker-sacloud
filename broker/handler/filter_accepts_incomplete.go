package handler

import (
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func filterAcceptsIncomplete(w http.ResponseWriter, req *http.Request) bool {

	//collect parameters
	instanceID := mux.Vars(req)[reqInstanceID]
	logFields := log.Fields{
		"instanceID": instanceID,
	}

	// This broker provisions everything asynchronously. If a client doesn't
	// explicitly indicate that they will accept an incomplete readResult, the
	// spec says to respond with a 422
	acceptsIncompleteStr := req.URL.Query().Get(reqAcceptsImcomplete)
	if acceptsIncompleteStr == "" {
		logFields["parameter"] = "accepts_incomplete=true" // nolint: goconst
		log.WithFields(logFields).Debug(
			"bad provisioning request: request is missing required query parameter",
		)
		writeResponse(
			w,
			http.StatusUnprocessableEntity,
			generateAsyncRequiredResponse(),
		)
		return true
	}

	acceptsIncomplete, err := strconv.ParseBool(acceptsIncompleteStr)
	if err != nil || !acceptsIncomplete {
		logFields["accepts_incomplete"] = acceptsIncompleteStr
		log.WithFields(logFields).Debug(
			`bad provisioning request: query parameter has invalid value; only ` +
				`"true" is accepted`,
		)
		writeResponse(
			w,
			http.StatusUnprocessableEntity,
			generateAsyncRequiredResponse(),
		)
		return true
	}
	return false
}
