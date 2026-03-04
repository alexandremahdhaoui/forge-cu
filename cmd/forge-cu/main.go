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

package main

import (
	"github.com/alexandremahdhaoui/forge-cu/internal/adapter"
	"github.com/alexandremahdhaoui/forge-cu/internal/controller"
	"github.com/alexandremahdhaoui/forge-cu/internal/controller/engine"
	"github.com/alexandremahdhaoui/forge/pkg/enginecli"
)

// Version is set via ldflags at build time.
var Version = "dev"

func main() {
	// Create adapters.
	git := adapter.NewGitAdapter()
	symlink := adapter.NewSymlinkAdapter()

	// Create controllers.
	compoSvc := controller.NewCompoService(git, symlink)
	commitSvc := controller.NewCommitService(git)
	goEngine := engine.NewGoCUEngine(commitSvc)

	enginecli.Bootstrap(enginecli.Config{
		Name:    "forge-cu",
		Version: Version,
		RunMCP:  runMCPServer(compoSvc, commitSvc, goEngine),
		RunCLI:  runCLI(compoSvc, commitSvc, goEngine),
	})
}
