# forge-cu Design

**forge-cu manages shared Go dependencies across multi-repository workspaces through symlinked dependency files stored in a dedicated git repository.**

## Problem Statement

Go workspaces contain 2+ repositories with separate go.mod and go.sum files.
Shared dependencies appear in each file independently.
Updating a shared dependency requires editing go.mod in every repository that uses it.
Developers forget repositories, introduce version drift, and break builds.
No standard tool coordinates go.mod changes across repository boundaries.

forge-cu answers: how do you update shared dependencies atomically across all repositories in a workspace?

It stores all go.mod and go.sum files in a single "CU repo" (Continuous Update repository).
Symlinks connect workspace repositories to the CU repo's copies.
Git branches in the CU repo represent dependency compositions.
Switching branches swaps all dependency files in one operation.

## Tenets

Listed in priority order. When tenets conflict, higher-ranked tenets win.

1. **Atomic composition switching.** Switching a branch replaces all dependency files at once. Partial updates do not exist.
2. **Transparency to the Go toolchain.** `go build`, `go test`, and `go mod tidy` work without modification. Symlinks are invisible to Go.
3. **Git as the source of truth.** The CU repo is a standard git repository. Branches, commits, and merges follow git semantics. No custom storage.
4. **Minimal surface area.** 5 operations: status, commit, go-get, checkout, list-branches. No operation requires more than 3 parameters.
5. **Dual execution modes.** Every operation runs as a CLI subcommand or as an MCP tool. Same logic, different entry points.

## Requirements

1. Store go.mod and go.sum for N repositories in a single git repository.
2. Create symlinks from workspace repos to the CU repo on initialization.
3. Detect uncommitted dependency changes via git status.
4. Commit dependency changes with auto-generated or user-provided messages.
5. Run `go get` + `go mod tidy` in a workspace repo and auto-commit results to the CU repo.
6. Switch between composition branches via git checkout.
7. List available composition branches.
8. Expose all operations as MCP tools (JSON-RPC 2.0 over stdio).
9. Expose all operations as CLI subcommands.
10. Parse and validate `compo.yaml` configuration.

## Out of Scope

- Remote git push/pull automation (users run git push manually).
- Conflict resolution when merging CU repo branches.
- Dependency version analysis or upgrade recommendations.
- Support for non-Go dependency files (build.gradle, package.json). The managed files list is generic, but testing targets Go only.
- CI/CD pipeline integration.

## Success Criteria

| Criterion | Target |
|-----------|--------|
| MCP tools registered | 5 |
| CLI subcommands | 5 |
| Test stages pass | 4 (lint-tags, lint-license, lint, unit) |
| Configuration fields validated | 4 (name, repo name, repo url, no duplicates) |
| Adapter interfaces | 2 (GitAdapter: 10 methods, SymlinkAdapter: 3 methods) |
| Controller interfaces | 3 (CompoService: 6 methods, CommitService: 1 method, GoCUEngine: 1 method) |

## Proposed Design

### Symlink Mechanism

```
  Workspace                         CU Repo
  +---------------------------+     +---------------------------+
  | forge/                    |     | compo.yaml                |
  |   go.mod ---symlink------------>| forge/go.mod              |
  |   go.sum ---symlink------------>| forge/go.sum              |
  |                           |     |                           |
  | forge-ui/                 |     | forge-ui/go.mod           |
  |   go.mod ---symlink------------>| forge-ui/go.sum           |
  |   go.sum ---symlink------------>|                           |
  +---------------------------+     +---------------------------+
```

Initialization flow:
1. Clone the CU repo to a local path.
2. Checkout the target branch.
3. Read `compo.yaml` to discover repos and managed files.
4. For each managed file: copy workspace content to CU repo, delete the original, create a symlink.

After initialization, the Go toolchain reads go.mod/go.sum through symlinks. All reads and writes operate on the CU repo's files.

### CU Repo Structure

```
  compo-cu-repo/
  |-- compo.yaml              # Composition configuration
  |-- forge/
  |   |-- go.mod              # forge's dependency file
  |   +-- go.sum
  |-- forge-ui/
  |   |-- go.mod              # forge-ui's dependency file
  |   +-- go.sum
  +-- forge-cu/
      |-- go.mod
      +-- go.sum
```

Each repository maps to a subdirectory named by `repos[].name` in `compo.yaml`.

### Branch Switching Workflow

```
  User                    forge-cu             CU Repo (git)
  |                       |                    |
  | checkout "experimental"|                    |
  |----------------------->|                    |
  |                       | git checkout       |
  |                       |------------------->|
  |                       |    files replaced  |
  |                       |<-------------------|
  |                       |                    |
  |                       | load compo.yaml    |
  |                       |------------------->|
  |                       |    config          |
  |                       |<-------------------|
  |                       |                    |
  | branch + repos info   |                    |
  |<-----------------------|                    |
```

