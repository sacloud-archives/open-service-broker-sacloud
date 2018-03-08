package handler

import (
	"fmt"

	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
)

var responseAsyncRequired = []byte(
	`{ "error": "AsyncRequired", "description": "This service plan requires ` +
		`client support for asynchronous service operations." }`,
)

func generateAsyncRequiredResponse() []byte {
	return responseAsyncRequired
}

var responseServiceIDRequired = []byte(
	`{ "error": "ServiceIdRequired", "description": "service_id is a required ` +
		`field." }`,
)

func generateServiceIDRequiredResponse() []byte {
	return responseServiceIDRequired
}

var responsePlanIDRequired = []byte(
	`{ "error": "PlanIdRequired", "description": "plan_id is a required ` +
		`field." }`)

func generatePlanIDRequiredResponse() []byte {
	return responsePlanIDRequired
}

var responseInvalidServiceID = []byte(
	`{ "error": "InvalidServiceId", "description": "The provided service_id is ` +
		`invalid." }`,
)

func generateInvalidServiceIDResponse() []byte {
	return responseInvalidServiceID
}

var responseInvalidPlanID = []byte(
	`{ "error": "InvalidPlanId", "description": "The provided plan_id is ` +
		`invalid." }`,
)

func generateInvalidPlanIDResponse() []byte {
	return responseInvalidPlanID
}

var responseStateMigrating = []byte(
	`{ "error": "UnprocessableState", "description": "Instance is ` +
		`migrating. Please try again later." }`,
)

func generateStateMigratingResponse() []byte {
	return responseStateMigrating
}

var responseProvisioningAccepted = []byte(
	fmt.Sprintf(`{ "operation": "%s" }`, operations.Provisioning),
)

func generateProvisionAcceptedResponse() []byte {
	return responseProvisioningAccepted
}

var responseDeprovisioningAccepted = []byte(
	fmt.Sprintf(`{ "operation": "%s" }`, operations.Deprovisioning),
)

func generateDeprovisionAcceptedResponse() []byte {
	return responseDeprovisioningAccepted
}

var responseInProgress = []byte(
	fmt.Sprintf(`{ "state": "%s" }`, operations.StateInProgress),
)

func generateOperationInProgressResponse() []byte {
	return responseInProgress
}

var responseSucceeded = []byte(
	fmt.Sprintf(`{ "state": "%s" }`, operations.StateSucceeded),
)

func generateOperationSucceededResponse() []byte {
	return responseSucceeded
}

var responseFailed = []byte(
	fmt.Sprintf(`{ "state": "%s" }`, operations.StateFailed),
)

func generateOperationFailedResponse() []byte {
	return responseFailed
}

var responseEmptyJSON = []byte("{}")

func generateEmptyResponse() []byte {
	return responseEmptyJSON
}

var responseConflict = []byte(`{ "description": "A service instance exists ` +
	`with the specified service id" }`)

func generateConflictResponse() []byte {
	return responseConflict
}

var responseBindingConflict = []byte(`{ "description": "A service binding exists ` +
	`with the specified binding id" }`)

func generateBindingConflictResponse() []byte {
	return responseBindingConflict
}

var instanceNotFoundText = []byte("Instance not found")

func generateInstanceNotFoundResponse() []byte {
	return []byte(fmt.Sprintf(responseMalformedParameterBody, instanceNotFoundText))
}

// The following are custom to this broker-- i.e. not explicitly declared by
// the OSB spec

var responseMalformedParameterBody = `{ "error": "MalformedRequestBody", "description": "The request body did ` +
	`not contain valid: %s" }`

func generateMalformedParameterResponse(detail string) []byte {
	return []byte(fmt.Sprintf(responseMalformedParameterBody, detail))
}

var responseMalformedRequestBody = []byte(
	`{ "error": "MalformedRequestBody", "description": "The request body did ` +
		`not contain valid, well-formed JSON" }`,
)

func generateMalformedRequestResponse() []byte {
	return responseMalformedRequestBody
}

var responseOperationRequired = []byte(
	`{ "error": "OperationRequired", "description": "The polling request did ` +
		`not include the required operation query parameter" }`,
)

func generateOperationRequiredResponse() []byte {
	return responseOperationRequired
}

var responseOperationInvalid = []byte(
	`{ "error": "OperationInvalid", "description": "The polling request ` +
		`included an invalid value for the required operation query parameter" }`,
)

func generateOperationInvalidResponse() []byte {
	return responseOperationInvalid
}
