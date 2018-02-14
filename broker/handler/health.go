package handler

import (
	"net/http"
)

func healthHandler(w http.ResponseWriter, req *http.Request) bool {
	writeResponse(w, http.StatusOK, generateEmptyResponse())
	return true
}
