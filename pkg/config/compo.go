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
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CompoConfig is the schema for compo.yaml.
type CompoConfig struct {
	Name  string            `yaml:"name"  json:"name"`
	Repos []CompoRepoConfig `yaml:"repos" json:"repos"`
}

// CompoRepoConfig defines one repository entry in compo.yaml.
type CompoRepoConfig struct {
	Name         string   `yaml:"name"         json:"name"`
	URL          string   `yaml:"url"          json:"url"`
	ManagedFiles []string `yaml:"managedFiles" json:"managedFiles"`
}

// LoadCompoConfig reads and parses a compo.yaml file from the given path.
func LoadCompoConfig(path string) (*CompoConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading compo config %s: %w", path, err)
	}

	var cfg CompoConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing compo config %s: %w", path, err)
	}

	return &cfg, nil
}

// ValidateCompoConfig validates that a CompoConfig has all required fields
// and no duplicate repo names.
func ValidateCompoConfig(cfg *CompoConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("compo config: name is required")
	}

	if len(cfg.Repos) == 0 {
		return fmt.Errorf("compo config: at least one repo is required")
	}

	seen := make(map[string]bool)
	for i, repo := range cfg.Repos {
		if repo.Name == "" {
			return fmt.Errorf("compo config: repo[%d]: name is required", i)
		}
		if repo.URL == "" {
			return fmt.Errorf("compo config: repo[%d] (%s): url is required", i, repo.Name)
		}
		if seen[repo.Name] {
			return fmt.Errorf("compo config: duplicate repo name %q", repo.Name)
		}
		seen[repo.Name] = true
	}

	return nil
}
