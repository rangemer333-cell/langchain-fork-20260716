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
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"

	"charm.land/fantasy"
)

var crushAgentInstrumenter = BuildCrushAgentInstrumenter()
var crushToolInstrumenter = BuildCrushToolInstrumenter()

// --- H1: coordinator.Run ---
//
// Signature: func (c *coordinator) Run(ctx context.Context, sessionID string, prompt string, attachments ...message.Attachment) (*fantasy.AgentResult, error)
//
// coordinator is unexported and lives in crush's internal/agent package, so
// we declare the receiver as interface{} and let the loongsuite tool
// reconcile the trampoline types via //go:linkname. attachments is a variadic
// of the internal message.Attachment type, so it is also declared as a
// variadic of interface{}.

//go:linkname coordinatorRunOnEnter github.com/charmbracelet/crush/internal/agent.coordinatorRunOnEnter
func coordinatorRunOnEnter(call api.CallContext, c interface{}, ctx context.Context, sessionID string, prompt string, attachments ...interface{}) {
	if !crushEnabler.Enable() {
		return
	}
	request := crushAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindWorkflow,
		sessionID:     sessionID,
		userMessage:   truncate(prompt, 4096),
	}
	instrumentedCtx := crushAgentInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname coordinatorRunOnExit github.com/charmbracelet/crush/internal/agent.coordinatorRunOnExit
func coordinatorRunOnExit(call api.CallContext, result *fantasy.AgentResult, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(crushAgentRequest)
	if ctx == nil {
		return
	}
	if result != nil {
		request.finishReasons = []string{string(result.Response.FinishReason)}
		request.inputTokens = result.TotalUsage.InputTokens
		request.outputTokens = result.TotalUsage.OutputTokens
		request.cacheReadTokens = result.TotalUsage.CacheReadTokens
		request.cacheWriteTokens = result.TotalUsage.CacheCreationTokens
	}
	crushAgentInstrumenter.End(ctx, request, crushAgentResponse{}, err)
}

// --- H2: sessionAgent.Run ---
//
// Signature: func (a *sessionAgent) Run(ctx context.Context, call SessionAgentCall) (*fantasy.AgentResult, error)
//
// sessionAgent is unexported and SessionAgentCall is exported but lives in
// crush's internal/agent package, so both are declared as interface{}.
// Fields are read via reflection since the SessionAgentCall type is not
// importable from this plugin module.

//go:linkname sessionAgentRunOnEnter github.com/charmbracelet/crush/internal/agent.sessionAgentRunOnEnter
func sessionAgentRunOnEnter(call api.CallContext, a interface{}, ctx context.Context, sessionAgentCall interface{}) {
	if !crushEnabler.Enable() {
		return
	}
	request := crushAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindWorkflow,
		sessionID:     readStringField(sessionAgentCall, "SessionID"),
		userMessage:   truncate(readStringField(sessionAgentCall, "Prompt"), 4096),
		maxTokens:     readInt64Field(sessionAgentCall, "MaxOutputTokens"),
	}
	if temp := readFloat64PtrField(sessionAgentCall, "Temperature"); temp != nil {
		request.temperature = temp
	}
	if topP := readFloat64PtrField(sessionAgentCall, "TopP"); topP != nil {
		request.topP = topP
	}
	instrumentedCtx := crushAgentInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname sessionAgentRunOnExit github.com/charmbracelet/crush/internal/agent.sessionAgentRunOnExit
func sessionAgentRunOnExit(call api.CallContext, result *fantasy.AgentResult, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(crushAgentRequest)
	if ctx == nil {
		return
	}
	if result != nil {
		request.finishReasons = []string{string(result.Response.FinishReason)}
		request.inputTokens = result.TotalUsage.InputTokens
		request.outputTokens = result.TotalUsage.OutputTokens
		request.cacheReadTokens = result.TotalUsage.CacheReadTokens
		request.cacheWriteTokens = result.TotalUsage.CacheCreationTokens
		if l := len(result.Steps); l > 0 {
			request.stepCount = int64(l)
		}
	}
	crushAgentInstrumenter.End(ctx, request, crushAgentResponse{}, err)
}

// --- H3: hookedTool.Run ---
//
// Signature: func (h *hookedTool) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error)
//
// hookedTool is unexported, so the receiver is interface{}. fantasy.ToolCall
// and fantasy.ToolResponse are exported and importable.

//go:linkname hookedToolRunOnEnter github.com/charmbracelet/crush/internal/agent.hookedToolRunOnEnter
func hookedToolRunOnEnter(call api.CallContext, h interface{}, ctx context.Context, toolCall fantasy.ToolCall) {
	if !crushEnabler.Enable() {
		return
	}
	request := crushToolRequest{
		operationName: OperationExecuteTool,
		spanKind:      ai.GenAISpanKindWorkflow,
		toolName:      toolCall.Name,
		toolInput:     truncate(toolCall.Input, 4096),
	}
	instrumentedCtx := crushToolInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname hookedToolRunOnExit github.com/charmbracelet/crush/internal/agent.hookedToolRunOnExit
func hookedToolRunOnExit(call api.CallContext, resp fantasy.ToolResponse, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(crushToolRequest)
	if ctx == nil {
		return
	}
	request.toolOutput = truncate(resp.Content, 4096)
	if resp.IsError {
		request.isError = true
	}
	crushToolInstrumenter.End(ctx, request, crushToolResponse{}, err)
}

// --- reflection helpers for unexported crush types ---
//
// SessionAgentCall is an exported type but lives in crush's internal/agent
// package, which cannot be imported from this plugin module. We read its
// exported fields via reflection so the hook can extract sessionID / prompt /
// maxTokens / temperature / topP without a direct import.

func readStringField(v interface{}, name string) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	f := rv.FieldByName(name)
	if !f.IsValid() {
		return ""
	}
	if f.Kind() == reflect.String {
		return f.String()
	}
	s, _ := f.Interface().(string)
	return s
}

func readInt64Field(v interface{}, name string) int64 {
	if v == nil {
		return 0
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return 0
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0
	}
	f := rv.FieldByName(name)
	if !f.IsValid() {
		return 0
	}
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f.Int()
	}
	return 0
}

func readFloat64PtrField(in interface{}, name string) *float64 {
	if in == nil {
		return nil
	}
	rv := reflect.ValueOf(in)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	f := rv.FieldByName(name)
	if !f.IsValid() {
		return nil
	}
	if f.Kind() != reflect.Ptr || f.IsNil() {
		return nil
	}
	elem := f.Elem()
	if elem.Kind() != reflect.Float64 {
		return nil
	}
	val := elem.Float()
	return &val
}
