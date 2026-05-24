---
id: CONFIG
summary: ".agentsmdrc.yaml schema reference — every field, default, and validation rule"
---

# `.agentsmdrc.yaml` reference

This is the full schema for `version: 1`. Every field has a sensible default;
the minimal valid config is two lines:

```yaml
version: 1
sources:
  - glob: "**/SKILL.md"
    label: "skill"
```

## Top-level

| Field | Type | Required | Description |
|---|---|---|---|
| `version` | int | yes | Schema version. Must be `1`. agents-toc refuses to run against newer schema versions and prints a migrate hint. |
| `sources` | list | yes | At least one source. Evaluated in order; deduplication keeps the first source's interpretation when globs overlap. |
| `target` | object | no | AGENTS.md location and markers. Defaults shown below. |
| `render` | object | no | Block appearance. Defaults shown below. |
| `ignore` | list of globs | no | Project-relative doublestar globs. A match excludes the file from every source. |

## `sources[]`

| Field | Type | Required | Description |
|---|---|---|---|
| `glob` | string | yes | Doublestar pattern relative to the project root. `**/` matches any number of path segments. |
| `label` | string | yes | Short tag used by the renderer to group entries. |
| `frontmatter` | object | no | How to extract name/summary from this source's frontmatter. |

### `sources[].frontmatter`

| Field | Type | Default | Description |
|---|---|---|---|
| `name_field` | string | "name" | YAML key for the entry name. Falls back to the file's basename when the field is missing. |
| `summary_field` | string | "description" | YAML key for the one-line summary. |
| `require` | list of strings | `[]` | Required fields. `agents-toc validate` exits non-zero when any required field is missing. |
| `respect_disabled` | bool | true | Skip files whose frontmatter contains `enabled: false`. |

## `target`

| Field | Type | Default | Description |
|---|---|---|---|
| `file` | string | "AGENTS.md" | Path relative to the project root. |
| `marker_start` | string | `<!-- INDEX:START -->` | Fence marker. Must differ from `marker_end`. |
| `marker_end` | string | `<!-- INDEX:END -->` | Fence marker. Tool owns everything between the two; outside bytes are untouched. |
| `create_if_missing` | bool | true | When `AGENTS.md` doesn't exist, `sync` scaffolds a minimal one with markers. When markers are missing, they're appended. |

## `render`

| Field | Type | Default | Description |
|---|---|---|---|
| `group_by` | `"label"` \| `"none"` | "label" | Group entries by source label, or emit a flat list. |
| `sort` | `"alphabetical"` \| `"mtime"` | "alphabetical" | Sort within each group. |
| `line_template` | string | `` - `{path}` — {summary} `` | Per-entry template. Placeholders: `{path}`, `{name}`, `{summary}`, `{label}`. |
| `group_template` | string | `### {label}/\n\n{lines}\n` | Group wrapper. Placeholders: `{label}`, `{lines}`, `{count}`. |
| `empty_message` | string | `_no entries_` | Body when nothing matches. |
| `include_disabled` | bool | false | Render disabled entries by default. Equivalent to `agents-toc sync --include-disabled`. |

## Frontmatter conventions

agents-toc consumes whatever YAML you put in your files. Two shapes are
mainstream and supported out of the box:

### Anthropic SKILL.md

```yaml
---
name: wt
description: Git worktree manager with herdr tabs and branch-shape enforcement
---
```

### Knowledge markdown

```yaml
---
id: git
summary: Git workflow patterns — commits, branches, rebasing, recovery
tags: [git, workflow]
enabled: true
---
```

The `enabled: false` flag is honored project-wide. Use it for sensitive
files, drafts, or local-only content you don't want in the public index.

## Glob examples

| Pattern | Matches |
|---|---|
| `skills/**/SKILL.md` | every SKILL.md anywhere under skills/ |
| `docs/*.md` | only the docs/ root, not subdirectories |
| `**/*.md` | every markdown file in the project |
| `apps/*/SKILL.md` | one level deep only |

agents-toc uses [doublestar](https://github.com/bmatcuk/doublestar) — full
syntax reference there.

## Validation rules

- `sources` cannot be empty.
- Every source must have non-empty `glob` and `label`.
- `marker_start` and `marker_end` must both be non-empty and distinct.
- `render.group_by` must be `"label"` or `"none"`.
- `render.sort` must be `"alphabetical"` or `"mtime"`.

Run `agents-toc validate` to also check that every file's frontmatter
satisfies its source's `require` list.
