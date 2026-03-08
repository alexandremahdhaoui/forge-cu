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
	"os/exec"
	"path/filepath"
	"strings"
)

// repoSpec describes a repository to create during workspace initialization.
type repoSpec struct {
	Name         string
	ManagedFiles []string
}

// InitWorkspace sets up the test workspace with git repos, a CU repo,
// managed files, and symlinks.
//
// Input fields:
//   - repos: list of {name, managedFiles} specs
//
// The git server runs inside a Docker container identified by
// data.GitServerContainerID. Server-side git commands use docker exec.
func InitWorkspace(data *TemplateData, input map[string]interface{}) error {
	specs, err := parseRepoSpecs(input)
	if err != nil {
		return fmt.Errorf("init-workspace: %w", err)
	}

	if data.Repos == nil {
		data.Repos = make(map[string]map[string]interface{})
	}

	containerID := data.GitServerContainerID

	// 1. Create each repo: bare on server, clone to workspace, initial commit.
	for _, spec := range specs {
		if err := createRepo(data, containerID, spec); err != nil {
			return fmt.Errorf("init-workspace: creating repo %q: %w", spec.Name, err)
		}
	}

	// 2. Create the CU repo.
	if err := createCURepo(data, containerID, specs); err != nil {
		return fmt.Errorf("init-workspace: creating cu-repo: %w", err)
	}

	return nil
}

// WriteFile writes content to a file at the given path. Both path and content
// are rendered as templates.
//
// Input fields:
//   - path: file path (template)
//   - content: file content (template)
func WriteFile(data *TemplateData, input map[string]interface{}) error {
	pathRaw, ok := input["path"]
	if !ok {
		return fmt.Errorf("write-file: missing 'path' in input")
	}
	pathStr, ok := pathRaw.(string)
	if !ok {
		return fmt.Errorf("write-file: 'path' must be a string")
	}

	contentRaw, ok := input["content"]
	if !ok {
		return fmt.Errorf("write-file: missing 'content' in input")
	}
	contentStr, ok := contentRaw.(string)
	if !ok {
		return fmt.Errorf("write-file: 'content' must be a string")
	}

	renderedPath, err := RenderTemplate(pathStr, data)
	if err != nil {
		return fmt.Errorf("write-file: rendering path: %w", err)
	}

	renderedContent, err := RenderTemplate(contentStr, data)
	if err != nil {
		return fmt.Errorf("write-file: rendering content: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(renderedPath), 0o755); err != nil {
		return fmt.Errorf("write-file: creating directory: %w", err)
	}

	if err := os.WriteFile(renderedPath, []byte(renderedContent), 0o644); err != nil {
		return fmt.Errorf("write-file: writing %q: %w", renderedPath, err)
	}

	return nil
}

// parseRepoSpecs extracts repo specs from the input map.
func parseRepoSpecs(input map[string]interface{}) ([]repoSpec, error) {
	reposRaw, ok := input["repos"]
	if !ok {
		return nil, fmt.Errorf("missing 'repos' in input")
	}

	reposList, ok := reposRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'repos' must be a list")
	}

	var specs []repoSpec
	for i, item := range reposList {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("repos[%d]: must be a map", i)
		}

		nameRaw, ok := m["name"]
		if !ok {
			return nil, fmt.Errorf("repos[%d]: missing 'name'", i)
		}
		name, ok := nameRaw.(string)
		if !ok {
			return nil, fmt.Errorf("repos[%d]: 'name' must be a string", i)
		}

		var managedFiles []string
		if mfRaw, ok := m["managedFiles"]; ok {
			mfList, ok := mfRaw.([]interface{})
			if !ok {
				return nil, fmt.Errorf("repos[%d]: 'managedFiles' must be a list", i)
			}
			for j, f := range mfList {
				s, ok := f.(string)
				if !ok {
					return nil, fmt.Errorf("repos[%d]: managedFiles[%d] must be a string", i, j)
				}
				managedFiles = append(managedFiles, s)
			}
		}

		specs = append(specs, repoSpec{Name: name, ManagedFiles: managedFiles})
	}

	return specs, nil
}

// createRepo creates a bare repo on the git server, clones it to the
// workspace, creates an initial commit with managed files, and pushes.
func createRepo(data *TemplateData, containerID string, spec repoSpec) error {
	barePath := fmt.Sprintf("/srv/git/%s.git", spec.Name)
	clonePath := filepath.Join(data.Workspace, spec.Name)
	cloneURL := fmt.Sprintf("%s/%s.git", data.GitServerURL, spec.Name)

	// Create bare repo on git server via docker exec.
	if err := dockerExec(containerID, "git", "init", "--bare", barePath); err != nil {
		return fmt.Errorf("creating bare repo: %w", err)
	}
	// Mark repo as exportable for git daemon.
	if err := dockerExec(containerID, "touch", barePath+"/git-daemon-export-ok"); err != nil {
		return fmt.Errorf("marking repo exportable: %w", err)
	}

	// Clone to workspace.
	if err := runCmd(data.Workspace, "git", "clone", cloneURL, clonePath); err != nil {
		return fmt.Errorf("cloning repo: %w", err)
	}

	// Configure git user for commits.
	if err := runCmd(clonePath, "git", "config", "user.email", "test@test.com"); err != nil {
		return err
	}
	if err := runCmd(clonePath, "git", "config", "user.name", "Test"); err != nil {
		return err
	}

	// Create managed files (empty) and commit.
	for _, f := range spec.ManagedFiles {
		fPath := filepath.Join(clonePath, f)
		if err := os.MkdirAll(filepath.Dir(fPath), 0o755); err != nil {
			return fmt.Errorf("creating dir for %q: %w", f, err)
		}
		if err := os.WriteFile(fPath, []byte(""), 0o644); err != nil {
			return fmt.Errorf("creating file %q: %w", f, err)
		}
	}

	if err := runCmd(clonePath, "git", "add", "."); err != nil {
		return err
	}
	// Use --allow-empty in case files are already tracked from a prior clone.
	if err := runCmd(clonePath, "git", "commit", "--allow-empty", "-m", "initial commit"); err != nil {
		return err
	}
	if err := runCmd(clonePath, "git", "push", "origin", "HEAD"); err != nil {
		return fmt.Errorf("pushing initial commit: %w", err)
	}

	// Store repo context.
	data.Repos[spec.Name] = map[string]interface{}{
		"Name": spec.Name,
		"Path": clonePath,
		"URL":  cloneURL,
	}

	return nil
}

