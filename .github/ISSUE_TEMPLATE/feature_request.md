---
name: Feature request
about: Propose a new feature.
labels: enhancement
---

## What problem does this solve?

<!-- Concrete user story. "I want X so I can Y." -->

## Proposed surface

<!-- Flag, command, config field, or behavior change. -->

## Scope check

agents-toc is intentionally narrow. Before opening this, please check the
non-goals in the README:

- It is **not** a skill marketplace, registry, or package manager.
- It does **not** ship skill or knowledge content.
- It does **not** serve skills at runtime (that's `skills-mcp`).
- It does **not** mirror to CLAUDE.md / .cursorrules / GEMINI.md.

If your proposal lands in a non-goal, it likely belongs in `agents-md`,
`skills-mcp`, or as a separate tool that composes with agents-toc.
