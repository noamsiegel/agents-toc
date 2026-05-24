package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/noamsiegel/agents-toc/internal/config"
)

// scaffoldProject lays out a sample project under root for end-to-end tests.
func scaffoldProject(t *testing.T, root string) *config.Config {
	t.Helper()
	mustWrite(t, filepath.Join(root, "skills/wt/SKILL.md"),
		"---\nname: wt\ndescription: worktree manager\n---\n# wt\n")
	mustWrite(t, filepath.Join(root, "skills/guardrails/SKILL.md"),
		"---\nname: guardrails\ndescription: git hook layer\n---\n# guardrails\n")
	mustWrite(t, filepath.Join(root, "knowledge/git.md"),
		"---\nid: git\nsummary: git patterns\n---\n# git\n")
	mustWrite(t, filepath.Join(root, "knowledge/testing.md"),
		"---\nid: testing\nsummary: vitest patterns\n---\n# testing\n")
	mustWrite(t, filepath.Join(root, "knowledge/disabled.md"),
		"---\nid: disabled\nsummary: not visible\nenabled: false\n---\n")

	cfg := config.Default()
	cfgPath := filepath.Join(root, ".agentsmdrc.yaml")
	data, err := cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, cfgPath, string(data))
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSyncScaffoldsWhenMissing(t *testing.T) {
	root := t.TempDir()
	cfg := scaffoldProject(t, root)
	res, err := Sync(cfg, Options{Version: "v0.0.0-test"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created {
		t.Errorf("expected Created=true, got %+v", res)
	}
	if res.EntryCount != 4 {
		t.Errorf("expected 4 entries, got %d", res.EntryCount)
	}
	if res.SkippedDisabled != 1 {
		t.Errorf("expected 1 skipped, got %d", res.SkippedDisabled)
	}
	body, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	for _, want := range []string{
		"skills/wt/SKILL.md",
		"skills/guardrails/SKILL.md",
		"knowledge/git.md",
		"knowledge/testing.md",
		"<!-- INDEX:START -->",
		"<!-- INDEX:END -->",
		"v0.0.0-test",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("AGENTS.md missing %q\n--- file ---\n%s", want, got)
		}
	}
	if strings.Contains(got, "knowledge/disabled.md") {
		t.Errorf("disabled entry leaked into output")
	}
}

func TestSyncIsIdempotent(t *testing.T) {
	root := t.TempDir()
	cfg := scaffoldProject(t, root)
	if _, err := Sync(cfg, Options{Version: "v0"}); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := Sync(cfg, Options{Version: "v0"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Changed {
		t.Errorf("second sync should not change content")
	}
	second, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Errorf("byte mismatch between two syncs")
	}
}

func TestSyncInsertsMarkersIntoExistingFile(t *testing.T) {
	root := t.TempDir()
	cfg := scaffoldProject(t, root)
	preamble := "# AGENTS.md\n\nManually written content the user owns.\n"
	mustWrite(t, filepath.Join(root, "AGENTS.md"), preamble)
	res, err := Sync(cfg, Options{Version: "v0"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.MarkersInserted {
		t.Errorf("expected MarkersInserted=true")
	}
	body, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !strings.HasPrefix(string(body), preamble) {
		t.Errorf("preamble not preserved\n%s", body)
	}
	if !strings.Contains(string(body), "<!-- INDEX:START -->") {
		t.Errorf("start marker not inserted")
	}
}

func TestSyncIncludeDisabledRendersAll(t *testing.T) {
	root := t.TempDir()
	cfg := scaffoldProject(t, root)
	_, err := Sync(cfg, Options{Version: "v0", IncludeDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !strings.Contains(string(body), "knowledge/disabled.md") {
		t.Errorf("--include-disabled did not surface disabled entry\n%s", body)
	}
}

func TestSyncPreservesUserContent(t *testing.T) {
	root := t.TempDir()
	cfg := scaffoldProject(t, root)
	const userText = "ABOVE the block\nstays untouched."
	const userBelow = "BELOW the block\nalso untouched."
	// Pre-create AGENTS.md with markers and surrounding content.
	body := userText + "\n\n<!-- INDEX:START -->\nOLD CONTENT\n<!-- INDEX:END -->\n\n" + userBelow + "\n"
	mustWrite(t, filepath.Join(root, "AGENTS.md"), body)
	if _, err := Sync(cfg, Options{Version: "v0"}); err != nil {
		t.Fatal(err)
	}
	updated, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	got := string(updated)
	if !strings.HasPrefix(got, userText) {
		t.Errorf("user text above lost\n%s", got)
	}
	if !strings.HasSuffix(got, userBelow+"\n") {
		t.Errorf("user text below lost\n%s", got)
	}
	if strings.Contains(got, "OLD CONTENT") {
		t.Errorf("old block content not replaced")
	}
}

func TestValidateReportsMissingFields(t *testing.T) {
	root := t.TempDir()
	cfg := scaffoldProject(t, root)
	// knowledge/git.md requires id+summary; remove summary.
	mustWrite(t, filepath.Join(root, "knowledge/broken.md"), "---\nid: broken\n---\n")
	issues, err := Validate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, i := range issues {
		if i.Path == "knowledge/broken.md" && i.Field == "summary" {
			found = true
		}
	}
	if !found {
		t.Errorf("validate did not flag missing summary: %+v", issues)
	}
}

func TestList(t *testing.T) {
	root := t.TempDir()
	cfg := scaffoldProject(t, root)
	res, err := List(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) != 4 {
		t.Errorf("expected 4 visible, got %d", len(res.Entries))
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(res.Skipped))
	}
}
