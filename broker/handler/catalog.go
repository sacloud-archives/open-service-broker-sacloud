package handler

import (
	"net/http"

	"github.com/sacloud/open-service-broker-sacloud/service"
)

func catalogHandler(w http.ResponseWriter, req *http.Request) bool {
	writeResponse(w, http.StatusOK, service.CurrentCatalogData)
	return true
}
