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
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alexandremahdhaoui/forge-cu/internal/controller"
	"github.com/alexandremahdhaoui/forge-cu/internal/controller/engine"
)

func runCLI(compoSvc controller.CompoService, commitSvc controller.CommitService, goEngine engine.GoCUEngine) func() error {
	return func() error {
		if len(os.Args) < 2 {
			return fmt.Errorf("usage: forge-cu <command> [flags]\ncommands: status, commit, checkout, list-branches, go-get")
		}

		subcmd := os.Args[1]
		ctx := context.Background()

		switch subcmd {
		case "status":
			return runStatus(ctx, compoSvc)
		case "commit":
			return runCommit(ctx, commitSvc)
		case "checkout":
			return runCheckout(ctx, compoSvc)
		case "list-branches":
			return runListBranches(ctx, compoSvc)
		case "go-get":
			return runGoGet(ctx, goEngine)
		default:
			return fmt.Errorf("usage: forge-cu <command> [flags]\ncommands: status, commit, checkout, list-branches, go-get")
		}
	}
}

func runStatus(ctx context.Context, compoSvc controller.CompoService) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	cuRepoPath := fs.String("cu-repo-path", ".", "path to the CU repo")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	changes, err := compoSvc.Status(ctx, *cuRepoPath)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}

	fmt.Printf("%-20s %-20s %s\n", "REPO", "FILE", "STATUS")
	for _, c := range changes {
		fmt.Printf("%-20s %-20s %s\n", c.RepoName, c.File, c.Status)
	}

	return nil
}

func runCommit(ctx context.Context, commitSvc controller.CommitService) error {
	fs := flag.NewFlagSet("commit", flag.ContinueOnError)
	cuRepoPath := fs.String("cu-repo-path", ".", "path to the CU repo")
	message := fs.String("message", "", "commit message")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	changes, hash, err := commitSvc.Commit(ctx, *cuRepoPath, *message)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Printf("Commit: %s\n", hash)
	fmt.Printf("%-20s %-20s %s\n", "REPO", "FILE", "STATUS")
	for _, c := range changes {
		fmt.Printf("%-20s %-20s %s\n", c.RepoName, c.File, c.Status)
	}

	return nil
}

func runCheckout(ctx context.Context, compoSvc controller.CompoService) error {
	fs := flag.NewFlagSet("checkout", flag.ContinueOnError)
	cuRepoPath := fs.String("cu-repo-path", ".", "path to the CU repo")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	args := fs.Args()
	if len(args) < 1 {
		return fmt.Errorf("usage: forge-cu checkout [--cu-repo-path <path>] <branch>")
	}
	branch := args[0]

	if err := compoSvc.Checkout(ctx, *cuRepoPath, branch); err != nil {
		return fmt.Errorf("checkout: %w", err)
	}

	fmt.Printf("Checked out branch: %s\n", branch)

	return nil
}

func runListBranches(ctx context.Context, compoSvc controller.CompoService) error {
	fs := flag.NewFlagSet("list-branches", flag.ContinueOnError)
	cuRepoPath := fs.String("cu-repo-path", ".", "path to the CU repo")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	branches, err := compoSvc.ListBranches(ctx, *cuRepoPath)
	if err != nil {
		return fmt.Errorf("list-branches: %w", err)
	}

	for _, b := range branches {
		fmt.Println(b)
	}

	return nil
}

func runGoGet(ctx context.Context, goEngine engine.GoCUEngine) error {
	fs := flag.NewFlagSet("go-get", flag.ContinueOnError)
	cuRepoPath := fs.String("cu-repo-path", ".", "path to the CU repo")
	repoDir := fs.String("repo-dir", ".", "path to the repo directory")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	args := fs.Args()
	if len(args) < 1 {
		return fmt.Errorf("usage: forge-cu go-get [--cu-repo-path <path>] [--repo-dir <path>] <pkg>@<version>")
	}

	pkgVersion := args[0]
	parts := strings.SplitN(pkgVersion, "@", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: expected <pkg>@<version>, got %q", pkgVersion)
	}
	pkg, version := parts[0], parts[1]

	changes, hash, err := goEngine.GoGet(ctx, *repoDir, *cuRepoPath, pkg, version)
	if err != nil {
		return fmt.Errorf("go-get: %w", err)
	}

	fmt.Printf("Commit: %s\n", hash)
	fmt.Printf("%-20s %-20s %s\n", "REPO", "FILE", "STATUS")
	for _, c := range changes {
		fmt.Printf("%-20s %-20s %s\n", c.RepoName, c.File, c.Status)
	}

	return nil
}
