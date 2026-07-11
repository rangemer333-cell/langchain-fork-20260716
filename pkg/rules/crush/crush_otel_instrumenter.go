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
	"context"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// --- Agent (invoke_agent) getters ---

type crushAgentCommonGetter struct{}

func (crushAgentCommonGetter) GetAIOperationName(request crushAgentRequest) string {
	return request.operationName
}

func (crushAgentCommonGetter) GetAISystem(request crushAgentRequest) string {
	return SystemCrush
}

func (crushAgentCommonGetter) GetGenAISpanKind(request crushAgentRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindWorkflow
	}
	return request.spanKind
}

// crushAgentAttrsExtractor adds gen_ai.span.kind and gen_ai.other_input.*
// attributes for the crush agent invocation span (coordinator.Run and
// sessionAgent.Run).
type crushAgentAttrsExtractor struct {
	Base ai.AICommonAttrsExtractor[crushAgentRequest, crushAgentResponse, crushAgentCommonGetter]
}

func (e crushAgentAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request crushAgentRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, crushAgentCommonGetter{}.GetGenAISpanKind(request).Attribute())
	if request.sessionID != "" {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.session_id").String(request.sessionID))
	}
	if request.userMessage != "" {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.user_message").String(truncate(request.userMessage, 4096)))
	}
	if request.maxTokens > 0 {
		attributes = append(attributes, attribute.Key("gen_ai.request.max_tokens").Int64(request.maxTokens))
	}
	if request.temperature != nil {
		attributes = append(attributes, attribute.Key("gen_ai.request.temperature").Float64(*request.temperature))
	}
	if request.topP != nil {
		attributes = append(attributes, attribute.Key("gen_ai.request.top_p").Float64(*request.topP))
	}
	return attributes, parentContext
}

func (e crushAgentAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request crushAgentRequest, response crushAgentResponse, err error) ([]attribute.KeyValue, context.Context) {
	attributes, ctx = e.Base.OnEnd(attributes, ctx, request, response, err)
	if len(request.finishReasons) > 0 {
		attributes = append(attributes, attribute.Key("gen_ai.response.finish_reasons").StringSlice(request.finishReasons))
	}
	if request.inputTokens > 0 {
		attributes = append(attributes, attribute.Key("gen_ai.usage.input_tokens").Int64(request.inputTokens))
	}
	if request.outputTokens > 0 {
		attributes = append(attributes, attribute.Key("gen_ai.usage.output_tokens").Int64(request.outputTokens))
	}
	if request.cacheReadTokens > 0 {
		attributes = append(attributes, attribute.Key("gen_ai.usage.cache_read_input_tokens").Int64(request.cacheReadTokens))
	}
	if request.cacheWriteTokens > 0 {
		attributes = append(attributes, attribute.Key("gen_ai.usage.cache_creation_input_tokens").Int64(request.cacheWriteTokens))
	}
	if request.stepCount > 0 {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.step_count").Int64(request.stepCount))
	}
	return attributes, ctx
}

// BuildCrushAgentInstrumenter builds the instrumenter for crush agent
// invocation spans (both coordinator.Run and sessionAgent.Run).
func BuildCrushAgentInstrumenter() instrumenter.Instrumenter[crushAgentRequest, crushAgentResponse] {
	builder := instrumenter.Builder[crushAgentRequest, crushAgentResponse]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[crushAgentRequest, crushAgentResponse]{
			Getter: crushAgentCommonGetter{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[crushAgentRequest]{}).
		AddAttributesExtractor(&crushAgentAttrsExtractor{
			Base: ai.AICommonAttrsExtractor[crushAgentRequest, crushAgentResponse, crushAgentCommonGetter]{
				CommonGetter: crushAgentCommonGetter{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.CRUSH_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

// --- Tool (execute_tool) getters ---

type crushToolCommonGetter struct{}

func (crushToolCommonGetter) GetAIOperationName(request crushToolRequest) string {
	return request.operationName
}

func (crushToolCommonGetter) GetAISystem(request crushToolRequest) string {
	return SystemCrush
}

func (crushToolCommonGetter) GetGenAISpanKind(request crushToolRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindWorkflow
	}
	return request.spanKind
}

// crushToolAttrsExtractor adds gen_ai.span.kind, gen_ai.tool.name, and
// gen_ai.tool.input/output attributes for the crush tool invocation span.
type crushToolAttrsExtractor struct {
	Base ai.AICommonAttrsExtractor[crushToolRequest, crushToolResponse, crushToolCommonGetter]
}

func (e crushToolAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request crushToolRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, crushToolCommonGetter{}.GetGenAISpanKind(request).Attribute())
	if request.toolName != "" {
		attributes = append(attributes, attribute.Key("gen_ai.tool.name").String(request.toolName))
	}
	if request.toolInput != "" {
		attributes = append(attributes, attribute.Key("gen_ai.tool.input").String(truncate(request.toolInput, 4096)))
	}
	return attributes, parentContext
}

func (e crushToolAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request crushToolRequest, response crushToolResponse, err error) ([]attribute.KeyValue, context.Context) {
	attributes, ctx = e.Base.OnEnd(attributes, ctx, request, response, err)
	if request.toolOutput != "" {
		attributes = append(attributes, attribute.Key("gen_ai.tool.output").String(truncate(request.toolOutput, 4096)))
	}
	if request.isError {
		attributes = append(attributes, attribute.Key("error.type").String("tool_error"))
	}
	return attributes, ctx
}

// BuildCrushToolInstrumenter builds the instrumenter for crush tool
// invocation spans (hookedTool.Run).
func BuildCrushToolInstrumenter() instrumenter.Instrumenter[crushToolRequest, crushToolResponse] {
	builder := instrumenter.Builder[crushToolRequest, crushToolResponse]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[crushToolRequest, crushToolResponse]{
			Getter: crushToolCommonGetter{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[crushToolRequest]{}).
		AddAttributesExtractor(&crushToolAttrsExtractor{
			Base: ai.AICommonAttrsExtractor[crushToolRequest, crushToolResponse, crushToolCommonGetter]{
				CommonGetter: crushToolCommonGetter{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.CRUSH_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

// truncate caps a string at maxBytes bytes (UTF-8 safe) to avoid oversized
// span attributes.
func truncate(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Walk rune-by-rune so we don't split a multi-byte sequence.
	n := 0
	for i := range s {
		if n >= maxBytes {
			return s[:i] + "...[truncated]"
		}
		n++
	}
	return s
}
