package render

import (
	"strings"
	"testing"

	"github.com/noamsiegel/agents-toc/internal/config"
	"github.com/noamsiegel/agents-toc/internal/scan"
)

func makeEntries() []scan.Entry {
	return []scan.Entry{
		{Path: "skills/wt/SKILL.md", Label: "skill", Name: "wt", Summary: "worktree manager"},
		{Path: "skills/guardrails/SKILL.md", Label: "skill", Name: "guardrails", Summary: "git hook layer"},
		{Path: "knowledge/git.md", Label: "knowledge", Name: "git", Summary: "git patterns"},
		{Path: "knowledge/testing.md", Label: "knowledge", Name: "testing", Summary: "vitest patterns"},
	}
}

func TestRenderGroupedAlphabetical(t *testing.T) {
	cfg := config.Default()
	entries := makeEntries()
	scan.SortEntries(entries, "alphabetical")
	got := Render(entries, &cfg)
	want := `### skill/

- ` + "`skills/guardrails/SKILL.md`" + ` — git hook layer
- ` + "`skills/wt/SKILL.md`" + ` — worktree manager

### knowledge/

- ` + "`knowledge/git.md`" + ` — git patterns
- ` + "`knowledge/testing.md`" + ` — vitest patterns
`
	if got != want {
		t.Errorf("render mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestRenderEmpty(t *testing.T) {
	cfg := config.Default()
	got := Render(nil, &cfg)
	if got != cfg.Render.EmptyMessage {
		t.Errorf("empty render: got %q want %q", got, cfg.Render.EmptyMessage)
	}
}

func TestRenderFlat(t *testing.T) {
	cfg := config.Default()
	cfg.Render.GroupBy = "none"
	entries := makeEntries()
	scan.SortEntries(entries, "alphabetical")
	got := Render(entries, &cfg)
	if strings.Contains(got, "### skill") {
		t.Errorf("flat render should not have group headers: %s", got)
	}
	if !strings.Contains(got, "knowledge/git.md") || !strings.Contains(got, "skills/wt/SKILL.md") {
		t.Errorf("flat render missing entries: %s", got)
	}
}

func TestBlockWrapsMarkers(t *testing.T) {
	cfg := config.Default()
	block := Block(makeEntries(), &cfg, "v0.1.0")
	if !strings.HasPrefix(block, cfg.Target.MarkerStart) {
		t.Errorf("block missing start marker")
	}
	if !strings.HasSuffix(block, cfg.Target.MarkerEnd) {
		t.Errorf("block missing end marker")
	}
	if !strings.Contains(block, "v0.1.0") {
		t.Errorf("block missing version stamp")
	}
}

func TestRenderRespectsCustomTemplate(t *testing.T) {
	cfg := config.Default()
	cfg.Render.LineTemplate = "* {name} ({label}): {summary}"
	cfg.Render.GroupBy = "none"
	entries := []scan.Entry{
		{Path: "x/a.md", Label: "k", Name: "alpha", Summary: "first"},
	}
	got := Render(entries, &cfg)
	want := "* alpha (k): first"
	if got != want {
		t.Errorf("custom template: got %q want %q", got, want)
	}
}

func TestRenderGroupCount(t *testing.T) {
	cfg := config.Default()
	cfg.Render.GroupTemplate = "## {label} ({count})\n{lines}"
	entries := makeEntries()
	scan.SortEntries(entries, "alphabetical")
	got := Render(entries, &cfg)
	if !strings.Contains(got, "## skill (2)") || !strings.Contains(got, "## knowledge (2)") {
		t.Errorf("group count missing: %s", got)
	}
}
