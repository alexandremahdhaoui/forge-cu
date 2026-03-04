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
	"os/exec"
	"path/filepath"
	"testing"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	// Create initial commit so we have a branch.
	writeFile(t, filepath.Join(dir, "README.md"), "# test")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "initial")
	return dir
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGitAdapter_Status_NoChanges(t *testing.T) {
	dir := initTestRepo(t)
	adapter := NewGitAdapter()
	ctx := context.Background()

	changes, err := adapter.Status(ctx, dir)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes, got %d: %v", len(changes), changes)
	}
}

func TestGitAdapter_Status_ModifiedFile(t *testing.T) {
	dir := initTestRepo(t)
	adapter := NewGitAdapter()
	ctx := context.Background()

	// Create and commit a file under a "forge" directory.
	writeFile(t, filepath.Join(dir, "forge", "go.mod"), "module test")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "add go.mod")

	// Modify the file.
	writeFile(t, filepath.Join(dir, "forge", "go.mod"), "module test v2")

	changes, err := adapter.Status(ctx, dir)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].RepoName != "forge" {
		t.Errorf("expected RepoName=forge, got %q", changes[0].RepoName)
	}
	if changes[0].File != "go.mod" {
		t.Errorf("expected File=go.mod, got %q", changes[0].File)
	}
	if changes[0].Status != "modified" {
		t.Errorf("expected Status=modified, got %q", changes[0].Status)
	}
}

func TestGitAdapter_Status_AddedFile(t *testing.T) {
	dir := initTestRepo(t)
	adapter := NewGitAdapter()
	ctx := context.Background()

	// Create an untracked file.
	writeFile(t, filepath.Join(dir, "forge", "go.sum"), "h1:abc123")

	changes, err := adapter.Status(ctx, dir)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Status != "added" {
		t.Errorf("expected Status=added, got %q", changes[0].Status)
	}
}

func TestGitAdapter_Commit(t *testing.T) {
	dir := initTestRepo(t)
	adapter := NewGitAdapter()
	ctx := context.Background()

	// Create a new file.
	writeFile(t, filepath.Join(dir, "forge", "go.mod"), "module test")

	// Commit using the adapter.
	if err := adapter.Commit(ctx, dir, "add forge/go.mod"); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	// Verify status is clean.
	changes, err := adapter.Status(ctx, dir)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes after commit, got %d: %v", len(changes), changes)
	}
}

func TestGitAdapter_ListBranches(t *testing.T) {
	dir := initTestRepo(t)
	adapter := NewGitAdapter()
	ctx := context.Background()

	branches, err := adapter.ListBranches(ctx, dir)
	if err != nil {
		t.Fatalf("ListBranches returned error: %v", err)
	}
	if len(branches) < 1 {
		t.Fatalf("expected at least 1 branch, got %d", len(branches))
	}
}

func TestGitAdapter_Checkout(t *testing.T) {
	dir := initTestRepo(t)
	adapter := NewGitAdapter()
	ctx := context.Background()

	// Create a branch manually.
	run(t, dir, "git", "branch", "test-branch")

	// Use the adapter to switch to the new branch.
	if err := adapter.Checkout(ctx, dir, "test-branch"); err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}

	// Verify we are on the correct branch.
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse failed: %v", err)
	}
	branch := string(out)
	// Trim trailing newline.
	branch = branch[:len(branch)-1]
	if branch != "test-branch" {
		t.Errorf("expected branch test-branch, got %q", branch)
	}
}
