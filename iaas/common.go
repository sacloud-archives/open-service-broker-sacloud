package iaas

import (
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

var error404 = api.NewError( // nolint
	404, &sacloud.ResultErrorValue{
		ErrorCode:    "not_found",
		ErrorMessage: "The target can not be found. Object or state that can not be available, there is an error in the ID or path.",
		IsFatal:      true,
		Serial:       "",
		Status:       "404 Not Found",
	})
