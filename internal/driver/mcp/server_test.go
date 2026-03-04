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

package mcp

import (
	"context"
	"testing"

	"github.com/alexandremahdhaoui/forge-cu/internal/controller"
	"github.com/alexandremahdhaoui/forge-cu/internal/controller/engine"
	"github.com/alexandremahdhaoui/forge-cu/internal/types"
	"github.com/alexandremahdhaoui/forge/pkg/mcpserver"
)

// Compile-time interface checks.
var (
	_ controller.CompoService  = (*mockCompoSvc)(nil)
	_ controller.CommitService = (*mockCommitSvc)(nil)
	_ engine.GoCUEngine        = (*mockGoEngine)(nil)
)

// mockCompoSvc implements controller.CompoService.
type mockCompoSvc struct{}

func (m *mockCompoSvc) Init(_ context.Context, _, _, _, _ string) (*types.Compo, error) {
	return nil, nil
}

func (m *mockCompoSvc) Status(_ context.Context, _ string) ([]types.DepChange, error) {
	return nil, nil
}

func (m *mockCompoSvc) ListBranches(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockCompoSvc) Checkout(_ context.Context, _, _ string) error { return nil }
func (m *mockCompoSvc) LoadCompo(_ context.Context, _ string) (*types.Compo, error) {
	return nil, nil
}

func (m *mockCompoSvc) CurrentBranch(_ context.Context, _ string) (string, error) {
	return "main", nil
}

// mockCommitSvc implements controller.CommitService.
type mockCommitSvc struct{}

func (m *mockCommitSvc) Commit(_ context.Context, _, _ string) ([]types.DepChange, string, error) {
	return nil, "", nil
}

// mockGoEngine implements engine.GoCUEngine.
type mockGoEngine struct{}

func (m *mockGoEngine) GoGet(_ context.Context, _, _, _, _ string) ([]types.DepChange, string, error) {
	return nil, "", nil
}

func TestRegisterTools_NoPanic(t *testing.T) {
	server := mcpserver.New("test-server", "0.0.0")

	compoSvc := &mockCompoSvc{}
	commitSvc := &mockCommitSvc{}
	goEngine := &mockGoEngine{}

	// RegisterTools should not panic with valid mock implementations.
	RegisterTools(server, compoSvc, commitSvc, goEngine)
}
