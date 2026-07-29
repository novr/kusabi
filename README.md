# Kusabi (楔)

*Connect Repositories. Unify Contexts.*

## Install

```bash
brew tap novr/taps && brew install kusabi
# or
go install github.com/novr/kusabi/cmd/kusabi@latest
```

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

## Manifest (`kusabi.yaml`)

```yaml
version: "1"
name: "my-ecosystem"

context:
  agents: "./AGENTS.md"
  paths:
    - "team-knowledge/ADR.md"
  includes:
    - "README.md"

repositories:
  app-ios:
    path: "packages/app-ios"
    url: "git@github.com:org/app-ios.git"
    branch: "develop"
    role: "iOS Client App"
    tags: ["frontend", "ios"]

  app-backend:
    path: "packages/app-backend"
    url: "git@github.com:org/app-backend.git"
    sync: false
```

| Field | |
| :--- | :--- |
| `branch` | Clone and track on sync (remote default when omitted) |
| `sync: false` | Excluded from `sync` only |
| `context.paths` | Parent files included in `context` output |
| `includes` | Per-child context paths; inherits parent `context.includes` when omitted |

## Commands

| Command | |
| :--- | :--- |
| `init [--force]` | Scaffold `kusabi.yaml`, `AGENTS.md`, `.gitignore` |
| `add` / `remove` | Edit declaration (`--path`, `--role`, `--tags`, `--branch`) |
| `sync [--depth]` | Clone or pull; align to declared branch |
| `status [--json]` | Branch and working tree per child |
| `exec` | Shell command across children (`--repo`, `--tag`, `--skip-uncloned`) |
| `context [--tree] [--json]` | Observed context to stdout |
| `doctor` | Health checks (`--fix-remote`, `--migrate-gitignore`) |
| `version` | Print version |

Failures exit non-zero; skips are warnings only.

## Agent boundaries

[AGENTS.md](AGENTS.md)

## Release

Push a `v*` tag — see [release.yml](.github/workflows/release.yml).

## Development

```bash
go test ./...
go run ./cmd/kusabi --help
```

## License

[MIT](LICENSE)
