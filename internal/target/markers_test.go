package target

import (
	"errors"
	"strings"
	"testing"
)

const (
	start = "<!-- INDEX:START -->"
	end   = "<!-- INDEX:END -->"
)

func TestLocateFound(t *testing.T) {
	content := "intro\n\n" + start + "\nbody\n" + end + "\n\noutro"
	ss, se, es, ee, err := Locate(content, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if content[ss:se] != start {
		t.Errorf("start span wrong: %q", content[ss:se])
	}
	if content[es:ee] != end {
		t.Errorf("end span wrong: %q", content[es:ee])
	}
}

func TestLocateMissing(t *testing.T) {
	_, _, _, _, err := Locate("no markers here", start, end)
	if !errors.Is(err, ErrMissingMarkers) {
		t.Errorf("want ErrMissingMarkers, got %v", err)
	}
}

func TestLocateUnbalanced(t *testing.T) {
	_, _, _, _, err := Locate("only start "+start, start, end)
	if !errors.Is(err, ErrUnbalancedMarkers) {
		t.Errorf("want ErrUnbalancedMarkers (start only), got %v", err)
	}
	_, _, _, _, err = Locate("only end "+end, start, end)
	if !errors.Is(err, ErrUnbalancedMarkers) {
		t.Errorf("want ErrUnbalancedMarkers (end only), got %v", err)
	}
	_, _, _, _, err = Locate(end+" then "+start, start, end)
	if !errors.Is(err, ErrUnbalancedMarkers) {
		t.Errorf("want ErrUnbalancedMarkers (out of order), got %v", err)
	}
}

func TestReplacePreservesSurroundingBytes(t *testing.T) {
	content := "PRE\n" + start + "\nold\n" + end + "\nPOST"
	block := start + "\nNEW\n" + end
	out, err := Replace(content, start, end, block)
	if err != nil {
		t.Fatal(err)
	}
	want := "PRE\n" + block + "\nPOST"
	if out != want {
		t.Errorf("Replace mismatch\n got: %q\nwant: %q", out, want)
	}
}

func TestReplaceIdempotent(t *testing.T) {
	content := "PRE\n" + start + "\nold\n" + end + "\nPOST"
	block := start + "\nNEW\n" + end
	first, err := Replace(content, start, end, block)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Replace(first, start, end, block)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Errorf("Replace not idempotent:\n first: %q\nsecond: %q", first, second)
	}
}

func TestInsertAppendsWithSpacing(t *testing.T) {
	block := start + "\nbody\n" + end
	got := Insert("# heading\n", block)
	if !strings.HasSuffix(got, block+"\n") {
		t.Errorf("Insert did not append block: %q", got)
	}
	if !strings.Contains(got, "# heading\n\n"+start) {
		t.Errorf("Insert did not add blank line separator: %q", got)
	}
}

func TestInsertHandlesNoTrailingNewline(t *testing.T) {
	block := start + "\nbody\n" + end
	got := Insert("# heading", block)
	if !strings.HasSuffix(got, block+"\n") {
		t.Errorf("Insert did not append: %q", got)
	}
}

func TestValidateBlock(t *testing.T) {
	if err := ValidateBlock(start+"\nbody\n"+end, start, end); err != nil {
		t.Errorf("valid block rejected: %v", err)
	}
	if err := ValidateBlock("oops"+end, start, end); err == nil {
		t.Errorf("missing start should fail")
	}
	if err := ValidateBlock(start+"oops", start, end); err == nil {
		t.Errorf("missing end should fail")
	}
}
