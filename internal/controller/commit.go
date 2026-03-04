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

package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexandremahdhaoui/forge-cu/internal/adapter"
	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

// CommitService handles committing dependency changes in the CU repo.
type CommitService interface {
	Commit(ctx context.Context, cuRepoPath, message string) ([]types.DepChange, string, error)
}

// Compile-time interface check.
var _ CommitService = (*commitService)(nil)

type commitService struct {
	git adapter.GitAdapter
}

// NewCommitService creates a CommitService with the given git adapter.
func NewCommitService(git adapter.GitAdapter) CommitService {
	return &commitService{git: git}
}

// Commit stages and commits pending changes in the CU repo. If message is empty,
// it auto-generates one from the changed files. Returns the list of changes and
// the commit hash.
func (c *commitService) Commit(ctx context.Context, cuRepoPath, message string) ([]types.DepChange, string, error) {
	changes, err := c.git.Status(ctx, cuRepoPath)
	if err != nil {
		return nil, "", fmt.Errorf("getting status: %w", err)
	}

	if len(changes) == 0 {
		return nil, "", fmt.Errorf("no pending changes")
	}

	if message == "" {
		message = generateCommitMessage(changes)
	}

	if err := c.git.Commit(ctx, cuRepoPath, message); err != nil {
		return nil, "", fmt.Errorf("committing changes: %w", err)
	}

	hash, err := c.git.CurrentCommitHash(ctx, cuRepoPath)
	if err != nil {
		return nil, "", fmt.Errorf("getting commit hash: %w", err)
	}

	return changes, hash, nil
}

// generateCommitMessage creates a commit message from the list of changed files.
func generateCommitMessage(changes []types.DepChange) string {
	var parts []string
	for _, c := range changes {
		parts = append(parts, c.RepoName+"/"+c.File)
	}
	return "cu: update " + strings.Join(parts, ", ")
}
