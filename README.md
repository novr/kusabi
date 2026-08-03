# Kusabi (楔)

*Connect Repositories. Unify Contexts.*

Kusabi is a **meta-repo CLI**: declare child Git repositories, keep their files out of the parent history, sync and inspect them together, and emit **`kusabi context`** for agents. It is not a submodule wrapper or a monorepo build orchestrator.

Commit in the meta repo: `kusabi.yaml`, policy files (`AGENTS.md`, `context.paths`), and the managed `.gitignore` block — not child repository contents.

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
kusabi add app-ios git@github.com:org/app-ios.git --role "iOS App" --branch develop
kusabi add app-backend git@github.com:org/app-backend.git --role "API Server"
kusabi sync
kusabi status
kusabi doctor
```

`sync` checks out each child's declared `branch` when the working tree is clean, then pulls `--ff-only`. Dirty children are skipped (warning, exit 0) — read the per-child lines; exit 0 is not "everything updated". Product commits happen inside each child directory, not at the meta-repo root.

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
| `context.paths` | Parent files included in `context` (missing paths are reported) |
| `includes` | Per-child context paths; inherits parent `context.includes` when omitted |

### Context includes

Per child, files are read in this order:

1. `repositories.<name>.includes`
2. `context.includes` in the manifest
3. Default: `README.md`, `CLAUDE.md` (only when both above are unset)

Missing child include files are skipped silently. `context --max-bytes` caps observed file/tree content (headings and overview stay); omitted paths are listed at the end (or in JSON `meta.omitted`).

## Commands

| Command | |
| :--- | :--- |
| `init [--force]` | Scaffold `kusabi.yaml`, `AGENTS.md`, `.gitignore` |
| `add` / `remove` | Edit declaration (`--path`, `--role`, `--tags`, `--branch`) |
| `sync [--depth] [--jobs]` | Clone or pull; align to declared branch |
| `status [--json]` | Branch and working tree per child |
| `exec` | Shell command across children (`--repo`, `--tag`, `--skip-uncloned`) |
| `context [--tree] [--json] [--repo] [--max-bytes]` | Observed context to stdout |
| `doctor` | Health checks (`--fix-remote`, `--migrate-gitignore`) |
| `version` | Print version |

Failures exit non-zero; skips are warnings only.

## Agent skill

Install into each **meta repo** (so the team shares the same agent behavior):

```bash
cd /path/to/your-meta-repo
npx skills add novr/kusabi
```

Use `-g` only for personal use across several meta repos. The skill assumes `kusabi` is already on `PATH` (see Install). It teaches agents to run `status` before `sync`, treat skipped dirty children as a partial update, and keep product commits inside children — see [.agents/skills/kusabi](.agents/skills/kusabi/SKILL.md).

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
