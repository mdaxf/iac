package services

import "context"

// agentNotifConfigKey is the unexported context key used to pass per-agent
// notification config (map[string]string from agent_definitions.notification_config)
// through the tool execution context.
type agentNotifConfigKey struct{}

// WithAgentNotifConfig returns a child context that carries the agent's
// notification_config JSONB map. The runner injects this before the tool loop
// so that notification tools can resolve per-agent credentials without requiring
// the LLM to supply them explicitly.
func WithAgentNotifConfig(ctx context.Context, cfg map[string]string) context.Context {
	return context.WithValue(ctx, agentNotifConfigKey{}, cfg)
}

// agentNotifConfigFromCtx retrieves the per-agent notification config from ctx.
// Returns an empty map if none was injected.
func agentNotifConfigFromCtx(ctx context.Context) map[string]string {
	if v, ok := ctx.Value(agentNotifConfigKey{}).(map[string]string); ok {
		return v
	}
	return map[string]string{}
}
