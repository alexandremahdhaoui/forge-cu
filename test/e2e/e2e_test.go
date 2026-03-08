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

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandremahdhaoui/forge-cu/test/e2e/testrunner"
)

var (
	binaryPath         string
	gitServerURL       string
	gitServerContainer string
)

func TestMain(m *testing.M) {
	// Locate the forge-cu binary.
	bin, err := findBinary("forge-cu")
	if err != nil {
		fmt.Fprintf(os.Stderr, "finding forge-cu binary: %v\n", err)
		os.Exit(1)
	}
	binaryPath = bin

	// Determine the project root (tests run from the package directory).
	root, err := projectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolving project root: %v\n", err)
		os.Exit(1)
	}

	// Build the git-server Docker image.
	if err := buildGitServerImage(root); err != nil {
		fmt.Fprintf(os.Stderr, "building git-server image: %v\n", err)
		os.Exit(1)
	}

	// Start the git-server container.
	containerID, err := startGitServerContainer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "starting git-server container: %v\n", err)
		os.Exit(1)
	}
	gitServerContainer = containerID
	gitServerURL = "git://localhost:9418"

	// Run all tests.
	code := m.Run()

	// Cleanup: stop and remove the git-server container.
	stopGitServerContainer(gitServerContainer)

	os.Exit(code)
}

func TestE2E(t *testing.T) {
	root, err := projectRoot()
	if err != nil {
		t.Fatalf("resolving project root: %v", err)
	}
	testdataDir := filepath.Join(root, "test", "e2e", "testdata")

	testFiles, filePaths, err := testrunner.LoadTestFiles(testdataDir)
	if err != nil {
		t.Fatalf("loading test files: %v", err)
	}

	for i, tf := range testFiles {
		// Use the filename (without directory and extension) as the
		// top-level subtest name.
		fileName := filepath.Base(filePaths[i])
		fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

		t.Run(fileName, func(t *testing.T) {
			for _, tc := range tf.TestCases {
				t.Run(tc.Name, func(t *testing.T) {
					workspace := t.TempDir()

					data := &testrunner.TemplateData{
						Workspace:            workspace,
						CURepoPath:           "",
						GitServerURL:         gitServerURL,
						Binary:               binaryPath,
						GitServerContainerID: gitServerContainer,
						Repos:                make(map[string]map[string]interface{}),
						Steps:                make(map[string]map[string]interface{}),
					}

					if err := testrunner.RunTestCase(data, tc); err != nil {
						t.Fatalf("test case %q failed: %v", tc.Name, err)
					}
				})
			}
		})
	}
}
