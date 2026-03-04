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
	"path/filepath"

	"github.com/alexandremahdhaoui/forge-cu/internal/adapter"
	"github.com/alexandremahdhaoui/forge-cu/internal/types"
	"github.com/alexandremahdhaoui/forge-cu/pkg/config"
)

// CompoService manages composition lifecycle and state queries.
type CompoService interface {
	Init(ctx context.Context, cuRepoURL, cuRepoPath, workspacePath, branch string) (*types.Compo, error)
	Status(ctx context.Context, cuRepoPath string) ([]types.DepChange, error)
	ListBranches(ctx context.Context, cuRepoPath string) ([]string, error)
	Checkout(ctx context.Context, cuRepoPath, branch string) error
	LoadCompo(ctx context.Context, cuRepoPath string) (*types.Compo, error)
	CurrentBranch(ctx context.Context, cuRepoPath string) (string, error)
}

// Compile-time interface check.
var _ CompoService = (*compoService)(nil)

type compoService struct {
	git     adapter.GitAdapter
	symlink adapter.SymlinkAdapter
}

// NewCompoService creates a CompoService with the given adapters.
func NewCompoService(git adapter.GitAdapter, symlink adapter.SymlinkAdapter) CompoService {
	return &compoService{git: git, symlink: symlink}
}

func (c *compoService) Init(ctx context.Context, cuRepoURL, cuRepoPath, workspacePath, branch string) (*types.Compo, error) {
	if err := c.git.Clone(ctx, cuRepoURL, cuRepoPath); err != nil {
		return nil, fmt.Errorf("cloning CU repo: %w", err)
	}

	if err := c.git.Checkout(ctx, cuRepoPath, branch); err != nil {
		return nil, fmt.Errorf("checking out branch %s: %w", branch, err)
	}

	compo, err := c.LoadCompo(ctx, cuRepoPath)
	if err != nil {
		return nil, fmt.Errorf("loading compo config: %w", err)
	}

	if err := c.symlink.Create(ctx, cuRepoPath, workspacePath, *compo); err != nil {
		return nil, fmt.Errorf("creating symlinks: %w", err)
	}

	return compo, nil
}

func (c *compoService) Status(ctx context.Context, cuRepoPath string) ([]types.DepChange, error) {
	return c.git.Status(ctx, cuRepoPath)
}

func (c *compoService) ListBranches(ctx context.Context, cuRepoPath string) ([]string, error) {
	return c.git.ListBranches(ctx, cuRepoPath)
}

func (c *compoService) Checkout(ctx context.Context, cuRepoPath, branch string) error {
	return c.git.Checkout(ctx, cuRepoPath, branch)
}

func (c *compoService) CurrentBranch(ctx context.Context, cuRepoPath string) (string, error) {
	return c.git.CurrentBranch(ctx, cuRepoPath)
}

func (c *compoService) LoadCompo(ctx context.Context, cuRepoPath string) (*types.Compo, error) {
	configPath := filepath.Join(cuRepoPath, "compo.yaml")

	cfg, err := config.LoadCompoConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading compo config: %w", err)
	}

	if err := config.ValidateCompoConfig(cfg); err != nil {
		return nil, fmt.Errorf("validating compo config: %w", err)
	}

	compo := types.Compo{
		Name: cfg.Name,
	}

	for _, repo := range cfg.Repos {
		compo.Repos = append(compo.Repos, types.RepoEntry{
			Name:         repo.Name,
			Path:         repo.Name,
			ManagedFiles: repo.ManagedFiles,
		})
	}

	return &compo, nil
}
