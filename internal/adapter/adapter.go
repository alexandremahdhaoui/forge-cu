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
	"context"

	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

// GitAdapter abstracts git operations on the CU repo.
// Auth is handled via environment (SSH agent, git credentials file).
type GitAdapter interface {
	// Clone clones a git repository to the destination path.
	Clone(ctx context.Context, url, dest string) error

	// Checkout switches the CU repo to the specified branch.
	Checkout(ctx context.Context, repoPath, branch string) error

	// Status returns uncommitted changes in the CU repo.
	Status(ctx context.Context, repoPath string) ([]types.DepChange, error)

	// Commit stages all changes and creates a commit with the given message.
	Commit(ctx context.Context, repoPath, message string) error

	// Push pushes commits to the remote.
	Push(ctx context.Context, repoPath string) error

	// Pull pulls changes from the remote.
	Pull(ctx context.Context, repoPath string) error

	// ListBranches returns all branches in the CU repo.
	ListBranches(ctx context.Context, repoPath string) ([]string, error)

	// Diff returns the unified diff of uncommitted changes.
	Diff(ctx context.Context, repoPath string) (string, error)

	// CurrentCommitHash returns the current HEAD commit hash.
	CurrentCommitHash(ctx context.Context, repoPath string) (string, error)

	// CurrentBranch returns the current branch name.
	CurrentBranch(ctx context.Context, repoPath string) (string, error)
}

// SymlinkAdapter manages symlinks between workspace repos and the CU repo.
type SymlinkAdapter interface {
	// Create creates symlinks for all managed files in the composition.
	// cuRepoPath: absolute path to the CU repo clone.
	// workspacePath: absolute path to the workspace root.
	Create(ctx context.Context, cuRepoPath, workspacePath string, compo types.Compo) error

	// Remove removes all symlinks and restores original files.
	Remove(ctx context.Context, workspacePath string, compo types.Compo) error

	// Verify checks that all expected symlinks exist and point to valid targets.
	Verify(ctx context.Context, workspacePath string, compo types.Compo) (bool, error)
}
