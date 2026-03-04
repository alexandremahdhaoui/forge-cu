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

package types

// Compo represents a loaded composition with resolved paths.
type Compo struct {
	Name  string      `json:"name"`
	Repos []RepoEntry `json:"repos"`
}

// RepoEntry represents one repository within a composition.
type RepoEntry struct {
	Name         string   `json:"name"`         // e.g., "forge", "forge-ui"
	Path         string   `json:"path"`         // Relative path in CU repo (e.g., "forge/")
	ManagedFiles []string `json:"managedFiles"` // e.g., ["go.mod", "go.sum"]
}

// DepChange represents a single file change detected in the CU repo.
type DepChange struct {
	RepoName string `json:"repoName"` // Which repo's file changed
	File     string `json:"file"`     // e.g., "go.mod"
	Status   string `json:"status"`   // "modified", "added", "deleted"
}
