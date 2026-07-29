# Kusabi (楔)

*Connect Repositories. Unify Contexts.*

A CLI that binds multiple Git repositories and surfaces declaration plus each child's documents as **read-only observations** in one context. Child content never enters the parent's Git history, and Kusabi does not take on Submodule friction.

## Install

### Homebrew (recommended)

```bash
brew tap novr/taps
brew install kusabi
```

Installs `kusabi` (`ksb` / `git-kusabi` are symlinks to the same binary). Release tarballs ship `kusabi` only.

### Go install

```bash
go install github.com/novr/kusabi/cmd/kusabi@latest
```

Short alias `ksb` and Git integration `git kusabi` are also available:

```bash
go install github.com/novr/kusabi/cmd/ksb@latest
go install github.com/novr/kusabi/cmd/git-kusabi@latest
```

## Release

Pushing a `v*` tag creates a GitHub Release and updates the [homebrew-taps](https://github.com/novr/homebrew-taps) formula.

```bash
git tag v0.1.0
git push origin v0.1.0
```

Manual runs: Actions workflow **Release Assets**.

**Prerequisites:** repository secrets `NOVRD_BOT_CLIENT_ID` / `NOVRD_BOT_KEY`, and org access to the `novr/homebrew-taps` reusable workflow ([new-tool.md](https://github.com/novr/homebrew-taps/blob/main/docs/new-tool.md)).

## Quick start

```bash
mkdir my-meta && cd my-meta
git init

kusabi init
kusabi add app-ios git@github.com:org/app-ios.git --role "iOS App"
kusabi add app-backend git@github.com:org/app-backend.git --role "API Server"
kusabi sync
kusabi doctor
kusabi context | pbcopy
```

## Layout

```plaintext
my-project-meta/
├── kusabi.yaml          # binding declaration (required)
├── AGENTS.md            # global policy (optional; observed by context)
├── .gitignore           # child paths excluded on init/add/remove
└── packages/            # child repositories (excluded from parent history)
    ├── app-ios/
    └── app-backend/
```

## Manifest (`kusabi.yaml`)

```yaml
version: "1"
name: "my-ecosystem"
description: "Cross-platform ecosystem bound by Kusabi."

context:
  agents: "./AGENTS.md"
  paths:
    - "team-knowledge/ADR.md"
  includes:
    - "README.md"
    - "CLAUDE.md"

repositories:
  app-ios:
    path: "packages/app-ios"
    url: "git@github.com:org/app-ios.git"
    branch: "develop"
    role: "iOS Client App (Swift / SwiftUI)"
    tags: ["frontend", "ios"]

  app-backend:
    path: "packages/app-backend"
    url: "git@github.com:org/app-backend.git"
    role: "Core API Server (Go / gRPC)"
    tags: ["backend", "api"]
    sync: false   # optional: exclude from kusabi sync
```

| Field | Meaning |
| :--- | :--- |
| `branch` | Branch to clone and track on sync (default: remote default) |
| `sync: false` | Skip `kusabi sync` (still included in `exec` / `context` / `status`) |
| `includes` | Per-child context paths (inherits parent `context.includes` when omitted) |

## Commands

| Command | Description |
| :--- | :--- |
| `init` | Initialize `kusabi.yaml`, `AGENTS.md`, `.gitignore` |
| `add` / `remove` | Add or remove declarations (`--branch` supported) |
| `sync` | Clone missing repos, align to declared branch, pull (`updated` / `updated: no change`; dirty or undeclared detached HEAD → skip) |
| `status` | Branch and working tree per child (`--json` for machine-readable output) |
| `exec` | Run a command across declared repos (`--repo` / `--tag` / `--skip-uncloned`) |
| `context` | Emit declaration and child documents as observations |
| `doctor` | Check declaration, clone, branch, remote (`--fix-remote` / `--migrate-gitignore`) |

Action commands exit non-zero when any repository **fails** (`Err`). **Skips** (dirty, detached, sync disabled, etc.) are warnings only and exit 0.

Binary: `kusabi` (short: `ksb`). With `git-kusabi` on `PATH`, `git kusabi` works too.

Design boundaries: [AGENTS.md](AGENTS.md).

## Development

```bash
go test ./...
go run ./cmd/kusabi --help
```

## License

[MIT](LICENSE)
