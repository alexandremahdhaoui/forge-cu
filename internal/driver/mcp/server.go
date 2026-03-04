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
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/alexandremahdhaoui/forge-cu/internal/controller"
	"github.com/alexandremahdhaoui/forge-cu/internal/controller/engine"
	"github.com/alexandremahdhaoui/forge/pkg/mcpserver"
)

// RegisterTools registers all CU MCP tools on the given server.
func RegisterTools(
	server *mcpserver.Server,
	compoSvc controller.CompoService,
	commitSvc controller.CommitService,
	goEngine engine.GoCUEngine,
) {
	registerCommitTool(server, commitSvc)
	registerGoGetTool(server, goEngine)
	registerStatusTool(server, compoSvc)
	registerCheckoutTool(server, compoSvc)
	registerListBranchesTool(server, compoSvc)
}

// --- cu-commit ---

type commitInput struct {
	CURepoPath string `json:"cuRepoPath" jsonschema:"Path to CU repo"`
	Message    string `json:"message,omitempty" jsonschema:"Commit message (auto-generated if empty)"`
}

func registerCommitTool(server *mcpserver.Server, commitSvc controller.CommitService) {
	mcpserver.RegisterTool(server, &gomcp.Tool{
		Name:        "cu-commit",
		Description: "Stage and commit pending dependency changes in the CU repo.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input commitInput) (*gomcp.CallToolResult, any, error) {
		changes, hash, err := commitSvc.Commit(ctx, input.CURepoPath, input.Message)
		if err != nil {
			return errResult(err)
		}
		return jsonResult(map[string]any{"commitHash": hash, "changes": changes})
	})
}

// --- cu-go-get ---

type goGetInput struct {
	RepoDir    string `json:"repoDir" jsonschema:"Directory of the Go repository"`
	CURepoPath string `json:"cuRepoPath" jsonschema:"Path to CU repo"`
	Pkg        string `json:"pkg" jsonschema:"Go package to get"`
	Version    string `json:"version" jsonschema:"Package version"`
}

func registerGoGetTool(server *mcpserver.Server, goEngine engine.GoCUEngine) {
	mcpserver.RegisterTool(server, &gomcp.Tool{
		Name:        "cu-go-get",
		Description: "Run go get for a package and commit the resulting changes in the CU repo.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input goGetInput) (*gomcp.CallToolResult, any, error) {
		changes, hash, err := goEngine.GoGet(ctx, input.RepoDir, input.CURepoPath, input.Pkg, input.Version)
		if err != nil {
			return errResult(err)
		}
		return jsonResult(map[string]any{"commitHash": hash, "changes": changes})
	})
}

// --- cu-status ---

type statusInput struct {
	CURepoPath string `json:"cuRepoPath" jsonschema:"Path to CU repo"`
}

func registerStatusTool(server *mcpserver.Server, compoSvc controller.CompoService) {
	mcpserver.RegisterTool(server, &gomcp.Tool{
		Name:        "cu-status",
		Description: "Show pending dependency changes in the CU repo.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input statusInput) (*gomcp.CallToolResult, any, error) {
		changes, err := compoSvc.Status(ctx, input.CURepoPath)
		if err != nil {
			return errResult(err)
		}
		return jsonResult(map[string]any{"changes": changes})
	})
}

// --- cu-checkout ---

type checkoutInput struct {
	CURepoPath string `json:"cuRepoPath" jsonschema:"Path to CU repo"`
	Branch     string `json:"branch" jsonschema:"Branch name to check out"`
}

func registerCheckoutTool(server *mcpserver.Server, compoSvc controller.CompoService) {
	mcpserver.RegisterTool(server, &gomcp.Tool{
		Name:        "cu-checkout",
		Description: "Check out a branch in the CU repo.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input checkoutInput) (*gomcp.CallToolResult, any, error) {
		if err := compoSvc.Checkout(ctx, input.CURepoPath, input.Branch); err != nil {
			return errResult(err)
		}
		compo, err := compoSvc.LoadCompo(ctx, input.CURepoPath)
		if err != nil {
			return errResult(err)
		}
		return jsonResult(map[string]any{
			"branch": input.Branch,
			"repos":  compo.Repos,
		})
	})
}

// --- cu-list-branches ---

type listBranchesInput struct {
	CURepoPath string `json:"cuRepoPath" jsonschema:"Path to CU repo"`
}

func registerListBranchesTool(server *mcpserver.Server, compoSvc controller.CompoService) {
	mcpserver.RegisterTool(server, &gomcp.Tool{
		Name:        "cu-list-branches",
		Description: "List branches in the CU repo.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input listBranchesInput) (*gomcp.CallToolResult, any, error) {
		branches, err := compoSvc.ListBranches(ctx, input.CURepoPath)
		if err != nil {
			return errResult(err)
		}
		current, err := compoSvc.CurrentBranch(ctx, input.CURepoPath)
		if err != nil {
			return errResult(err)
		}
		return jsonResult(map[string]any{"branches": branches, "current": current})
	})
}

// --- helpers ---

func jsonResult(v any) (*gomcp.CallToolResult, any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("marshaling result: %w", err)
	}
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{
			&gomcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

func errResult(err error) (*gomcp.CallToolResult, any, error) {
	return nil, nil, err
}
