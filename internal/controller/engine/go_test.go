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

package engine

import (
	"context"
	"testing"

	"github.com/alexandremahdhaoui/forge-cu/internal/types"
)

type mockCommitService struct {
	commitFn func(ctx context.Context, cuRepoPath, message string) ([]types.DepChange, string, error)
}

func (m *mockCommitService) Commit(ctx context.Context, cuRepoPath, message string) ([]types.DepChange, string, error) {
	return m.commitFn(ctx, cuRepoPath, message)
}

func TestNewGoCUEngine(t *testing.T) {
	mock := &mockCommitService{
		commitFn: func(_ context.Context, _, _ string) ([]types.DepChange, string, error) {
			return nil, "", nil
		},
	}
	engine := NewGoCUEngine(mock)
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestGoCUEngine_GoGet_InvalidRepo(t *testing.T) {
	// GoGet should fail when running go get in a directory that's not a Go module.
	mock := &mockCommitService{
		commitFn: func(_ context.Context, _, _ string) ([]types.DepChange, string, error) {
			t.Error("commit should not be called when go get fails")
			return nil, "", nil
		},
	}

	e := NewGoCUEngine(mock)
	tmpDir := t.TempDir()

	_, _, err := e.GoGet(context.Background(), tmpDir, "/tmp/cu-repo", "example.com/fake", "v1.0.0")
	if err == nil {
		t.Fatal("expected error when running go get in non-module directory")
	}
}
