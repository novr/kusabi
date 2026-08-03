---
name: kusabi
description: Drive the kusabi CLI against a meta-repository - check child repository state, sync them, run commands across them, and pull aggregated context. Use when kusabi.yaml exists in the workspace or an ancestor directory, or when the user names kusabi. Do not use for git submodules, monorepo build tools, or multi-package workspaces that have no kusabi.yaml.
---

# Kusabi

`kusabi` binds separate Git repositories under one parent. The parent tracks only `kusabi.yaml`, policy files, and a managed `.gitignore` block. Child contents stay out of the parent history.

## Where you are working

| Task | Directory |
| :--- | :--- |
| Edit `kusabi.yaml`, policy, parent `.gitignore` | Meta-repo root |
| Edit product code, `git commit` / `git push` / branch work | That child's directory (`repositories.<name>.path`) |

Never commit child contents from the parent. Never commit binding changes from inside a child. If unsure which child owns a path, read `kusabi.yaml` or run `kusabi status --json` — do not invent a `packages/` layout.

## Before you touch anything

```bash
kusabi status --json
```

If that fails with `command not found`, install the CLI (`brew tap novr/taps && brew install kusabi`, or `go install github.com/novr/kusabi/cmd/kusabi@latest`) and retry. If it fails with `kusabi.yaml not found (searched from ...)`, this is not a Kusabi meta-repo — stop and use ordinary `git`. Do not run `kusabi init` unless the user asked to create a meta-repo.

Do **not** run `kusabi version` or `kusabi doctor` on every turn. Use `doctor` when status looks wrong, after cloning a fresh machine, or when the user asks to check health. `--fix-remote` and `--migrate-gitignore` mutate the workspace — only with explicit consent.

## Sync is the dangerous command

`kusabi sync` clones missing children, then for each clean clone: checks out the **declared** `branch` (if set), and `git pull --ff-only`. A clean tree on the wrong branch is **not** skipped — sync will switch it. Dirty trees, detached HEAD with no declared branch, and `sync: false` are skipped with a warning and exit 0.

Before any sync the user did not explicitly request:

1. Run `kusabi status --json`.
2. If any child has `modified` / `untracked` > 0, or `branch` ≠ `declared_branch`, **stop and report**. Do not sync past local work or an intentional feature branch.
3. Only then run `kusabi sync` (add `--jobs 1` when SSH/HTTPS will prompt; parallel jobs hide auth failures. Use `--depth 1` only when history is not needed).

After sync, re-read the per-child lines. Exit 0 does not mean every child moved.

| Line | Meaning | What to do |
| :--- | :--- | :--- |
| `cloned` / `updated` / `switched …, updated` | Applied | Continue |
| `updated: no change` | Already current | Continue; do not retry |
| `skipped: dirty working tree` | Local edits | Tell the user which child; never stash or discard without asking |
| `skipped: detached HEAD` | No declared branch | Ask for a branch, then set `branch:` via `add`/`kusabi.yaml` |
| `skipped: sync disabled` | `sync: false` | Leave it |
| clone/pull error (non-zero overall) | Auth, network, or non-ff | Fix that child; do not re-run full parallel sync blindly |

A sync that exits 0 with several `skipped: dirty…` lines left the workspace **partially** updated. Say so. Do not claim "all repos are synced".

## Everyday routes

| User wants | Do this |
| :--- | :--- |
| What is going on? | `kusabi status --json` (default first move) |
| Pull latest | Gate with status, then `kusabi sync` |
| Same shell command in several children | `kusabi exec --repo a --repo b '…'` — one quoted argument |
| Bind a new child | `kusabi add <name> <url> [--path …] [--branch …] [--role …] [--tags …]`, commit the parent, then sync as a **separate** step |
| Drop a binding | `kusabi remove <name>` updates yaml + managed gitignore only; the directory remains — delete it yourself if the user wants the disk gone |
| Health / `.gitignore` / remote drift | `kusabi doctor` (+ repair flags only if asked) |

Prefer `--repo` / `--tag` over unfiltered `exec`. Prefer declaration names over scanning directories. Do not loop `git -C` over `packages/*` when these commands cover the need.

## Context (rarely)

Use `kusabi context` only when the question truly spans several children (who owns X, onboarding summary). For work inside one known child, read that child's files directly.

```bash
kusabi context --repo app-ios --repo api --max-bytes 40000
```

Always pass `--repo` and a `--max-bytes` fit to the remaining window. Check `Omitted` / `meta.omitted` before concluding something is missing. Widen `includes` in `kusabi.yaml` if the wrong files are read; do not hand-build a substitute dump. Defaults when unset: `README.md`, `CLAUDE.md`.

## Hard stops

- Do not build, release, lock versions, or invent a cross-repo change plan. Kusabi will not do that; say so.
- Do not edit children to make `context` look better, and do not invent context for uncloned children.
- Do not hand-edit the managed `.gitignore` block; `add` / `remove` / `sync` own it.
- Parent commits: declaration + policy + `.gitignore` only.
