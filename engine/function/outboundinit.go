package funcs

import "context"

// SendToHubFunc is injected from contollershandlers.go after HubExecutorManager is initialised.
// It dispatches an outbound send to the correct protocol handler (HTTP, MQTT, Kafka, MCP, …)
// based on the protocol configured in the Integration Hub for the given protocol group.
//
// Signature mirrors HubExecutorManager.SendToEndpoint:
//
//	hubID            – MongoDB _id of the IntegrationHub document
//	protocolGroupID  – ID of the outbound protocol group within that hub
//	endpointID       – ID of the endpoint within that protocol group
//	payload          – raw message body
//	contentType      – e.g. "application/json" (only used for HTTP-family protocols)
//	method           – HTTP method override; empty means use the endpoint's configured method
//
// Returns: responseBody, statusCode, error
var SendToHubFunc func(
	ctx context.Context,
	hubID, protocolGroupID, endpointID string,
	payload []byte,
	contentType, method string,
) (string, int, error)

// InitializeOutboundRuntime injects the hub send function.
// Called from contollershandlers.go once HubExecutorManager is ready.
func InitializeOutboundRuntime(fn func(
	ctx context.Context,
	hubID, protocolGroupID, endpointID string,
	payload []byte,
	contentType, method string,
) (string, int, error)) {
	SendToHubFunc = fn
}
