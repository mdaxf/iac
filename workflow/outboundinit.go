package workflow

import "context"

// SendToHubFunc is injected from contollershandlers.go after HubExecutorManager is initialised.
// See engine/function/outboundinit.go for full documentation.
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
