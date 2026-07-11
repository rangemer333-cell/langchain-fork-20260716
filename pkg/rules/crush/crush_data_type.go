// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package crush

import (
	"os"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
)

type crushInnerEnabler struct {
	enabled bool
}

func (e crushInnerEnabler) Enable() bool {
	return e.enabled
}

var crushEnabler = crushInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_CRUSH_ENABLED") != "false"}

const (
	OperationInvokeAgent = "invoke_agent"
	OperationExecuteTool  = "execute_tool"
	SystemCrush           = "crush"
)

// crushAgentRequest holds the data extracted from a crush agent invocation
// (coordinator.Run or sessionAgent.Run). The sessionAgentCall field carries
// the original SessionAgentCall value (an unexported crush struct) extracted
// via reflection at hook time, so we can read its exported fields without
// importing crush's internal package.
type crushAgentRequest struct {
	operationName  string
	spanKind       ai.GenAISpanKind
	sessionID      string
	userMessage    string
	maxTokens      int64
	temperature    *float64
	topP           *float64
	stepCount      int64
	inputTokens    int64
	outputTokens   int64
	cacheReadTokens int64
	cacheWriteTokens int64
	finishReasons  []string
}

// crushToolRequest holds the data extracted from a hookedTool.Run call.
type crushToolRequest struct {
	operationName string
	spanKind      ai.GenAISpanKind
	toolName      string
	toolInput     string
	toolOutput    string
	isError       bool
}

// crushAgentResponse and crushToolResponse are reserved; the actual return
// values are inspected in the OnExit hook and folded into the request fields
// before End is called.
type crushAgentResponse struct{}
type crushToolResponse struct{}