Because workspace symlinks point to CU repo files, `git checkout` in the CU repo replaces all dependency files atomically. No symlink recreation is needed.

### go-get Workflow

```
  User                    GoCUEngine       CommitService      Git
  |                       |                |                  |
  | go-get pkg@v          |                |                  |
  |---------------------->|                |                  |
  |                       | exec: go get   |                  |
  |                       |--------------->|                  |
  |                       | exec: go mod tidy                 |
  |                       |--------------->|                  |
  |                       |                |                  |
  |                       | commit(msg)    |                  |
  |                       |--------------->|                  |
  |                       |                | git status       |
  |                       |                |----------------->|
  |                       |                | git add .        |
  |                       |                |----------------->|
  |                       |                | git commit       |
  |                       |                |----------------->|
  |                       |                | rev-parse HEAD   |
  |                       |                |----------------->|
  |                       |                |    hash          |
  |                       |                |<-----------------|
  |                       |  changes, hash |                  |
  |                       |<---------------|                  |
  | changes, hash         |                |                  |
  |<----------------------|                |                  |
```

`go get` and `go mod tidy` run in the workspace repo directory. Because go.mod is a symlink to the CU repo, changes land in the CU repo automatically. CommitService then stages and commits those changes.

### Execution Modes

```
  +------------------+     +------------------+
  |   CLI Mode       |     |   MCP Mode       |
  | forge-cu <cmd>   |     | forge-cu --mcp   |
  +--------+---------+     +--------+---------+
           |                        |
           v                        v
  +--------+---------+     +--------+---------+
  | runCLI()         |     | runMCPServer()   |
  | flag parsing     |     | JSON-RPC 2.0     |
  +--------+---------+     +--------+---------+
           |                        |
           +----------+  +----------+
                      |  |
                      v  v
           +----------+--+---------+
           |  Controller Layer     |
           | CompoService         |
           | CommitService        |
           | GoCUEngine           |
           +----------+-----------+
                      |
                      v
           +----------+-----------+
           |  Adapter Layer       |
           | GitAdapter (os/exec) |
           | SymlinkAdapter (os)  |
           +-----------------------+
```

`enginecli.Bootstrap` from the forge SDK dispatches to `runCLI()` or `runMCPServer()` based on the `--mcp` flag. Both modes share the same controller and adapter instances.

## Technical Design

### Data Model

```go
// pkg/config/compo.go -- exported configuration schema
type CompoConfig struct {
    Name  string            `yaml:"name"`
    Repos []CompoRepoConfig `yaml:"repos"`
}

type CompoRepoConfig struct {
    Name         string   `yaml:"name"`
    URL          string   `yaml:"url"`
    ManagedFiles []string `yaml:"managedFiles"`
}

// internal/types/types.go -- internal domain types
type Compo struct {
    Name  string
    Repos []RepoEntry
}

type RepoEntry struct {
    Name         string   // "forge"
    Path         string   // "forge/" (relative path in CU repo)
    ManagedFiles []string // ["go.mod", "go.sum"]
}

type DepChange struct {
    RepoName string // "forge"
    File     string // "go.mod"
    Status   string // "modified", "added", "deleted"
}
```

`CompoConfig` is the exported YAML schema. `Compo` is the internal resolved representation. `DepChange` is the universal change type returned by status, commit, and go-get operations.

### MCP Tools

5 tools registered via `mcpserver.RegisterTool`:

| Tool | Input Type | Controller | Returns |
|------|-----------|------------|---------|
| `cu-status` | `{cuRepoPath}` | CompoService.Status | `{changes: []DepChange}` |
| `cu-commit` | `{cuRepoPath, message?}` | CommitService.Commit | `{commitHash, changes}` |
| `cu-go-get` | `{repoDir, cuRepoPath, pkg, version}` | GoCUEngine.GoGet | `{commitHash, changes}` |
| `cu-checkout` | `{cuRepoPath, branch}` | CompoService.Checkout | `{branch, repos}` |
| `cu-list-branches` | `{cuRepoPath}` | CompoService.ListBranches | `{branches, current}` |

All tools return JSON via `gomcp.TextContent`. Errors propagate as MCP error responses.

### Package Catalog

**Exported packages:**

| Package | Purpose |
|---------|---------|
| `pkg/config` | CompoConfig schema, YAML loading, validation |

**Internal packages:**

| Package | Purpose |
|---------|---------|
| `internal/adapter` | GitAdapter and SymlinkAdapter interfaces + implementations |
| `internal/controller` | CompoService and CommitService business logic |
| `internal/controller/engine` | GoCUEngine -- Go-specific dependency operations |
| `internal/driver/mcp` | MCP tool registrations |
| `internal/types` | Compo, RepoEntry, DepChange domain types |
| `internal/util/mocks/mockadapter` | Generated mocks for adapter interfaces |
| `internal/util/mocks/mockcontroller` | Generated mocks for controller interfaces |
| `internal/util/mocks/mockengine` | Generated mocks for engine interfaces |

