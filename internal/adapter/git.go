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

package adapter

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

// Compile-time interface check.
var _ GitAdapter = (*gitAdapter)(nil)

// gitAdapter implements GitAdapter using os/exec to run git commands.
type gitAdapter struct{}

// NewGitAdapter returns a new GitAdapter backed by os/exec.
func NewGitAdapter() GitAdapter {
	return &gitAdapter{}
}

func (g *gitAdapter) Clone(ctx context.Context, url, dest string) error {
	if err := g.run(ctx, "", "clone", url, dest); err != nil {
		return fmt.Errorf("git clone %s to %s: %w", url, dest, err)
	}
	return nil
}

func (g *gitAdapter) Checkout(ctx context.Context, repoPath, branch string) error {
	if err := g.run(ctx, repoPath, "checkout", branch); err != nil {
		// Branch might not exist yet; try creating it.
		if err2 := g.run(ctx, repoPath, "checkout", "-b", branch); err2 != nil {
			return fmt.Errorf("git checkout %s in %s: %w", branch, repoPath, err2)
		}
	}
	return nil
}

func (g *gitAdapter) Status(ctx context.Context, repoPath string) ([]types.DepChange, error) {
	out, err := g.output(ctx, repoPath, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git status in %s: %w", repoPath, err)
	}

	changes := make([]types.DepChange, 0)
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 4 {
			// Porcelain format: 2 status chars + space + path = minimum 4 chars.
			continue
		}

		statusCode := line[:2]
		filePath := line[3:]

		status := parseStatusCode(statusCode)
		repoName, file := splitRepoPath(filePath)

		changes = append(changes, types.DepChange{
			RepoName: repoName,
			File:     file,
			Status:   status,
		})
	}

	return changes, nil
}

func (g *gitAdapter) Commit(ctx context.Context, repoPath, message string) error {
	if err := g.run(ctx, repoPath, "add", "."); err != nil {
		return fmt.Errorf("git add in %s: %w", repoPath, err)
	}
	if err := g.run(ctx, repoPath, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit in %s: %w", repoPath, err)
	}
	return nil
}

func (g *gitAdapter) Push(ctx context.Context, repoPath string) error {
	if err := g.run(ctx, repoPath, "push"); err != nil {
		return fmt.Errorf("git push in %s: %w", repoPath, err)
	}
	return nil
}

func (g *gitAdapter) Pull(ctx context.Context, repoPath string) error {
	if err := g.run(ctx, repoPath, "pull"); err != nil {
		return fmt.Errorf("git pull in %s: %w", repoPath, err)
	}
	return nil
}

func (g *gitAdapter) ListBranches(ctx context.Context, repoPath string) ([]string, error) {
	out, err := g.output(ctx, repoPath, "branch", "--list", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("git branch --list in %s: %w", repoPath, err)
	}

	var branches []string
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		if b := strings.TrimSpace(scanner.Text()); b != "" {
			branches = append(branches, b)
		}
	}

	return branches, nil
}

func (g *gitAdapter) Diff(ctx context.Context, repoPath string) (string, error) {
	out, err := g.output(ctx, repoPath, "diff")
	if err != nil {
		return "", fmt.Errorf("git diff in %s: %w", repoPath, err)
	}
	return out, nil
}

func (g *gitAdapter) CurrentCommitHash(ctx context.Context, repoPath string) (string, error) {
	out, err := g.output(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD in %s: %w", repoPath, err)
	}
	return strings.TrimSpace(out), nil
}

func (g *gitAdapter) CurrentBranch(ctx context.Context, repoPath string) (string, error) {
	out, err := g.output(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD in %s: %w", repoPath, err)
	}
	return strings.TrimSpace(out), nil
}

// run executes a git command in the given directory without capturing output.
func (g *gitAdapter) run(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// output executes a git command in the given directory and returns its stdout.
func (g *gitAdapter) output(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// parseStatusCode maps git porcelain status codes to human-readable status strings.
func parseStatusCode(code string) string {
	// Porcelain format: XY where X is staging status, Y is working tree status.
	// We check both characters to determine the overall status.
	code = strings.TrimSpace(code)
	switch {
	case code == "??" || strings.ContainsAny(code, "A"):
		return "added"
	case strings.ContainsAny(code, "D"):
		return "deleted"
	case strings.ContainsAny(code, "R"):
		return "renamed"
	default:
		return "modified"
	}
}

// splitRepoPath splits a file path like "forge/go.mod" into repo name ("forge")
// and file ("go.mod"). If there is no slash, repo name is empty.
func splitRepoPath(path string) (repoName, file string) {
	idx := strings.Index(path, "/")
	if idx < 0 {
		return "", path
	}
	return path[:idx], path[idx+1:]
}
