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

package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

// --- Mock types ---

type mockGitAdapter struct {
	cloneFn             func(ctx context.Context, url, dest string) error
	checkoutFn          func(ctx context.Context, repoPath, branch string) error
	statusFn            func(ctx context.Context, repoPath string) ([]types.DepChange, error)
	commitFn            func(ctx context.Context, repoPath, message string) error
	pushFn              func(ctx context.Context, repoPath string) error
	pullFn              func(ctx context.Context, repoPath string) error
	listBranchesFn      func(ctx context.Context, repoPath string) ([]string, error)
	diffFn              func(ctx context.Context, repoPath string) (string, error)
	currentCommitHashFn func(ctx context.Context, repoPath string) (string, error)
	currentBranchFn     func(ctx context.Context, repoPath string) (string, error)
}

func (m *mockGitAdapter) Clone(ctx context.Context, url, dest string) error {
	return m.cloneFn(ctx, url, dest)
}

func (m *mockGitAdapter) Checkout(ctx context.Context, repoPath, branch string) error {
	return m.checkoutFn(ctx, repoPath, branch)
}

func (m *mockGitAdapter) Status(ctx context.Context, repoPath string) ([]types.DepChange, error) {
	return m.statusFn(ctx, repoPath)
}

func (m *mockGitAdapter) Commit(ctx context.Context, repoPath, message string) error {
	return m.commitFn(ctx, repoPath, message)
}

func (m *mockGitAdapter) Push(ctx context.Context, repoPath string) error {
	return m.pushFn(ctx, repoPath)
}

func (m *mockGitAdapter) Pull(ctx context.Context, repoPath string) error {
	return m.pullFn(ctx, repoPath)
}

func (m *mockGitAdapter) ListBranches(ctx context.Context, repoPath string) ([]string, error) {
	return m.listBranchesFn(ctx, repoPath)
}

func (m *mockGitAdapter) Diff(ctx context.Context, repoPath string) (string, error) {
	return m.diffFn(ctx, repoPath)
}

func (m *mockGitAdapter) CurrentCommitHash(ctx context.Context, repoPath string) (string, error) {
	if m.currentCommitHashFn != nil {
		return m.currentCommitHashFn(ctx, repoPath)
	}
	return "abc123", nil
}

func (m *mockGitAdapter) CurrentBranch(ctx context.Context, repoPath string) (string, error) {
	if m.currentBranchFn != nil {
		return m.currentBranchFn(ctx, repoPath)
	}
	return "main", nil
}

type mockSymlinkAdapter struct {
	createFn func(ctx context.Context, cuRepoPath, workspacePath string, compo types.Compo) error
	removeFn func(ctx context.Context, workspacePath string, compo types.Compo) error
	verifyFn func(ctx context.Context, workspacePath string, compo types.Compo) (bool, error)
}

func (m *mockSymlinkAdapter) Create(ctx context.Context, cuRepoPath, workspacePath string, compo types.Compo) error {
	return m.createFn(ctx, cuRepoPath, workspacePath, compo)
}

func (m *mockSymlinkAdapter) Remove(ctx context.Context, workspacePath string, compo types.Compo) error {
	return m.removeFn(ctx, workspacePath, compo)
}

func (m *mockSymlinkAdapter) Verify(ctx context.Context, workspacePath string, compo types.Compo) (bool, error) {
	return m.verifyFn(ctx, workspacePath, compo)
}

// --- Tests ---

func TestCompoService_Status_Delegates(t *testing.T) {
	expected := []types.DepChange{
		{RepoName: "forge", File: "go.mod", Status: "modified"},
		{RepoName: "forge", File: "go.sum", Status: "modified"},
	}

	git := &mockGitAdapter{
		statusFn: func(_ context.Context, _ string) ([]types.DepChange, error) {
			return expected, nil
		},
	}

	svc := NewCompoService(git, &mockSymlinkAdapter{})
	changes, err := svc.Status(context.Background(), "/tmp/cu-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changes) != len(expected) {
		t.Fatalf("expected %d changes, got %d", len(expected), len(changes))
	}

	for i, c := range changes {
		if c.RepoName != expected[i].RepoName || c.File != expected[i].File || c.Status != expected[i].Status {
			t.Errorf("change[%d]: expected %+v, got %+v", i, expected[i], c)
		}
	}
}