### Adapter Interfaces

**GitAdapter** (10 methods): Clone, Checkout, Status, Commit, Push, Pull, ListBranches, Diff, CurrentCommitHash, CurrentBranch. All execute git via `os/exec`. Auth relies on the environment (SSH agent, git credential helper).

**SymlinkAdapter** (3 methods): Create, Remove, Verify. Create copies workspace files into the CU repo, deletes originals, and creates symlinks. Remove reads through symlinks, deletes them, and writes regular files. Verify checks symlink existence and target validity via `os.Lstat` and `os.Readlink`.

## Design Patterns

**Ports and Adapters.** Controller interfaces define ports (CompoService, CommitService, GoCUEngine). Adapter interfaces abstract external systems (git, filesystem). The MCP driver and CLI are two entry points into the same controller layer.

**Compile-time interface checks.** Every concrete type includes `var _ Interface = (*impl)(nil)` to catch interface drift at compile time.

**Auto-generated commit messages.** When CommitService receives an empty message, it generates one from the list of changed files (e.g., `cu: update forge/go.mod, forge/go.sum`).

## Alternatives Considered

**Do nothing (manual coordination).** Developers edit go.mod in each repository independently. Rejected because version drift grows with repository count. A workspace with 4 repositories and 12 shared dependencies requires 48 coordinated edits per dependency update.

**Monorepo.** Merge all repositories into one. Rejected because monorepo migration is disruptive for established projects. forge-cu preserves multi-repo structure.

**go.work replace directives.** Use `go.work` replace to point all repositories at local modules. Rejected because replace directives do not manage upstream dependency versions. They solve local development, not dependency coordination.

**Vendoring per repo.** Each repository vendors its dependencies. Rejected because vendoring duplicates dependency source code N times and does not coordinate versions.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Symlink breaks if CU repo moves | All go.mod reads fail | Verify method checks symlink validity; re-run init to recreate |
| Uncommitted CU repo changes lost on branch switch | Dependency edits disappear | Status check before checkout; git warns on dirty working tree |
| Concurrent `go get` in 2 repos | Race condition in CU repo commits | Single-user workflow; CU repo is local-only during development |
| CU repo merge conflicts | Branch merge fails | Standard git merge resolution; compo.yaml changes are rare |
| Git not installed | All operations fail | GitAdapter surfaces clear error from os/exec |

## Testing Strategy

4 test stages configured in `forge.yaml`:

| Stage | Runner | What it validates |
|-------|--------|-------------------|
| `lint-tags` | `go://go-lint-tags` | Build tags present on test files |
| `lint-license` | `go://go-lint-licenses` | Apache 2.0 headers on all Go files |
| `lint` | `go://go-lint` | golangci-lint v2 with `-tags=unit` |
| `unit` | `go://go-test` | Unit tests with mockery-generated mocks |

Mock generation uses mockery with the testify template. 3 mock packages cover all adapter, controller, and engine interfaces.

Test files:
- `internal/adapter/git_test.go`
- `internal/adapter/symlink_test.go`
- `internal/controller/compo_test.go`
- `internal/controller/commit_test.go`
- `internal/controller/engine/go_test.go`
- `internal/driver/mcp/server_test.go`
- `pkg/config/compo_test.go`

## FAQ

**Why a separate git repository instead of a branch in each workspace repo?**
A single CU repo provides one commit history for all dependency changes. Correlating changes across repositories requires one `git log`, not N. Branch switching in one repository replaces all files at once.

**Why symlinks instead of file copies?**
Symlinks make Go toolchain writes (from `go get` or `go mod tidy`) land directly in the CU repo. File copies would require a sync step after every dependency change.

**Why does GitAdapter shell out to git instead of using a Go git library?**
Shelling out to `git` via `os/exec` reuses the user's existing git configuration, SSH keys, and credential helpers. A library-based approach would need to replicate that authentication surface.

**Can forge-cu manage non-Go dependency files?**
Yes. The `managedFiles` list accepts any filename. The symlink mechanism is language-agnostic. Only `cu-go-get` is Go-specific.

## Appendix: compo.yaml Example

```yaml
name: forge-ai-compo
repos:
  - name: forge
    url: git@github.com:alexandremahdhaoui/forge.git
    managedFiles: [go.mod, go.sum]
  - name: forge-ui
    url: git@github.com:alexandremahdhaoui/forge-ui.git
    managedFiles: [go.mod, go.sum]
  - name: forge-cu
    url: git@github.com:alexandremahdhaoui/forge-cu.git
    managedFiles: [go.mod, go.sum]
```
