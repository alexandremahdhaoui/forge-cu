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
	"strings"
	"testing"

	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

func TestCommitService_Commit_WithChanges(t *testing.T) {
	expected := []types.DepChange{
		{RepoName: "forge", File: "go.mod", Status: "modified"},
	}

	var committedMessage string

	git := &mockGitAdapter{
		statusFn: func(_ context.Context, _ string) ([]types.DepChange, error) {
			return expected, nil
		},
		commitFn: func(_ context.Context, _, message string) error {
			committedMessage = message
			return nil
		},
		currentCommitHashFn: func(_ context.Context, _ string) (string, error) {
			return "abc123def456", nil
		},
	}

	svc := NewCommitService(git)
	changes, hash, err := svc.Commit(context.Background(), "/tmp/cu-repo", "test commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changes) != len(expected) {
		t.Fatalf("expected %d changes, got %d", len(expected), len(changes))
	}

	if changes[0].RepoName != "forge" || changes[0].File != "go.mod" {
		t.Errorf("unexpected change: %+v", changes[0])
	}

	if committedMessage != "test commit" {
		t.Errorf("expected message %q, got %q", "test commit", committedMessage)
	}

	if hash != "abc123def456" {
		t.Errorf("expected hash %q, got %q", "abc123def456", hash)
	}
}

func TestCommitService_Commit_NoChanges(t *testing.T) {
	git := &mockGitAdapter{
		statusFn: func(_ context.Context, _ string) ([]types.DepChange, error) {
			return nil, nil
		},
	}

	svc := NewCommitService(git)
	_, _, err := svc.Commit(context.Background(), "/tmp/cu-repo", "test commit")
	if err == nil {
		t.Fatal("expected error for no changes, got nil")
	}

	if !strings.Contains(err.Error(), "no pending changes") {
		t.Errorf("expected error containing %q, got %q", "no pending changes", err.Error())
	}
}

func TestCommitService_Commit_AutoMessage(t *testing.T) {
	changes := []types.DepChange{
		{RepoName: "forge", File: "go.mod", Status: "modified"},
	}

	var committedMessage string

	git := &mockGitAdapter{
		statusFn: func(_ context.Context, _ string) ([]types.DepChange, error) {
			return changes, nil
		},
		commitFn: func(_ context.Context, _, message string) error {
			committedMessage = message
			return nil
		},
	}

	svc := NewCommitService(git)
	_, _, err := svc.Commit(context.Background(), "/tmp/cu-repo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(committedMessage, "cu: update forge/go.mod") {
		t.Errorf("expected auto message containing %q, got %q", "cu: update forge/go.mod", committedMessage)
	}
}
