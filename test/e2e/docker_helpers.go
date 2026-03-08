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
	"os/exec"
	"path/filepath"
	"strings"
)

// buildGitServerImage builds the git-server Docker image from the
// project's containers/git-server/Containerfile.
func buildGitServerImage(projectRoot string) error {
	containerfile := filepath.Join(projectRoot, "containers", "git-server", "Containerfile")
	cmd := exec.Command("docker", "build",
		"-f", containerfile,
		"-t", "forge-cu-test-git-server",
		projectRoot,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("building git-server image: %w\n%s", err, string(out))
	}
	return nil
}

// startGitServerContainer starts the git-server Docker container and
// returns its container ID. The container binds host port 9418 to
// container port 9418 (git protocol).
func startGitServerContainer() (string, error) {
	cmd := exec.Command("docker", "run",
		"-d",
		"--name", "forge-cu-e2e-git-server",
		"-p", "9418:9418",
		"forge-cu-test-git-server",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("starting git-server container: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// stopGitServerContainer stops and removes the git-server Docker
// container. Errors are logged but not returned, since this runs
// during cleanup.
func stopGitServerContainer(containerID string) {
	stop := exec.Command("docker", "stop", containerID)
	_ = stop.Run()

	rm := exec.Command("docker", "rm", "-f", containerID)
	_ = rm.Run()
}

// projectRoot returns the project root by walking up from the test
// package directory (test/e2e/) to find go.mod.
func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// findBinary locates the forge-cu binary at <project-root>/build/bin/<name>
// and returns its absolute path.
func findBinary(name string) (string, error) {
	root, err := projectRoot()
	if err != nil {
		return "", fmt.Errorf("finding project root: %w", err)
	}
	abs := filepath.Join(root, "build", "bin", name)

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("binary not found at %q: %w", abs, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory, not a binary", abs)
	}

	return abs, nil
}
