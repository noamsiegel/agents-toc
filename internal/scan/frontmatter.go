// Package scan walks the configured source globs, extracts YAML frontmatter,
// and produces ordered entries the renderer and validator consume.
package scan

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Frontmatter is the parsed top-of-file YAML block, plus whether the file
// actually had one.
type Frontmatter struct {
	// Present is false when the file had no leading `---` block at all. The
	// scanner still emits an entry in this case; the validator decides
	// whether that's an error.
	Present bool

	// Fields is the parsed YAML mapping. Nil when Present is false.
	Fields map[string]any

	// Body is the rest of the file after the closing `---`. Kept for
	// completeness; the indexer itself only needs Fields.
	Body []byte
}

var (
	openDelim  = []byte("---\n")
	closeDelim = []byte("\n---\n")
)

// ParseFrontmatter extracts a leading YAML block delimited by `---` lines.
//
// Rules:
//   - The file must start with `---\n` (after an optional UTF-8 BOM).
//   - The block ends at the next `\n---\n` (or `\n---` at EOF).
//   - The block content must parse as a YAML mapping. A scalar or sequence
//     block is rejected: it cannot represent named fields.
//   - A file without a leading `---\n` returns Frontmatter{Present:false} and
//     no error — the caller decides whether absence is fatal.
func ParseFrontmatter(data []byte) (Frontmatter, error) {
	body := stripBOM(data)
	if !bytes.HasPrefix(body, openDelim) {
		return Frontmatter{Present: false, Body: body}, nil
	}
	after := body[len(openDelim):]
	var raw, rest []byte
	switch {
	case bytes.HasPrefix(after, []byte("---\n")):
		// `---\n---\n…` — empty block, closed immediately.
		raw = nil
		rest = after[len("---\n"):]
	case bytes.Equal(after, []byte("---")):
		// `---\n---` at EOF — empty block, no trailing newline.
		raw = nil
		rest = nil
	default:
		if idx := bytes.Index(after, closeDelim); idx >= 0 {
			raw = after[:idx]
			rest = after[idx+len(closeDelim):]
		} else if bytes.HasSuffix(after, []byte("\n---")) {
			raw = after[:len(after)-len("\n---")]
			rest = nil
		} else {
			return Frontmatter{}, errors.New("frontmatter: opening `---` without matching closing delimiter")
		}
	}
	var fields map[string]any
	if err := yaml.Unmarshal(raw, &fields); err != nil {
		return Frontmatter{}, fmt.Errorf("frontmatter: parse YAML: %w", err)
	}
	if fields == nil {
		// Empty frontmatter block. Treat as present-but-empty so the caller
		// can distinguish "no block" from "empty block".
		fields = map[string]any{}
	}
	return Frontmatter{Present: true, Fields: fields, Body: rest}, nil
}

// stripBOM removes a UTF-8 byte-order mark if present.
func stripBOM(b []byte) []byte {
	bom := []byte{0xEF, 0xBB, 0xBF}
	if bytes.HasPrefix(b, bom) {
		return b[len(bom):]
	}
	return b
}

// String returns the value of key as a non-empty string, or "" if missing.
// It accepts the YAML representations that yaml.v3 emits for string scalars.
func (fm Frontmatter) String(key string) string {
	if !fm.Present {
		return ""
	}
	v, ok := fm.Fields[key]
	if !ok {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

// Bool returns the value of key as a bool. The second return reports whether
// the key was present and bool-typed.
func (fm Frontmatter) Bool(key string) (value bool, ok bool) {
	if !fm.Present {
		return false, false
	}
	v, exists := fm.Fields[key]
	if !exists {
		return false, false
	}
	b, isBool := v.(bool)
	if !isBool {
		return false, false
	}
	return b, true
}