func TestCompoService_ListBranches_Delegates(t *testing.T) {
	expected := []string{"main", "feature/test", "release/v1"}

	git := &mockGitAdapter{
		listBranchesFn: func(_ context.Context, _ string) ([]string, error) {
			return expected, nil
		},
	}

	svc := NewCompoService(git, &mockSymlinkAdapter{})
	branches, err := svc.ListBranches(context.Background(), "/tmp/cu-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(branches) != len(expected) {
		t.Fatalf("expected %d branches, got %d", len(expected), len(branches))
	}

	for i, b := range branches {
		if b != expected[i] {
			t.Errorf("branch[%d]: expected %q, got %q", i, expected[i], b)
		}
	}
}

func TestCompoService_Checkout_Delegates(t *testing.T) {
	var calledBranch string

	git := &mockGitAdapter{
		checkoutFn: func(_ context.Context, _ string, branch string) error {
			calledBranch = branch
			return nil
		},
	}

	svc := NewCompoService(git, &mockSymlinkAdapter{})
	err := svc.Checkout(context.Background(), "/tmp/cu-repo", "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calledBranch != "feature/test" {
		t.Errorf("expected checkout branch %q, got %q", "feature/test", calledBranch)
	}
}

func TestCompoService_LoadCompo(t *testing.T) {
	tmpDir := t.TempDir()

	compoYAML := `name: test-compo
repos:
  - name: forge
    url: git@github.com:test/forge.git
    managedFiles:
      - go.mod
      - go.sum
`

	if err := os.WriteFile(filepath.Join(tmpDir, "compo.yaml"), []byte(compoYAML), 0o644); err != nil {
		t.Fatalf("writing compo.yaml: %v", err)
	}

	svc := NewCompoService(&mockGitAdapter{}, &mockSymlinkAdapter{})
	compo, err := svc.LoadCompo(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if compo.Name != "test-compo" {
		t.Errorf("expected name %q, got %q", "test-compo", compo.Name)
	}

	if len(compo.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(compo.Repos))
	}

	repo := compo.Repos[0]
	if repo.Name != "forge" {
		t.Errorf("expected repo name %q, got %q", "forge", repo.Name)
	}

	if repo.Path != "forge" {
		t.Errorf("expected repo path %q, got %q", "forge", repo.Path)
	}

	if len(repo.ManagedFiles) != 2 {
		t.Fatalf("expected 2 managed files, got %d", len(repo.ManagedFiles))
	}

	if repo.ManagedFiles[0] != "go.mod" {
		t.Errorf("expected managed file[0] %q, got %q", "go.mod", repo.ManagedFiles[0])
	}

	if repo.ManagedFiles[1] != "go.sum" {
		t.Errorf("expected managed file[1] %q, got %q", "go.sum", repo.ManagedFiles[1])
	}
}

func TestCompoService_Init(t *testing.T) {
	// Create a temp dir with compo.yaml for LoadCompo to read
	tmpDir := t.TempDir()
	cuRepoPath := filepath.Join(tmpDir, "cu-repo")

	// Track call order
	var callOrder []string

	git := &mockGitAdapter{
		cloneFn: func(_ context.Context, url, dest string) error {
			callOrder = append(callOrder, "clone")
			// Create the dest dir and compo.yaml so LoadCompo works
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
			compoYAML := `name: test-compo
repos:
  - name: forge
    url: git@github.com:test/forge.git
    managedFiles:
      - go.mod
`
			return os.WriteFile(filepath.Join(dest, "compo.yaml"), []byte(compoYAML), 0o644)
		},
		checkoutFn: func(_ context.Context, _, _ string) error {
			callOrder = append(callOrder, "checkout")
			return nil
		},
	}

	symlink := &mockSymlinkAdapter{
		createFn: func(_ context.Context, _, _ string, _ types.Compo) error {
			callOrder = append(callOrder, "symlink-create")
			return nil
		},
	}

	svc := NewCompoService(git, symlink)
	compo, err := svc.Init(context.Background(), "git@github.com:test/cu.git", cuRepoPath, "/workspace", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify call order
	if len(callOrder) != 3 {
		t.Fatalf("expected 3 calls, got %d: %v", len(callOrder), callOrder)
	}
	if callOrder[0] != "clone" || callOrder[1] != "checkout" || callOrder[2] != "symlink-create" {
		t.Errorf("expected call order [clone, checkout, symlink-create], got %v", callOrder)
	}

	// Verify returned compo
	if compo == nil {
		t.Fatal("expected non-nil compo")
	}
	if compo.Name != "test-compo" {
		t.Errorf("expected name %q, got %q", "test-compo", compo.Name)
	}
}
