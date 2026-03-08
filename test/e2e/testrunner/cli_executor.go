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

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ExecuteCLIStep runs a forge-cu CLI subcommand and returns parsed JSON output.
//
//  1. Renders step.Input values through templates.
//  2. Builds command: binary <step.Command> --<key>=<value> for each input field.
//  3. Runs via exec.Command, captures stdout+stderr.
//  4. Parses stdout as JSON.
//  5. Returns parsed result with exitCode included.
func ExecuteCLIStep(binary string, step Step, data *TemplateData) (map[string]interface{}, error) {
	// Render input values.
	var renderedInput map[string]interface{}
	if step.Input != nil {
		var err error
		renderedInput, err = RenderMapValues(step.Input, data)
		if err != nil {
			return nil, fmt.Errorf("cli step %q: rendering input: %w", step.Command, err)
		}
	}

	// Build command arguments: flags first, then positional args.
	// Keys starting with "_" are treated as positional args (e.g., "_args").
	args := []string{step.Command}
	var positionalArgs []string
	for key, val := range renderedInput {
		if strings.HasPrefix(key, "_") {
			// Positional args: expect a slice or single value.
			switch v := val.(type) {
			case []interface{}:
				for _, item := range v {
					positionalArgs = append(positionalArgs, fmt.Sprintf("%v", item))
				}
			default:
				positionalArgs = append(positionalArgs, fmt.Sprintf("%v", v))
			}
			continue
		}
		args = append(args, fmt.Sprintf("--%s=%v", key, val))
	}
	args = append(args, positionalArgs...)

	cmd := exec.Command(binary, args...)
	stdout, err := cmd.Output()

	// Build result map with exit code.
	result := make(map[string]interface{})
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			// Include stderr in result for debugging.
			result["stderr"] = string(exitErr.Stderr)
		} else {
			return nil, fmt.Errorf("cli step %q: running command: %w", step.Command, err)
		}
	}
	result["exitCode"] = float64(exitCode)

	// Parse stdout as JSON if non-empty.
	stdoutStr := strings.TrimSpace(string(stdout))
	if stdoutStr != "" {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(stdoutStr), &parsed); err != nil {
			// If stdout is not valid JSON, store it as raw output.
			result["stdout"] = stdoutStr
		} else {
			// Merge parsed JSON into result.
			for k, v := range parsed {
				result[k] = v
			}
		}
	}

	return result, nil
}
