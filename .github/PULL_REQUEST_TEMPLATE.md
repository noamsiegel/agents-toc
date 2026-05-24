<!--
Thanks for the PR.

Before submitting, please check:

- [ ] Tests added or updated under `*_test.go` next to the code you changed.
- [ ] `go test -race ./...` passes locally.
- [ ] No file is touched outside its source-of-truth (e.g. you didn't paper over
      a config-default change in two places).
- [ ] Any user-facing surface change is reflected in `README.md` and the
      relevant `docs/*.md`.
- [ ] If you touched the `.agentsmdrc.yaml` schema, you bumped
      `config.SchemaVersion` and updated `docs/CONFIG.md`.

## Summary

<!-- One sentence: what changed and why. -->

## Marker-ownership invariant

agents-toc's hard contract is: the tool owns ONLY the bytes between the
configured markers in `AGENTS.md`. Confirm one:

- [ ] This PR does not change bytes outside the markers.
- [ ] This PR intentionally changes that contract; here's the rationale: ...
