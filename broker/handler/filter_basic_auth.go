package handler

import (
	"encoding/base64"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func newFilterBasicAuth(username, password string) handlerFunc {

	if username == "" || password == "" {

		log.Warn(
			`[Init] username or password is empty, disabled BASIC auth filter`)

		return func(w http.ResponseWriter, req *http.Request) bool {
			// noop
			return false
		}
	}

	return func(w http.ResponseWriter, req *http.Request) bool {
		headerValue := req.Header.Get(reqAuthorization)
		if headerValue == "" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return true
		}
		headerValueTokens := strings.SplitN(
			req.Header.Get(reqAuthorization),
			" ",
			2,
		)
		if len(headerValueTokens) != 2 || headerValueTokens[0] != "Basic" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return true
		}
		b64UsernameAndPassword := headerValueTokens[1]
		usernameAndPassword, err := base64.StdEncoding.DecodeString(
			b64UsernameAndPassword,
		)
		if err != nil {
			http.Error(w, "{}", http.StatusUnauthorized)
			return true
		}
		usernameAndPasswordTokens := strings.SplitN(
			string(usernameAndPassword),
			":",
			2,
		)
		if len(usernameAndPasswordTokens) != 2 ||
			usernameAndPasswordTokens[0] != username ||
			usernameAndPasswordTokens[1] != password {
			http.Error(w, "{}", http.StatusUnauthorized)
			return true
		}

		return false
	}
}
