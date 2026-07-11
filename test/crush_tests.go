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

package test

import "testing"

const crush_dependency_name = "github.com/charmbracelet/crush"
const crush_module_name = "crush"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase(
			"crush-basic-test",
			crush_module_name,
			"", "",
			"1.26", "",
			TestCrushBasic,
		),
		NewMuzzleTestCase(
			"crush-muzzle-test",
			crush_dependency_name,
			crush_module_name,
			"", "",
			"1.26", "",
			[]string{"go", "build", "test_crush_basic.go"},
		),
		NewLatestDepthTestCase(
			"crush-latest-depth-test",
			crush_dependency_name,
			crush_module_name,
			"", "",
			"1.26", "",
			TestCrushBasic,
		),
	)
}

func TestCrushBasic(t *testing.T, env ...string) {
	UseApp("crush/devel")
	RunGoBuild(t, "go", "build", "test_crush_basic.go")
	RunApp(t, "test_crush_basic", env...)
}
