package handler

import (
	"net/http"
	"strings"
)

func filterAPIVersion(w http.ResponseWriter, req *http.Request) bool {

	apiVersion := req.Header.Get(reqBrokerAPIVersion)
	if apiVersion == "" {
		sendError(w, responseMissingAPIVersion)
		return true
	}
	// Allow Broker API 2.xx only
	if !strings.HasPrefix(apiVersion, "2.") {
		sendError(w, responseAPIVersionIncorrect)
		return true
	}

	return false
}

var (
	responseMissingAPIVersion = []byte(
		`{ "error": "MissingAPIVersion", "description": "The request did not ` +
			`include the ` + reqBrokerAPIVersion + ` header"}`,
	)
	responseAPIVersionIncorrect = []byte(
		`{ "error": "APIVersionIncorrect", "description": "` + reqBrokerAPIVersion +
			` header includes an incompatible version"}`,
	)
)
