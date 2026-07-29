# AGENTS.md — Concerns, Responsibilities, Boundaries

Kusabi's job is limited to **binding** multiple repositories and **surfacing** them as one context.
The moment it tries to be smarter, it slides into Submodule territory or monorepo orchestration.

## In scope / out of scope

In scope:

- Declaring where children live, their roles, and how they are bound
- Defending the parent history from child content
- Per-child actions on the declared set (sync, status, bulk exec)
- Read-only context aggregation from the declaration and each child's local documents

Out of scope:

- Understanding design, build, or release semantics inside child repositories
- Coordinating cross-repo changes, dependency resolution, or version-lock governance
- Drafting agent behavior policy (aggregation only; it does not own correctness)
- Merging parent and children into one Git history

## Concerns

### Binding declaration

What to bind, where, and with what role. This is the single source of truth for composition.

- **Responsibility**: Hold, change, and validate the binding definition.
- **Boundary**: Do not reach into children. Do not assemble context documents.
- **Collateral**: Keep parent working-tree exclusions in sync with the declaration so the parent history stays clean. Exclusion alone is not a standalone policy.

### Actions on children

Deliver effects to each child repository using the declaration as input.

- **Responsibility**: Fetch, update, inspect state, and apply commands per child. Target set comes from the declaration.
- **Boundary**: Do not rewrite the declaration (read only). Do not produce aggregated documents. Do not interpret child internals.

### Context aggregation

Fold the meta workspace into a form agents and developers can read at once.

- **Responsibility**: Combine global policy, roles, and each child's surfaced documents into one observation artifact.
- **Boundary**: Do not modify repositories. Do not judge policy correctness. Do not fill gaps to fabricate a plausible-looking context.

## Collision rules

Concerns run together in real use cases. These lines must not break:

1. **One side-effect owner** — Declaration updates, parent exclusions, child actions, and context output are separate outcomes. If one operation has multiple side effects, name the owner for each.
2. **Actions are subordinate to the declaration** — Action target selection may depend on the declaration. Dependency is reference-only; actions must not silently change how things are bound.
3. **Aggregation is always observation** — Aggregation I/O must never be a reason to rewrite children or the declaration.
4. **Delivery is not a concern** — CLI and other entry points are wiring that exposes the three concerns above. Judgment and action logic do not live at the entry point.
