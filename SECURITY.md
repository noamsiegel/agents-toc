# Security Policy

## Supported versions

agents-toc is pre-1.0. Only the latest tagged release receives security fixes.

| Version | Supported |
|---|---|
| Latest `v0.x` | yes |
| Earlier `v0.x` | no |

## Reporting a vulnerability

Email **noam.r.siegel@gmail.com** with subject `agents-toc security: <short description>`. Do **not** open a public issue.

Include if you have it:

- A minimal reproducer.
- The affected version (`agents-toc --version`).
- Your assessment of impact.

I'll acknowledge within 7 days and aim to ship a fix or mitigation within 30 days, sooner if exploitation is straightforward.

## What counts

This binary writes files in the project it runs in and installs git hooks. In scope:

- Supply-chain compromise of the binary distribution channels (Homebrew formula, GitHub Releases, `install.sh`).
- Arbitrary file writes outside the configured `target.file` or hook paths.
- Hook installation that compromises the host (e.g. shell injection through a malicious `target.file` value).
- Frontmatter or YAML parsing flaws that lead to RCE.

Out of scope:

- A maliciously crafted `AGENTS.md` already inside a repo you've chosen to run agents-toc against — the tool's contract is that you own the bytes outside the markers.
- Performance issues on pathological input.

## Verifying releases

Every GitHub Release includes `checksums.txt`. The `install.sh` script downloads and verifies that file before extracting any binary.
