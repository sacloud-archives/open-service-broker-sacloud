package service

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/sacloud/open-service-broker-sacloud/iaas"
)

// Factory is factory-method to return Handler according to arguments
var Factory func(operation, serviceID, planID string, rawParameter []byte) Handler

var sacloudAPI iaas.Client

func init() {
	Factory = factory
}

func factory(operation, serviceID, planID string, rawParameter []byte) Handler {
	if serviceID == "" || planID == "" {
		return nil
	}

	switch serviceID {
	case MariaDBServiceID:
		return newMariaDBServiceHandler(operation, serviceID, planID, rawParameter)
	case PostgreSQLServiceID:
		return newPostgreSQLServiceHandler(operation, serviceID, planID, rawParameter)
	default:
		return nil
	}
}

// Initialize makes handlers available
func Initialize(client iaas.Client) error {
	sacloudAPI = client

	// check auth-status
	_, err := sacloudAPI.AuthStatus()
	if err != nil {
		log.Error("service initialize error: Sacloud API authentication failed")
		return err
	}

	b, err := json.Marshal(CurrentCatalog)
	if err != nil {
		panic(err)
	}
	CurrentCatalogData = b

	// TODO Add more service initialization

	return nil
}
