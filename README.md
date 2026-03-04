# forge-cu

**Continuous Update system for dependency management across multi-repository Go workspaces.**

> "I maintain 4 Go repositories that share 12 dependencies.
> Updating one dependency means editing go.mod in each repo, hoping versions stay aligned.
> forge-cu stores all go.mod and go.sum files in a single git repo and symlinks them into my workspace.
> Switching a branch swaps all dependency files atomically."

## What problem does forge-cu solve?

Go workspaces with 2+ repositories share dependencies declared in separate go.mod files.
Updating a shared dependency requires coordinated edits across all repositories.
Manual coordination causes version drift, broken builds, and wasted developer time.
forge-cu solves this by storing dependency files (go.mod, go.sum) in a dedicated git repository -- the "CU repo."
Symlinks connect each workspace repository to the CU repo's copies.
Git branches in the CU repo represent dependency compositions; switching branches swaps all dependency files atomically.

## Quick Start

```bash
# Build forge-cu
forge build forge-cu

# Initialize: clone CU repo, create symlinks
forge-cu init --cu-repo-url git@github.com:org/my-compo.git \
              --cu-repo-path ./compo-cu \
              --workspace-path . \
              --branch main

# Check for uncommitted dependency changes
forge-cu status --cu-repo-path ./compo-cu

# Update a dependency across the workspace
forge-cu go-get --cu-repo-path ./compo-cu \
                --repo-dir ./forge \
                golang.org/x/net@v0.30.0

# Commit changes
forge-cu commit --cu-repo-path ./compo-cu --message "bump x/net"

# Switch dependency composition
forge-cu checkout --cu-repo-path ./compo-cu experimental
```

## How does it work?

```
  Workspace (go.work)             CU Repo (git)
  +---------------------------+   +---------------------------+
  | forge/                    |   | compo.yaml                |
  |   go.mod ---symlink------>|-->| forge/go.mod              |
  |   go.sum ---symlink------>|-->| forge/go.sum              |
  |                           |   |                           |
  | forge-ui/                 |   | forge-ui/go.mod           |
  |   go.mod ---symlink------>|-->| forge-ui/go.sum           |
  |   go.sum ---symlink------>|-->|                           |
  +---------------------------+   +---------------------------+
                                        |
                                  git branch = composition
                                        |
                                  main, experimental, v2-deps
```

forge-cu creates symlinks from each workspace repo's go.mod/go.sum to files in the CU repo.
The Go toolchain reads through symlinks transparently.
Each git branch in the CU repo holds a complete set of dependency files -- a "composition."
Switching branches replaces all dependency files in one atomic git checkout.

See [DESIGN.md](DESIGN.md) for the full technical design.

## Table of Contents

- [How do I configure a CU repo?](#how-do-i-configure-a-cu-repo)
- [How do I build and test?](#how-do-i-build-and-test)
- [What MCP tools are available?](#what-mcp-tools-are-available)
- [What CLI commands are available?](#what-cli-commands-are-available)
- [FAQ](#faq)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## How do I configure a CU repo?

Create a `compo.yaml` in the root of your CU repo:

```yaml
name: my-workspace-compo
repos:
  - name: forge
    url: git@github.com:org/forge.git
    managedFiles: [go.mod, go.sum]
  - name: forge-ui
    url: git@github.com:org/forge-ui.git
    managedFiles: [go.mod, go.sum]
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Composition name. Must be unique. |
| `repos[].name` | Yes | Repository name. Maps to a subdirectory in the CU repo. |
| `repos[].url` | Yes | Git URL for the repository. |
| `repos[].managedFiles` | Yes | Files to symlink (typically `go.mod` and `go.sum`). |

## How do I build and test?

```bash
# Build the binary
forge build forge-cu

# Run all tests (lint + unit)
forge test-all

# Run individual test stages
forge test run lint        # golangci-lint
forge test run unit        # go test with -tags=unit
forge test run lint-tags   # verify build tags
forge test run lint-license # verify license headers
```

Build targets: `format-code`, `forge-cu`, `generated-mocks`.
Test stages: `lint-tags`, `lint-license`, `lint`, `unit`.

## What MCP tools are available?

Start the MCP server with `forge-cu --mcp`. 5 tools register automatically:

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `cu-status` | Show uncommitted dependency changes | `cuRepoPath` |
| `cu-commit` | Stage and commit pending changes | `cuRepoPath`, `message` (optional) |
| `cu-go-get` | Run `go get` + `go mod tidy`, auto-commit | `repoDir`, `cuRepoPath`, `pkg`, `version` |
| `cu-checkout` | Switch composition branch | `cuRepoPath`, `branch` |
| `cu-list-branches` | List composition branches | `cuRepoPath` |

## What CLI commands are available?

Run `forge-cu <command> [flags]` directly:

| Command | Description | Flags |
|---------|-------------|-------|
| `status` | Show pending dependency changes | `--cu-repo-path` |
| `commit` | Commit pending changes | `--cu-repo-path`, `--message` |
| `go-get` | Run `go get`, commit result | `--cu-repo-path`, `--repo-dir`, positional: `<pkg>@<version>` |
| `checkout` | Switch composition branch | `--cu-repo-path`, positional: `<branch>` |
| `list-branches` | List composition branches | `--cu-repo-path` |

All commands default `--cu-repo-path` to `.` (current directory).

## FAQ

**How does forge-cu differ from `go.work`?**
`go.work` manages local module replacements. forge-cu manages go.mod/go.sum files across repositories. They complement each other: `go.work` defines the workspace; forge-cu keeps dependency versions synchronized.

**What happens if I edit go.mod directly?**
The symlink points to the CU repo. Direct edits modify the CU repo's copy. Run `forge-cu status` to see changes, then `forge-cu commit` to persist them.

**Can I manage files other than go.mod and go.sum?**
Yes. List any file in `managedFiles`. forge-cu symlinks whatever files you specify.

**How do I create a new composition branch?**
Create a branch in the CU repo with standard git: `cd compo-cu && git checkout -b experimental`. Then switch to it with `forge-cu checkout`.

**What Go version does forge-cu require?**
Go 1.25.7 or later.

**How does authentication work for the CU repo?**
forge-cu uses your environment's git credentials (SSH agent, credential helper). No separate auth configuration exists.

## Documentation

| Document | Audience | Description |
|----------|----------|-------------|
| [README.md](README.md) | Users | Quick start and command reference |
| [DESIGN.md](DESIGN.md) | Developers | Architecture, data model, design decisions |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributors | Build, test, commit conventions |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for build instructions, commit conventions, and project structure.

## License

Apache License 2.0. See [LICENSE](LICENSE).
