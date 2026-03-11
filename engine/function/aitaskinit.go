// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package funcs

import "context"

// InitializeAITaskConfig sets up the global AI config accessor
// This should be called from main during initialization to avoid import cycles
func InitializeAITaskConfig(getConfigFunc func() interface{}) {
	GetAIConfigFunc = getConfigFunc
}

// Agent runtime function variables — injected from main to avoid import cycles.
//
// RunAgentSyncFunc executes a named agent (by ID or name) synchronously and
// returns the final assistant text response.
// agentID takes priority; agentName is used when agentID is empty.
var RunAgentSyncFunc func(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error)

// ExecuteSkillFunc calls a single skill tool by name with a JSON arguments payload.
var ExecuteSkillFunc func(ctx context.Context, skillName, argsJSON string) (string, error)

// InitializeAgentRuntime wires up the agent runner and skill executor callbacks.
// Call this from main after both the agent runner service and tool registry are ready.
func InitializeAgentRuntime(
	runAgentSync func(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error),
	executeSkill func(ctx context.Context, skillName, argsJSON string) (string, error),
) {
	RunAgentSyncFunc = runAgentSync
	ExecuteSkillFunc = executeSkill
}
