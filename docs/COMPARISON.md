---
id: COMPARISON
summary: "How agents-toc differs from agents-md, skills-mcp, anthropics/skills, agent-rules-skill"
---

# How agents-toc compares

Short version: every adjacent tool solves a *different* problem. agents-toc
sits in a gap none of them fill.

## At a glance

| Project | Core verb | Output | Loaded at | Composable with agents-toc? |
|---|---|---|---|---|
| **agents-toc** | index | managed INDEX block inside your AGENTS.md | commit-time | — |
| [`ivawzh/agents-md`](https://github.com/ivawzh/agents-md) | compose | the whole AGENTS.md from fragments | commit-time | yes — agents-md owns the prose, agents-toc owns the INDEX block |
| [`skills-mcp/skills-mcp`](https://github.com/skills-mcp/skills-mcp) | serve | nothing on disk; serves skills via MCP | runtime | yes — same `SKILL.md` files feed both |
| [`netresearch/agent-rules-skill`](https://github.com/netresearch/agent-rules-skill) | onboard | a fresh AGENTS.md from repo metadata | one-shot setup | partial — useful once, not a steady-state tool |
| [`anthropics/skills`](https://github.com/anthropics/skills) | spec | nothing — defines the SKILL.md format | n/a | yes — agents-toc *consumes* this spec by default |
| [`wshobson/agents`](https://github.com/wshobson/agents) | marketplace | plugin artifacts for many harnesses | install | orthogonal |
| [`joventuraz/skillpack`](https://github.com/joventuraz/skillpack) | package | skill bundles via a lockfile | install | orthogonal |
| [`opencode-rules`](https://github.com/frap129/opencode-rules) | activate | dynamic rule injection by trigger | runtime | orthogonal |

## In words

### agents-md

`agents-md` composes a whole `AGENTS.md` by concatenating markdown fragments
you stage in a directory. It owns the file end-to-end.

**Difference**: agents-toc does **not** own the file. It owns one fenced
block inside it. Everything else — prose, rules, examples, headings — is
yours.

**Compose**: use `agents-md` for the prose composition, then put the
agents-toc fence into one of the composed fragments. agents-toc rewrites
that block on every commit; agents-md happily includes it as opaque text.

### skills-mcp

`skills-mcp` is an MCP server. At runtime, an MCP-compatible agent calls
`list_skills` / `get_skill` and the server reads SKILL.md files from disk.

**Difference**: agents-toc writes a static block to disk. The agent never
calls a server; it just reads `AGENTS.md` like any other file.

**Compose**: point both tools at the same `skills/` directory. agents-toc
indexes them at commit time; skills-mcp serves them at runtime. The same
`SKILL.md` is the source of truth for both.

### netresearch/agent-rules-skill

Generates a project's first AGENTS.md by scraping repo metadata (package.json,
README, etc.). One-shot bootstrap.

**Difference**: agents-toc is a steady-state tool. It keeps a section of an
*existing* AGENTS.md current with the skills/knowledge directory.

**Compose**: run `agent-rules-skill` once to scaffold AGENTS.md. Drop the
agents-toc fence into it. Use agents-toc from then on.

### anthropics/skills

Anthropic's official spec + example bundle for SKILL.md. Defines the
`name` / `description` frontmatter shape.

**Difference**: it's a spec, not a tool. agents-toc *implements* this spec
as its default source pattern.

**Compose**: write your skills following the spec. agents-toc indexes them
out of the box.

## Why these gaps add up to a tool

Every tool above either:

1. owns the whole AGENTS.md file (agents-md, agent-rules-skill),
2. doesn't touch AGENTS.md at all (skills-mcp, anthropics/skills, marketplaces),
3. or generates non-AGENTS.md artifacts (clinerules, opencode-rules,
   cursorrules collections).

Nothing in the ecosystem maintains a *managed sub-block* inside an existing
AGENTS.md from local source files. That sub-block is the cheapest possible
implementation of progressive disclosure — agent sees the index always,
loads bodies only when needed — and it's exactly what agents-toc does.
