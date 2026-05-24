# agents-toc CONTEXT

Architecture context for agents (human and AI) working on agents-toc itself.
For *user* documentation see `README.md` and `docs/*`.

## Load-bearing invariants

These do not change without a major version bump.

1. **Marker ownership**: agents-toc owns ONLY the bytes between the
   configured `marker_start` and `marker_end` inside the target file.
   Everything outside the markers is the user's. Any code path that writes
   outside this block is a bug — and a security-relevant bug. See
   `internal/target/markers.go` (Replace, Insert) and
   `internal/pipeline/pipeline.go` (Sync).
2. **Idempotency**: running `sync` twice in a row must not change the file.
   This is a unit-tested guarantee (`TestSyncIsIdempotent`).
3. **No content shipping**: the binary ships zero skill / knowledge content.
   The tool maintains an index of files the *user* puts in their project.
4. **Single binary**: no plugin loading, no runtime config sources beyond
   `.agentsmdrc.yaml`. Cross-platform via pure-Go stdlib + three deps.
5. **AGENTS.md only**: no mirroring into CLAUDE.md, .cursorrules, GEMINI.md.
   AGENTS.md is the cross-vendor convention; if a user wants the others
   they symlink.

## Module map

```
main.go                    cobra entrypoint
cmd/                       one file per subcommand; thin shells around pipeline
internal/buildinfo/        Version/Commit/Date stamped via -ldflags
internal/config/           .agentsmdrc.yaml parse, defaults, validation
internal/scan/             frontmatter parse + glob walk + Entry construction
internal/render/           templates → block body string
internal/target/           marker location + atomic write
internal/pipeline/         Sync / Validate / List — what cmd/ calls
internal/hook/             lefthook / husky / raw pre-commit adapters
```

The dependency graph is strict: `cmd` → `pipeline` → `{config,scan,render,target}`. The `hook` package is independent of all the others except `config`-shaped values (target file).

## Real seams

- **Hook adapters** (`internal/hook/{lefthook,husky,raw}.go`): three concrete
  adapters writing the same managed-block command into different
  hook-manager surfaces. This is a real seam — multiple adapters justify it.

## Hypothetical (don't introduce yet) seams

- **fs.FS for scanning**: only one adapter (the OS filesystem) exists. Tests
  use `t.TempDir()` with real I/O. Adding an `fs.FS` parameter would be a
  hypothetical seam with one adapter. Per the
  `improve-codebase-architecture` skill: don't.
- **YAML library swap**: yaml.v3 is fine. No second adapter needed.

## Public Go API stability

Everything in `cmd/` and `internal/` is internal to the binary. There is no
exported library API; the binary is the contract.

## When to bump SchemaVersion

The integer in `.agentsmdrc.yaml` `version:`. Bump it when an existing
config that worked yesterday produces different behavior today. Adding new
optional fields with safe defaults does NOT require a bump.

A bump requires:
1. `internal/config/config.go`: increment `SchemaVersion`.
2. Add a new branch in a future `cmd/migrate.go` (deleted in this codebase
   until v2 actually exists — see ADR-002).
3. Update `docs/CONFIG.md` with both schemas documented side-by-side.

## ADRs

ADR-001 — managed block ownership: see "Marker ownership" above. Decision:
ship a tool that never touches bytes outside the markers, even at the cost
of refusing to operate when markers are missing and `create_if_missing` is
false.

ADR-002 — no migrate command yet: the schema is v1 and there is no v2. A
pass-through `migrate` command fails the deletion test (`improve-codebase-architecture`
skill). It will return when there is a real v1→v2 migration to perform.

ADR-003 — no MCP server: agents-toc writes static files. Runtime serving is
`skills-mcp`'s job. Composability beats reimplementation.
