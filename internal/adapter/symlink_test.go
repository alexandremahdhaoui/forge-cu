//go:build unit

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
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

func TestSymlinkAdapter_Create(t *testing.T) {
	workspaceDir := t.TempDir()
	cuRepoDir := t.TempDir()
	ctx := context.Background()

	// Create a workspace file.
	writeFile(t, filepath.Join(workspaceDir, "myrepo", "go.mod"), "module test")

	compo := types.Compo{
		Name: "test",
		Repos: []types.RepoEntry{{
			Name:         "myrepo",
			Path:         "myrepo",
			ManagedFiles: []string{"go.mod"},
		}},
	}

	adapter := NewSymlinkAdapter()
	if err := adapter.Create(ctx, cuRepoDir, workspaceDir, compo); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Verify workspace path is a symlink.
	workspaceFilePath := filepath.Join(workspaceDir, "myrepo", "go.mod")
	fi, err := os.Lstat(workspaceFilePath)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected workspace file to be a symlink")
	}

	// Verify CU repo file exists with correct content.
	cuFilePath := filepath.Join(cuRepoDir, "myrepo", "go.mod")
	content, err := os.ReadFile(cuFilePath)
	if err != nil {
		t.Fatalf("reading CU repo file: %v", err)
	}
	if string(content) != "module test" {
		t.Errorf("expected content %q, got %q", "module test", string(content))
	}

	// Verify symlink target matches CU repo path.
	target, err := os.Readlink(workspaceFilePath)
	if err != nil {
		t.Fatalf("Readlink failed: %v", err)
	}
	if target != cuFilePath {
		t.Errorf("expected symlink target %q, got %q", cuFilePath, target)
	}
}

func TestSymlinkAdapter_Verify_Valid(t *testing.T) {
	workspaceDir := t.TempDir()
	cuRepoDir := t.TempDir()
	ctx := context.Background()

	writeFile(t, filepath.Join(workspaceDir, "myrepo", "go.mod"), "module test")

	compo := types.Compo{
		Name: "test",
		Repos: []types.RepoEntry{{
			Name:         "myrepo",
			Path:         "myrepo",
			ManagedFiles: []string{"go.mod"},
		}},
	}

	adapter := NewSymlinkAdapter()
	if err := adapter.Create(ctx, cuRepoDir, workspaceDir, compo); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	valid, err := adapter.Verify(ctx, workspaceDir, compo)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if !valid {
		t.Fatal("expected Verify to return true after Create")
	}
}

func TestSymlinkAdapter_Verify_NoSymlink(t *testing.T) {
	workspaceDir := t.TempDir()
	ctx := context.Background()

	// Create a regular file (not a symlink).
	writeFile(t, filepath.Join(workspaceDir, "myrepo", "go.mod"), "module test")

	compo := types.Compo{
		Name: "test",
		Repos: []types.RepoEntry{{
			Name:         "myrepo",
			Path:         "myrepo",
			ManagedFiles: []string{"go.mod"},
		}},
	}

	adapter := NewSymlinkAdapter()
	valid, err := adapter.Verify(ctx, workspaceDir, compo)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if valid {
		t.Fatal("expected Verify to return false for regular file")
	}
}

func TestSymlinkAdapter_Remove(t *testing.T) {
	workspaceDir := t.TempDir()
	cuRepoDir := t.TempDir()
	ctx := context.Background()

	writeFile(t, filepath.Join(workspaceDir, "myrepo", "go.mod"), "module test")

	compo := types.Compo{
		Name: "test",
		Repos: []types.RepoEntry{{
			Name:         "myrepo",
			Path:         "myrepo",
			ManagedFiles: []string{"go.mod"},
		}},
	}

	adapter := NewSymlinkAdapter()
	if err := adapter.Create(ctx, cuRepoDir, workspaceDir, compo); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := adapter.Remove(ctx, workspaceDir, compo); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	// Verify workspace file is a regular file (not a symlink).
	workspaceFilePath := filepath.Join(workspaceDir, "myrepo", "go.mod")
	fi, err := os.Lstat(workspaceFilePath)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatal("expected workspace file to be a regular file after Remove")
	}

	// Verify content is preserved.
	content, err := os.ReadFile(workspaceFilePath)
	if err != nil {
		t.Fatalf("reading restored file: %v", err)
	}
	if string(content) != "module test" {
		t.Errorf("expected content %q, got %q", "module test", string(content))
	}
}
