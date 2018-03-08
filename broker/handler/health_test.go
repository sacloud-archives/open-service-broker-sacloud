package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthzHandler(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handled := healthHandler(w, req)

	assert.True(t, handled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, generateEmptyResponse(), w.Body.Bytes())
}
