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

package engine

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/alexandremahdhaoui/forge-cu/internal/controller"
	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

// GoCUEngine handles Go-specific dependency operations.
type GoCUEngine interface {
	GoGet(ctx context.Context, repoDir, cuRepoPath, pkg, version string) ([]types.DepChange, string, error)
}

// Compile-time interface check.
var _ GoCUEngine = (*goCUEngine)(nil)

type goCUEngine struct {
	commitSvc controller.CommitService
}

// NewGoCUEngine creates a GoCUEngine with the given commit service.
func NewGoCUEngine(commitSvc controller.CommitService) GoCUEngine {
	return &goCUEngine{commitSvc: commitSvc}
}

// GoGet runs "go get pkg@version" and "go mod tidy" in repoDir, then commits
// the resulting changes in the CU repo.
func (e *goCUEngine) GoGet(ctx context.Context, repoDir, cuRepoPath, pkg, version string) ([]types.DepChange, string, error) {
	// Run go get pkg@version.
	cmd := exec.CommandContext(ctx, "go", "get", pkg+"@"+version)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("go get %s@%s: %w\n%s", pkg, version, err, out)
	}

	// Run go mod tidy.
	cmd = exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("go mod tidy: %w\n%s", err, out)
	}

	// Commit the changes in the CU repo.
	message := fmt.Sprintf("cu: go get %s@%s", pkg, version)
	changes, hash, err := e.commitSvc.Commit(ctx, cuRepoPath, message)
	if err != nil {
		return nil, "", fmt.Errorf("committing go get changes: %w", err)
	}

	return changes, hash, nil
}
