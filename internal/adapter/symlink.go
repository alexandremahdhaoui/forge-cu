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
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

var _ SymlinkAdapter = (*symlinkAdapter)(nil)

type symlinkAdapter struct{}

// NewSymlinkAdapter returns a new SymlinkAdapter.
func NewSymlinkAdapter() SymlinkAdapter {
	return &symlinkAdapter{}
}

// Create creates symlinks for all managed files in the composition.
// For each managed file it copies the workspace content into the CU repo,
// removes the original file, and creates a symlink pointing to the CU repo copy.
func (s *symlinkAdapter) Create(_ context.Context, cuRepoPath, workspacePath string, compo types.Compo) error {
	for _, repo := range compo.Repos {
		for _, file := range repo.ManagedFiles {
			workspaceFilePath := filepath.Join(workspacePath, repo.Name, file)
			cuFilePath := filepath.Join(cuRepoPath, repo.Path, file)

			content, err := os.ReadFile(workspaceFilePath)
			if err != nil {
				return fmt.Errorf("reading workspace file %s: %w", workspaceFilePath, err)
			}

			if err := os.MkdirAll(filepath.Dir(cuFilePath), 0o755); err != nil {
				return fmt.Errorf("creating CU repo directory for %s: %w", cuFilePath, err)
			}

			if err := os.WriteFile(cuFilePath, content, 0o644); err != nil {
				return fmt.Errorf("writing CU repo file %s: %w", cuFilePath, err)
			}

			if err := os.Remove(workspaceFilePath); err != nil {
				return fmt.Errorf("removing workspace file %s: %w", workspaceFilePath, err)
			}

			if err := os.Symlink(cuFilePath, workspaceFilePath); err != nil {
				return fmt.Errorf("creating symlink %s -> %s: %w", workspaceFilePath, cuFilePath, err)
			}
		}
	}

	return nil
}

// Remove removes all symlinks and restores original files by reading the content
// through the symlink before removing it.
func (s *symlinkAdapter) Remove(_ context.Context, workspacePath string, compo types.Compo) error {
	for _, repo := range compo.Repos {
		for _, file := range repo.ManagedFiles {
			workspaceFilePath := filepath.Join(workspacePath, repo.Name, file)

			content, err := os.ReadFile(workspaceFilePath)
			if err != nil {
				return fmt.Errorf("reading through symlink %s: %w", workspaceFilePath, err)
			}

			if err := os.Remove(workspaceFilePath); err != nil {
				return fmt.Errorf("removing symlink %s: %w", workspaceFilePath, err)
			}

			if err := os.WriteFile(workspaceFilePath, content, 0o644); err != nil {
				return fmt.Errorf("writing restored file %s: %w", workspaceFilePath, err)
			}
		}
	}

	return nil
}

// Verify checks that all expected symlinks exist and point to valid targets.
func (s *symlinkAdapter) Verify(_ context.Context, workspacePath string, compo types.Compo) (bool, error) {
	for _, repo := range compo.Repos {
		for _, file := range repo.ManagedFiles {
			workspaceFilePath := filepath.Join(workspacePath, repo.Name, file)

			fi, err := os.Lstat(workspaceFilePath)
			if err != nil {
				return false, nil
			}

			if fi.Mode()&os.ModeSymlink == 0 {
				return false, nil
			}

			target, err := os.Readlink(workspaceFilePath)
			if err != nil {
				return false, nil
			}

			if _, err := os.Stat(target); err != nil {
				return false, nil
			}
		}
	}

	return true, nil
}
