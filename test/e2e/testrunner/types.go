//go:build e2e

// Copyright 2024 Alexandre Mahdhaoui
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testrunner

// TestFile represents a parsed YAML test case file.
type TestFile struct {
	TestCases []TestCase `yaml:"testCases"`
}

// TestCase represents a single test case.
type TestCase struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Setup       []Step   `yaml:"setup,omitempty"`
	Steps       []Step   `yaml:"steps"`
	Teardown    []Step   `yaml:"teardown,omitempty"`
}

// Step represents a single action in a test case.
type Step struct {
	ID          string                 `yaml:"id,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Mode        string                 `yaml:"mode,omitempty"` // "cli" or "mcp"
	Tool        string                 `yaml:"tool,omitempty"` // MCP tool name (when mode=mcp)
	Command     string                 `yaml:"command"`        // CLI subcommand or harness command
	Input       map[string]interface{} `yaml:"input,omitempty"`
	Expected    map[string]interface{} `yaml:"expected,omitempty"`
	Capture     map[string]string      `yaml:"capture,omitempty"`
}

// TemplateData holds data available for Go template rendering.
type TemplateData struct {
	Workspace            string                            // temp workspace root
	CURepoPath           string                            // path to CU repo clone
	GitServerURL         string                            // git server base URL
	GitServerContainerID string                            // docker container ID for git server
	Binary               string                            // path to forge-cu binary
	Repos                map[string]map[string]interface{} // per-repo context
	Steps                map[string]map[string]interface{} // captured step results
}
