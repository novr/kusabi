# AGENTS.md — Boundaries

Not a submodule, monorepo orchestrator, or cross-repo change coordinator.

## Concerns

### Binding declaration

- Hold, change, and validate the binding definition.
- Do not reach into children or assemble context documents.
- Keep parent `.gitignore` exclusions in sync with the declaration.

### Actions on children

- Fetch, update, inspect, and run commands per child from the declaration.
- Read the declaration only; do not rewrite it or emit context documents.

### Context aggregation

- Combine policy, roles, and child documents into one observation.
- Do not modify repositories or fabricate missing context.

## Out of scope

- Child repo build, design, or release semantics
- Cross-repo dependency or version-lock governance
- Drafting agent policy (observe declared documents only)
- Merging parent and children into one Git history

## Collision rules

1. **One side-effect owner** — Declaration updates, parent exclusions, child actions, and context output are separate outcomes.
2. **Actions are subordinate to the declaration** — Reference-only dependency; do not silently change bindings.
3. **Aggregation is always observation** — Never rewrite children or the declaration for aggregation I/O.
4. **Delivery is not a concern** — CLI is wiring only; judgment and action logic live outside the entry point.
