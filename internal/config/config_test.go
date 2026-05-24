package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultIsValid(t *testing.T) {
	d := Default()
	d.path = "/x/.agentsmdrc.yaml"
	if err := d.Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agentsmdrc.yaml")
	body := `version: 1
sources:
  - glob: "things/**/*.md"
    label: thing
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Target.File != "AGENTS.md" {
		t.Errorf("target.file default missing, got %q", cfg.Target.File)
	}
	if cfg.Target.MarkerStart != "<!-- INDEX:START -->" {
		t.Errorf("marker_start default missing, got %q", cfg.Target.MarkerStart)
	}
	if cfg.Render.LineTemplate == "" {
		t.Errorf("line_template default missing")
	}
	if len(cfg.Ignore) == 0 {
		t.Errorf("ignore defaults missing")
	}
	if cfg.Sources[0].Frontmatter.RespectDisabled == nil || !*cfg.Sources[0].Frontmatter.RespectDisabled {
		t.Errorf("respect_disabled default missing or false")
	}
}

func TestLoadRejectsMissingVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agentsmdrc.yaml")
	if err := os.WriteFile(path, []byte("sources: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version error, got %v", err)
	}
}

func TestLoadRejectsFutureVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agentsmdrc.yaml")
	body := `version: 99
sources:
  - glob: "**/*.md"
    label: x
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "newer") {
		t.Fatalf("expected newer-schema error, got %v", err)
	}
}

func TestValidateRejectsEmptySources(t *testing.T) {
	c := Default()
	c.Sources = nil
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty sources")
	}
}

func TestValidateRejectsIdenticalMarkers(t *testing.T) {
	c := Default()
	c.Target.MarkerEnd = c.Target.MarkerStart
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for identical markers")
	}
}

func TestValidateRejectsBadGroupBy(t *testing.T) {
	c := Default()
	c.Render.GroupBy = "weird"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for bad group_by")
	}
}

func TestLocateFindsUpward(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".agentsmdrc.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nsources:\n  - {glob: 'x', label: y}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Locate(deep)
	if got != path {
		t.Errorf("Locate: got %q, want %q", got, path)
	}
}

func TestLocateReturnsEmptyWhenAbsent(t *testing.T) {
	if got := Locate(t.TempDir()); got != "" {
		t.Errorf("expected empty path, got %q", got)
	}
}