// createCURepo creates the CU composition repo with compo.yaml, copies
// managed files, creates symlinks, and pushes to the git server.
func createCURepo(data *TemplateData, containerID string, specs []repoSpec) error {
	barePath := "/srv/git/cu-repo.git"
	clonePath := filepath.Join(data.Workspace, "cu-repo")
	cloneURL := fmt.Sprintf("%s/cu-repo.git", data.GitServerURL)

	// Create bare repo on git server.
	if err := dockerExec(containerID, "git", "init", "--bare", barePath); err != nil {
		return fmt.Errorf("creating bare cu-repo: %w", err)
	}
	// Mark repo as exportable for git daemon.
	if err := dockerExec(containerID, "touch", barePath+"/git-daemon-export-ok"); err != nil {
		return fmt.Errorf("marking cu-repo exportable: %w", err)
	}

	// Clone to workspace.
	if err := runCmd(data.Workspace, "git", "clone", cloneURL, clonePath); err != nil {
		return fmt.Errorf("cloning cu-repo: %w", err)
	}

	// Configure git user.
	if err := runCmd(clonePath, "git", "config", "user.email", "test@test.com"); err != nil {
		return err
	}
	if err := runCmd(clonePath, "git", "config", "user.name", "Test"); err != nil {
		return err
	}

	// Build compo.yaml content.
	compoYAML := buildCompoYAML(data, specs)
	compoPath := filepath.Join(clonePath, "compo.yaml")
	if err := os.WriteFile(compoPath, []byte(compoYAML), 0o644); err != nil {
		return fmt.Errorf("writing compo.yaml: %w", err)
	}

	// For each repo, copy managed files into cu-repo/<reponame>/ and
	// create symlinks from workspace/<reponame>/<file> to cu-repo/<reponame>/<file>.
	for _, spec := range specs {
		repoSubdir := filepath.Join(clonePath, spec.Name)
		if err := os.MkdirAll(repoSubdir, 0o755); err != nil {
			return fmt.Errorf("creating cu-repo subdir %q: %w", spec.Name, err)
		}

		repoClonePath := filepath.Join(data.Workspace, spec.Name)

		for _, f := range spec.ManagedFiles {
			srcPath := filepath.Join(repoClonePath, f)
			cuFilePath := filepath.Join(repoSubdir, f)

			// Create parent dirs in CU repo.
			if err := os.MkdirAll(filepath.Dir(cuFilePath), 0o755); err != nil {
				return fmt.Errorf("creating dir for cu file %q: %w", f, err)
			}

			// Copy the file content from the repo clone to the CU repo.
			content, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("reading %q: %w", srcPath, err)
			}
			if err := os.WriteFile(cuFilePath, content, 0o644); err != nil {
				return fmt.Errorf("writing cu file %q: %w", cuFilePath, err)
			}

			// Remove the original file and create a symlink.
			if err := os.Remove(srcPath); err != nil {
				return fmt.Errorf("removing %q for symlink: %w", srcPath, err)
			}
			if err := os.Symlink(cuFilePath, srcPath); err != nil {
				return fmt.Errorf("creating symlink %q -> %q: %w", srcPath, cuFilePath, err)
			}
		}
	}

	// Commit and push.
	if err := runCmd(clonePath, "git", "add", "."); err != nil {
		return err
	}
	if err := runCmd(clonePath, "git", "commit", "--allow-empty", "-m", "initial cu-repo commit"); err != nil {
		return err
	}
	if err := runCmd(clonePath, "git", "push", "origin", "HEAD"); err != nil {
		return fmt.Errorf("pushing cu-repo: %w", err)
	}

	data.CURepoPath = clonePath

	return nil
}

// buildCompoYAML generates compo.yaml content from the repo specs.
func buildCompoYAML(data *TemplateData, specs []repoSpec) string {
	var b strings.Builder
	b.WriteString("name: test-compo\n")
	b.WriteString("repos:\n")
	for _, spec := range specs {
		repoURL := fmt.Sprintf("%s/%s.git", data.GitServerURL, spec.Name)
		b.WriteString(fmt.Sprintf("  - name: %s\n", spec.Name))
		b.WriteString(fmt.Sprintf("    url: %s\n", repoURL))
		if len(spec.ManagedFiles) > 0 {
			b.WriteString("    managedFiles:\n")
			for _, f := range spec.ManagedFiles {
				b.WriteString(fmt.Sprintf("      - %s\n", f))
			}
		}
	}
	return b.String()
}

// dockerExec runs a command inside the git server Docker container.
func dockerExec(containerID string, args ...string) error {
	cmdArgs := append([]string{"exec", containerID}, args...)
	cmd := exec.Command("docker", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker exec %v: %w\n%s", args, err, string(out))
	}
	return nil
}

// runCmd runs a command in the given directory.
func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, string(out))
	}
	return nil
}
