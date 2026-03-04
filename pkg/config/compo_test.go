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

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCompoConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "compo.yaml")

	content := `name: my-compo
repos:
  - name: forge
    url: https://github.com/example/forge
    managedFiles:
      - go.mod
      - go.sum
  - name: forge-ui
    url: https://github.com/example/forge-ui
    managedFiles:
      - package.json
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadCompoConfig(path)
	if err != nil {
		t.Fatalf("LoadCompoConfig returned error: %v", err)
	}

	if cfg.Name != "my-compo" {
		t.Errorf("expected Name=my-compo, got %q", cfg.Name)
	}
	if len(cfg.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0].Name != "forge" {
		t.Errorf("expected Repos[0].Name=forge, got %q", cfg.Repos[0].Name)
	}
	if cfg.Repos[0].URL != "https://github.com/example/forge" {
		t.Errorf("expected Repos[0].URL=https://github.com/example/forge, got %q", cfg.Repos[0].URL)
	}
	if len(cfg.Repos[0].ManagedFiles) != 2 {
		t.Errorf("expected 2 managed files for repos[0], got %d", len(cfg.Repos[0].ManagedFiles))
	}
	if cfg.Repos[1].Name != "forge-ui" {
		t.Errorf("expected Repos[1].Name=forge-ui, got %q", cfg.Repos[1].Name)
	}
}

func TestLoadCompoConfig_MissingFile(t *testing.T) {
	_, err := LoadCompoConfig("/nonexistent/path/compo.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadCompoConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "compo.yaml")

	content := `name: [invalid yaml: :::::`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadCompoConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestValidateCompoConfig_EmptyName(t *testing.T) {
	cfg := &CompoConfig{
		Name: "",
		Repos: []CompoRepoConfig{
			{Name: "forge", URL: "https://example.com/forge"},
		},
	}

	err := ValidateCompoConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("expected error to contain 'name is required', got %q", err.Error())
	}
}

func TestValidateCompoConfig_NoRepos(t *testing.T) {
	cfg := &CompoConfig{
		Name:  "my-compo",
		Repos: nil,
	}

	err := ValidateCompoConfig(cfg)
	if err == nil {
		t.Fatal("expected error for no repos, got nil")
	}
	if !strings.Contains(err.Error(), "at least one repo") {
		t.Errorf("expected error to contain 'at least one repo', got %q", err.Error())
	}
}

func TestValidateCompoConfig_DuplicateRepos(t *testing.T) {
	cfg := &CompoConfig{
		Name: "my-compo",
		Repos: []CompoRepoConfig{
			{Name: "forge", URL: "https://example.com/forge"},
			{Name: "forge", URL: "https://example.com/forge2"},
		},
	}

	err := ValidateCompoConfig(cfg)
	if err == nil {
		t.Fatal("expected error for duplicate repos, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected error to contain 'duplicate', got %q", err.Error())
	}
}

func TestValidateCompoConfig_MissingRepoName(t *testing.T) {
	cfg := &CompoConfig{
		Name: "my-compo",
		Repos: []CompoRepoConfig{
			{Name: "", URL: "https://example.com/forge"},
		},
	}

	err := ValidateCompoConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing repo name, got nil")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("expected error to contain 'name is required', got %q", err.Error())
	}
}

func TestValidateCompoConfig_MissingRepoURL(t *testing.T) {
	cfg := &CompoConfig{
		Name: "my-compo",
		Repos: []CompoRepoConfig{
			{Name: "forge", URL: ""},
		},
	}

	err := ValidateCompoConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing repo URL, got nil")
	}
	if !strings.Contains(err.Error(), "url is required") {
		t.Errorf("expected error to contain 'url is required', got %q", err.Error())
	}
}

func TestValidateCompoConfig_Valid(t *testing.T) {
	cfg := &CompoConfig{
		Name: "my-compo",
		Repos: []CompoRepoConfig{
			{Name: "forge", URL: "https://example.com/forge"},
			{Name: "forge-ui", URL: "https://example.com/forge-ui"},
		},
	}

	err := ValidateCompoConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
