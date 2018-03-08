package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

type handlerDefine struct {
	path     string
	method   string
	handlers []handlerFunc
}

type handlerFunc func(w http.ResponseWriter, req *http.Request) (handled bool)

var handlers = []handlerDefine{
	{
		path:     "/v2/catalog",
		method:   http.MethodGet,
		handlers: []handlerFunc{filterAPIVersion, catalogHandler},
	},
	{
		path:     "/v2/service_instances/{instance_id}",
		method:   http.MethodPut,
		handlers: []handlerFunc{filterAPIVersion, filterAcceptsIncomplete, provisionHandler},
	},
	{
		path:     "/v2/service_instances/{instance_id}",
		method:   http.MethodPatch,
		handlers: []handlerFunc{filterAPIVersion}, // TODO not implements
	},
	{
		path:     "/v2/service_instances/{instance_id}/last_operation",
		method:   http.MethodGet,
		handlers: []handlerFunc{filterAPIVersion, pollHandler},
	},
	{
		path:     "/v2/service_instances/{instance_id}/service_bindings/{binding_id}",
		method:   http.MethodPut,
		handlers: []handlerFunc{filterAPIVersion, bindingHandler},
	},
	{
		path:     "/v2/service_instances/{instance_id}/service_bindings/{binding_id}",
		method:   http.MethodDelete,
		handlers: []handlerFunc{filterAPIVersion, unbindHandler},
	},
	{
		path:     "/v2/service_instances/{instance_id}",
		method:   http.MethodDelete,
		handlers: []handlerFunc{filterAPIVersion, filterAcceptsIncomplete, deprovisionHandler},
	},
}

// Router returns Handler for handling broker-api-server
func Router(username, password string) http.Handler {
	router := mux.NewRouter()
	router.StrictSlash(true)

	authFilter := newFilterBasicAuth(username, password)

	for _, def := range handlers {
		h := append([]handlerFunc{authFilter}, def.handlers...)

		router.HandleFunc(
			def.path, handlerChain(h...),
		).Methods(def.method)
	}

	// add health check(without filters)
	router.HandleFunc("/healthz", handlerChain(healthHandler)).Methods(http.MethodGet)

	return router
}

func handlerChain(handlers ...handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		applyHandlers(w, req, handlers...)
	}
}

func applyHandlers(w http.ResponseWriter, req *http.Request, handlers ...handlerFunc) (handled bool) {
	for _, f := range handlers {
		handled = f(w, req)
		if handled {
			break
		}
	}
	return
}
