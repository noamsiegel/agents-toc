package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestDetectLefthook(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "lefthook.yml"), "pre-commit:\n  commands: {}\n")
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref: refs/heads/main\n")
	m, err := Detect(root)
	if err != nil {
		t.Fatal(err)
	}
	if m != ManagerLefthook {
		t.Errorf("got %q, want lefthook", m)
	}
}

func TestDetectHusky(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".husky"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")
	m, err := Detect(root)
	if err != nil {
		t.Fatal(err)
	}
	if m != ManagerHusky {
		t.Errorf("got %q, want husky", m)
	}
}

func TestDetectRawFallback(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")
	m, err := Detect(root)
	if err != nil {
		t.Fatal(err)
	}
	if m != ManagerRaw {
		t.Errorf("got %q, want raw", m)
	}
}

func TestDetectErrorsWithoutGit(t *testing.T) {
	root := t.TempDir()
	if _, err := Detect(root); err == nil {
		t.Errorf("expected error for non-git dir")
	}
}

func TestInstallLefthookIdempotent(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "lefthook.yml"),
		"pre-commit:\n  commands:\n    existing-job:\n      run: echo hi\n")
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")

	r1, err := Install(root, ManagerLefthook, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !r1.Modified {
		t.Errorf("first install should modify")
	}
	r2, err := Install(root, ManagerLefthook, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !r2.AlreadyPresent {
		t.Errorf("second install should be no-op, got %+v", r2)
	}
	body, _ := os.ReadFile(filepath.Join(root, "lefthook.yml"))
	got := string(body)
	if !strings.Contains(got, JobName) {
		t.Errorf("job name missing\n%s", got)
	}
	if !strings.Contains(got, "existing-job") {
		t.Errorf("existing job clobbered\n%s", got)
	}
	if !strings.Contains(got, BuildCommand("")) {
		t.Errorf("command missing\n%s", got)
	}
}

func TestInstallHuskyCreatesFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".husky"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")

	r, err := Install(root, ManagerHusky, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Created {
		t.Errorf("expected Created=true, got %+v", r)
	}
	body, err := os.ReadFile(filepath.Join(root, ".husky", "pre-commit"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), BuildCommand("")) {
		t.Errorf("husky hook missing command\n%s", body)
	}
	st, _ := os.Stat(filepath.Join(root, ".husky", "pre-commit"))
	if st.Mode().Perm()&0o111 == 0 {
		t.Errorf("husky hook not executable: %v", st.Mode())
	}
}

func TestInstallHuskyAppendsToExisting(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".husky/pre-commit"),
		"#!/usr/bin/env sh\nset -e\nnpm test\n")
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")

	r, err := Install(root, ManagerHusky, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Modified {
		t.Errorf("expected Modified=true, got %+v", r)
	}
	body, _ := os.ReadFile(filepath.Join(root, ".husky", "pre-commit"))
	got := string(body)
	if !strings.Contains(got, "npm test") {
		t.Errorf("existing line clobbered\n%s", got)
	}
	if !strings.Contains(got, BuildCommand("")) {
		t.Errorf("command missing\n%s", got)
	}
}

func TestInstallRawCreatesFile(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")

	r, err := Install(root, ManagerRaw, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Created {
		t.Errorf("expected Created=true, got %+v", r)
	}
	body, _ := os.ReadFile(filepath.Join(root, ".git/hooks/pre-commit"))
	got := string(body)
	if !strings.HasPrefix(got, "#!/usr/bin/env sh") {
		t.Errorf("shebang missing\n%s", got)
	}
	if !strings.Contains(got, BuildCommand("")) {
		t.Errorf("command missing\n%s", got)
	}
}

func TestInstallRawIdempotent(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")
	if _, err := Install(root, ManagerRaw, "", false); err != nil {
		t.Fatal(err)
	}
	r, err := Install(root, ManagerRaw, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !r.AlreadyPresent {
		t.Errorf("expected AlreadyPresent on second run, got %+v", r)
	}
}

func TestStripManagedNoOp(t *testing.T) {
	body := "untouched"
	got := stripManaged(body, "missing", "marker")
	if got != body {
		t.Errorf("strip should be no-op when fences missing: %q", got)
	}
}

func TestBuildCommandHonorsTargetFile(t *testing.T) {
	default_ := BuildCommand("")
	if !strings.Contains(default_, "AGENTS.md") {
		t.Errorf("empty targetFile should default to AGENTS.md, got %q", default_)
	}
	custom := BuildCommand("docs/AGENTS.md")
	if !strings.Contains(custom, "docs/AGENTS.md") || strings.Contains(custom, " AGENTS.md") {
		t.Errorf("custom target not threaded into command: %q", custom)
	}
}

func TestInstallRawWithCustomTarget(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")
	const custom = "docs/AGENTS.md"
	if _, err := Install(root, ManagerRaw, custom, false); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(root, ".git/hooks/pre-commit"))
	if !strings.Contains(string(body), "git add docs/AGENTS.md") {
		t.Errorf("custom target not written into raw hook: %s", body)
	}
}

func TestInstallLefthookWithCustomTarget(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "lefthook.yml"), "pre-commit:\n  commands: {}\n")
	mustWrite(t, filepath.Join(root, ".git/HEAD"), "ref\n")
	const custom = "docs/AGENTS.md"
	if _, err := Install(root, ManagerLefthook, custom, false); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(root, "lefthook.yml"))
	if !strings.Contains(string(body), "git add docs/AGENTS.md") {
		t.Errorf("custom target not written into lefthook yml: %s", body)
	}
}
