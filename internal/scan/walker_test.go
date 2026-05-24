package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/noamsiegel/agents-toc/internal/config"
)

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanFindsSkillsAndKnowledge(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "skills/wt/SKILL.md"),
		"---\nname: wt\ndescription: worktree manager\n---\n# wt\n")
	mustWrite(t, filepath.Join(root, "skills/guardrails/SKILL.md"),
		"---\nname: guardrails\ndescription: git hook layer\n---\n# guardrails\n")
	mustWrite(t, filepath.Join(root, "knowledge/git.md"),
		"---\nid: git\nsummary: git patterns\n---\n# git\n")
	mustWrite(t, filepath.Join(root, "knowledge/testing.md"),
		"---\nid: testing\nsummary: vitest patterns\n---\n# testing\n")
	mustWrite(t, filepath.Join(root, "node_modules/pkg/SKILL.md"),
		"---\nname: ignored\ndescription: should be excluded\n---\n")

	cfg := config.Default()
	entries, err := Scan(root, &cfg)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d: %+v", len(entries), entries)
	}
	want := map[string]string{
		"skills/wt/SKILL.md":           "skill",
		"skills/guardrails/SKILL.md":   "skill",
		"knowledge/git.md":             "knowledge",
		"knowledge/testing.md":         "knowledge",
	}
	for _, e := range entries {
		label, ok := want[e.Path]
		if !ok {
			t.Errorf("unexpected entry path %q", e.Path)
			continue
		}
		if e.Label != label {
			t.Errorf("%s: label got %q want %q", e.Path, e.Label, label)
		}
	}
}

func TestScanRespectsDisabled(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "knowledge/active.md"),
		"---\nid: active\nsummary: live\n---\n")
	mustWrite(t, filepath.Join(root, "knowledge/dormant.md"),
		"---\nid: dormant\nsummary: not yet\nenabled: false\n---\n")

	cfg := config.Default()
	entries, err := Scan(root, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	var dormant *Entry
	for i := range entries {
		if entries[i].Path == "knowledge/dormant.md" {
			dormant = &entries[i]
		}
	}
	if dormant == nil {
		t.Fatal("dormant entry not found")
	}
	if !dormant.Disabled {
		t.Errorf("expected dormant to be Disabled=true")
	}
	filtered := EntryFilter(entries, false)
	for _, e := range filtered {
		if e.Path == "knowledge/dormant.md" {
			t.Errorf("filter should drop disabled entry")
		}
	}
}

func TestScanFallsBackToBasename(t *testing.T) {
	root := t.TempDir()
	// SKILL.md without name field.
	mustWrite(t, filepath.Join(root, "skills/nameless/SKILL.md"),
		"---\ndescription: anon\n---\n")
	cfg := config.Default()
	entries, err := Scan(root, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "SKILL" {
		t.Errorf("name fallback wrong: %+v", entries)
	}
}

func TestScanDedupesAcrossSources(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "shared/SKILL.md"),
		"---\nname: dup\ndescription: appears in two source globs\n---\n")
	cfg := config.Default()
	// Overlapping globs that both match shared/SKILL.md.
	cfg.Sources = []config.SourceConfig{
		{Glob: "**/SKILL.md", Label: "first", Frontmatter: cfg.Sources[0].Frontmatter},
		{Glob: "shared/SKILL.md", Label: "second", Frontmatter: cfg.Sources[0].Frontmatter},
	}
	entries, err := Scan(root, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Label != "first" {
		t.Errorf("dedup wrong: %+v", entries)
	}
}

func TestSortEntriesAlphabetical(t *testing.T) {
	in := []Entry{{Path: "z"}, {Path: "a"}, {Path: "m"}}
	SortEntries(in, "alphabetical")
	if in[0].Path != "a" || in[2].Path != "z" {
		t.Errorf("alphabetical sort wrong: %+v", in)
	}
}
