package scan

import (
	"strings"
	"testing"
)

func TestParseFrontmatterValid(t *testing.T) {
	src := "---\nname: greet\ndescription: say hi\nenabled: true\n---\n\n# Body\n"
	fm, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fm.Present {
		t.Fatal("expected Present=true")
	}
	if got := fm.String("name"); got != "greet" {
		t.Errorf("name: got %q", got)
	}
	if got := fm.String("description"); got != "say hi" {
		t.Errorf("description: got %q", got)
	}
	if b, ok := fm.Bool("enabled"); !ok || !b {
		t.Errorf("enabled: got %v ok=%v", b, ok)
	}
	if string(fm.Body) != "\n# Body\n" {
		t.Errorf("body: got %q", string(fm.Body))
	}
}

func TestParseFrontmatterAbsent(t *testing.T) {
	src := "# Just a heading\n"
	fm, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if fm.Present {
		t.Fatal("expected Present=false")
	}
	if fm.String("name") != "" {
		t.Error("String() should return empty when not present")
	}
}

func TestParseFrontmatterStripsBOM(t *testing.T) {
	src := "\xEF\xBB\xBF---\nname: a\n---\nbody"
	fm, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !fm.Present || fm.String("name") != "a" {
		t.Errorf("BOM not handled: %#v", fm)
	}
}

func TestParseFrontmatterEmptyBlock(t *testing.T) {
	src := "---\n---\nbody"
	fm, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !fm.Present {
		t.Fatal("empty block should still be Present=true")
	}
	if len(fm.Fields) != 0 {
		t.Errorf("expected empty fields, got %v", fm.Fields)
	}
}

func TestParseFrontmatterUnterminated(t *testing.T) {
	src := "---\nname: oops\nbody continues forever\n"
	_, err := ParseFrontmatter([]byte(src))
	if err == nil || !strings.Contains(err.Error(), "closing delimiter") {
		t.Fatalf("expected unterminated error, got %v", err)
	}
}

func TestParseFrontmatterBadYAML(t *testing.T) {
	src := "---\n: this is not a mapping\n---\nbody"
	_, err := ParseFrontmatter([]byte(src))
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseFrontmatterEOFClose(t *testing.T) {
	// A frontmatter terminated by `\n---` at EOF (no trailing newline).
	src := "---\nname: tail\n---"
	fm, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !fm.Present || fm.String("name") != "tail" {
		t.Errorf("EOF close not handled: %#v", fm)
	}
	if len(fm.Body) != 0 {
		t.Errorf("expected empty body, got %q", fm.Body)
	}
}
