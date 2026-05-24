---
id: HOOKS
summary: "Hook adapters (lefthook, husky, raw) — detection order, idempotency, and what each writes"
---

# Hooks

`agents-toc install-hook` wires a pre-commit entry into your existing hook
manager. The entry runs:

```sh
agents-toc sync && git add AGENTS.md
```

This guarantees the INDEX block is up to date *in the same commit* that
introduces or renames the underlying skill / knowledge file.

## Detection order

1. **lefthook** — `lefthook.yml` or `lefthook.yaml` at the project root
2. **husky** — `.husky/` directory at the project root
3. **raw** — fallback: `.git/hooks/pre-commit` (when the project is a git repo)

Override with `--manager lefthook | husky | raw`.

## Idempotency

Running `install-hook` twice is a no-op:

- **lefthook**: the entry is identified by job name `agents-toc-sync` inside
  `pre-commit.commands`. If it already exists with the expected `run` line,
  nothing is written.
- **husky**: a fenced block `# >>> agents-toc-sync >>> ... # <<< agents-toc-sync <<<`
  inside `.husky/pre-commit`. Detected by the start fence.
- **raw**: identical fenced block inside `.git/hooks/pre-commit`.

Use `--force` to refresh a stale entry (e.g. after upgrading agents-toc and
changing the canonical command).

## What each adapter writes

### lefthook

```yaml
pre-commit:
  commands:
    agents-toc-sync:
      run: agents-toc sync && git add AGENTS.md
      stage_fixed: true
```

`stage_fixed: true` ensures lefthook re-stages `AGENTS.md` automatically if
the sync changed it.

> Caveat: lefthook config is round-tripped through YAML, which drops
> comments. If your `lefthook.yml` has comments you want to preserve, edit
> the file by hand and use the snippet above instead of `install-hook`.

### husky

`.husky/pre-commit` gets:

```sh
#!/usr/bin/env sh
# >>> agents-toc-sync >>>
agents-toc sync && git add AGENTS.md
# <<< agents-toc-sync <<<
```

The file is marked executable. If the file already existed with other
commands, the managed block is appended.

### raw

`.git/hooks/pre-commit` gets the same fenced block, with a `#!/usr/bin/env sh`
shebang and `set -e` if the file is created fresh.

## Uninstalling

Currently manual:

- **lefthook**: delete the `agents-toc-sync` key from
  `pre-commit.commands` in your YAML.
- **husky / raw**: remove the lines between the `# >>> agents-toc-sync >>>`
  and `# <<< agents-toc-sync <<<` fences (inclusive).

A future `agents-toc uninstall-hook` will mechanize this.

## CI considerations

The hook is only a convenience for local development. In CI run:

```sh
agents-toc validate          # frontmatter sanity
agents-toc sync --check      # fails if AGENTS.md drifts from sources
```

Both commands exit non-zero on problems. `sync --check` does not modify the
workspace, so it's safe to run before any test step.
