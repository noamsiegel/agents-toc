---
id: CONTRIBUTING
summary: "Dev setup, repo layout, testing rules, schema bump procedure, release flow"
---

# Contributing

agents-toc is small and stays small. Contributions that keep it that way are
welcome.

## Dev setup

```sh
git clone https://github.com/noamsiegel/agents-toc
cd agents-toc
make build           # build local binary into ./bin/agents-toc
make test            # full test suite (race detector enabled)
make snapshot        # goreleaser snapshot build (requires goreleaser installed)
```

Requirements:

- Go 1.22+
- (optional) goreleaser for `make snapshot` / `make release-dry-run`
- (optional) golangci-lint for `make lint`

## Layout

```
.
├── main.go                       # cobra entrypoint
├── cmd/                          # one file per subcommand
├── internal/
│   ├── buildinfo/                # version, commit, date (ldflags-stamped)
│   ├── config/                   # .agentsmdrc.yaml schema + Load/Default
│   ├── scan/                     # frontmatter parsing + glob walking
│   ├── render/                   # template-based block rendering
│   ├── target/                   # marker location + atomic write
│   ├── pipeline/                 # Sync / Validate / List orchestration
│   └── hook/                     # lefthook / husky / raw adapters
└── docs/                         # user-facing documentation
```

The CLI commands in `cmd/` are *thin*. Real logic lives in `internal/`;
commands just parse flags and call into `internal/pipeline`.

## Testing rules

- Real filesystem in `t.TempDir()`; no mocks of `os` or `io/fs`.
- Golden tests live alongside their package as table-driven cases.
- A test that needs git can write a stub `.git/HEAD` — most adapter tests
  do this rather than spawning real git.
- Race detector is required: `go test -race ./...` must stay green.

## Adding a new feature

1. Open an issue first if the change is non-trivial. The boring-by-design
   policy means "no" is the default answer to scope-expanding proposals.
2. Add a focused unit test before the implementation.
3. Update `docs/CONFIG.md` if you touched the schema, `docs/HOOKS.md` if you
   touched an adapter, and `README.md` if you changed the user surface.

## Schema bumps

Any breaking change to `.agentsmdrc.yaml` must:

1. Bump `config.SchemaVersion`.
2. Add a `cmd/migrate.go` branch that mechanically translates the old schema
   to the new one.
3. Update `docs/CONFIG.md` with both schema versions documented.

## Release flow

Tag a commit on `main` matching `v*.*.*`. GitHub Actions runs goreleaser,
which:

- Builds darwin/linux/windows × arm64/amd64 archives.
- Creates a GitHub Release with the archives and checksums.
- Updates the formula in `noamsiegel/homebrew-tap`.

No manual steps.
