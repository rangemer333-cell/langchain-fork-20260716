// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package gin

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/gin-gonic/gin"
)

//go:linkname nextOnEnter github.com/gin-gonic/gin.nextOnEnter
func nextOnEnter(call api.CallContext, c *gin.Context) {
	if !ginEnabler.Enable() {
		return
	}
	if c == nil {
		return
	}
	lcs := trace.LocalRootSpanFromGLS()
	if lcs == nil {
		return
	}
	// Override the default route-template span name with the handler function
	// name so traces point at the code that actually served the request.
	if handlerName := c.HandlerName(); handlerName != "" {
		lcs.SetName(handlerName)
	}
}
