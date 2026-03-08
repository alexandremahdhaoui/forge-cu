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
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadTestFiles globs for *.yaml files in the given directory (and one level
// of subdirectories) and parses each into a TestFile. It returns the parsed
// test files, their file paths (for naming test cases), and any error
// encountered. Note: Go's filepath.Glob does not support recursive ** globs,
// so only top-level and one-level-deep files are discovered.
func LoadTestFiles(dir string) ([]TestFile, []string, error) {
	pattern := filepath.Join(dir, "**", "*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("globbing %q: %w", pattern, err)
	}

	// filepath.Glob with ** does not recurse into subdirectories on all
	// platforms. Also glob the top-level directory to catch files that are
	// not nested.
	topLevel := filepath.Join(dir, "*.yaml")
	topMatches, err := filepath.Glob(topLevel)
	if err != nil {
		return nil, nil, fmt.Errorf("globbing %q: %w", topLevel, err)
	}

	// Deduplicate matches.
	seen := make(map[string]struct{}, len(matches)+len(topMatches))
	var allPaths []string
	for _, m := range append(topMatches, matches...) {
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		allPaths = append(allPaths, m)
	}

	if len(allPaths) == 0 {
		return nil, nil, fmt.Errorf("no YAML files found in %q", dir)
	}

	var testFiles []TestFile
	var filePaths []string
	for _, path := range allPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %q: %w", path, err)
		}

		var tf TestFile
		if err := yaml.Unmarshal(data, &tf); err != nil {
			return nil, nil, fmt.Errorf("parsing %q: %w", path, err)
		}

		testFiles = append(testFiles, tf)
		filePaths = append(filePaths, path)
	}

	return testFiles, filePaths, nil
}
