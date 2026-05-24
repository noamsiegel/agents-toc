package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/noamsiegel/agents-toc/internal/config"
)

// Entry is one indexable file plus its resolved metadata. Renderer and
// validator both consume it.
type Entry struct {
	// Path is the project-relative path, with forward slashes.
	Path string

	// AbsPath is the absolute on-disk path.
	AbsPath string

	// Label is copied from the matching SourceConfig.
	Label string

	// Name is the entry's name, resolved per Source.Frontmatter.NameField.
	// Falls back to the file's basename (without extension) when missing.
	Name string

	// Summary is the entry's one-line summary, resolved per
	// Source.Frontmatter.SummaryField. May be empty.
	Summary string

	// Disabled is true when the file's frontmatter has `enabled: false` and
	// the source respects that flag.
	Disabled bool

	// HasFrontmatter is true when the file had a leading `---` block.
	HasFrontmatter bool

	// Frontmatter is the parsed block. Used by `validate` to report missing
	// required fields.
	Frontmatter Frontmatter

	// SourceIndex is the index into Config.Sources that matched this file.
	SourceIndex int

	// ModTime is the file's modification time, used by the mtime sort order.
	ModTime time.Time
}

// Scan walks the configured sources under root, parses each candidate's
// frontmatter, and returns the resulting entries.
//
// Files are deduplicated: if multiple sources match the same path, only the
// first source's interpretation is kept (sources are evaluated in order).
//
// Ignore globs are evaluated against the project-relative path. A match
// excludes the file from every source.
func Scan(root string, cfg *config.Config) ([]Entry, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}
	seen := map[string]bool{}
	var out []Entry
	for i, src := range cfg.Sources {
		matches, err := walkSource(root, src.Glob)
		if err != nil {
			return nil, fmt.Errorf("source %d (%s): %w", i, src.Glob, err)
		}
		for _, m := range matches {
			rel := toRel(root, m)
			if seen[rel] {
				continue
			}
			if matchesAny(cfg.Ignore, rel) {
				continue
			}
			entry, err := readEntry(m, rel, src, i)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", rel, err)
			}
			seen[rel] = true
			out = append(out, entry)
		}
	}
	return out, nil
}

// walkSource expands a single doublestar glob against root and returns the
// matching files in deterministic order.
func walkSource(root, pattern string) ([]string, error) {
	fsRoot := os.DirFS(root)
	matches, err := doublestar.Glob(fsRoot, pattern)
	if err != nil {
		return nil, err
	}
	abs := make([]string, 0, len(matches))
	for _, m := range matches {
		full := filepath.Join(root, m)
		st, err := os.Stat(full)
		if err != nil {
			continue
		}
		if st.IsDir() {
			continue
		}
		abs = append(abs, full)
	}
	sort.Strings(abs)
	return abs, nil
}

// matchesAny reports whether any glob matches rel. Uses doublestar matching
// semantics; rel must use forward slashes.
func matchesAny(globs []string, rel string) bool {
	for _, g := range globs {
		ok, err := doublestar.Match(g, rel)
		if err == nil && ok {
			return true
		}
	}
	return false
}

// toRel converts an absolute path to project-relative with forward slashes.
func toRel(root, abs string) string {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return abs
	}
	return filepath.ToSlash(rel)
}

// readEntry reads one matched file, parses frontmatter, and builds an Entry.
func readEntry(abs, rel string, src config.SourceConfig, srcIdx int) (Entry, error) {
	data, err := os.ReadFile(abs)
	if err != nil {
		return Entry{}, err
	}
	fm, err := ParseFrontmatter(data)
	if err != nil {
		return Entry{}, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return Entry{}, err
	}
	entry := Entry{
		Path:           rel,
		AbsPath:        abs,
		Label:          src.Label,
		Frontmatter:    fm,
		HasFrontmatter: fm.Present,
		SourceIndex:    srcIdx,
		ModTime:        st.ModTime(),
	}
	if fm.Present {
		if src.Frontmatter.NameField != "" {
			entry.Name = fm.String(src.Frontmatter.NameField)
		}
		if src.Frontmatter.SummaryField != "" {
			entry.Summary = fm.String(src.Frontmatter.SummaryField)
		}
		if src.Frontmatter.RespectDisabled != nil && *src.Frontmatter.RespectDisabled {
			if enabled, ok := fm.Bool("enabled"); ok && !enabled {
				entry.Disabled = true
			}
		}
	}
	if entry.Name == "" {
		entry.Name = basenameNoExt(rel)
	}
	return entry, nil
}

// basenameNoExt returns the path's basename minus its extension.
func basenameNoExt(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// EntryFilter returns the subset of entries that should be rendered.
// disabled entries are dropped unless includeDisabled is true.
func EntryFilter(in []Entry, includeDisabled bool) []Entry {
	out := make([]Entry, 0, len(in))
	for _, e := range in {
		if e.Disabled && !includeDisabled {
			continue
		}
		out = append(out, e)
	}
	return out
}

// SortEntries sorts entries in place per the named strategy. Unknown strategy
// is a no-op (config validator already rejects unknown values).
func SortEntries(in []Entry, strategy string) {
	switch strategy {
	case "mtime":
		sort.SliceStable(in, func(i, j int) bool {
			if !in[i].ModTime.Equal(in[j].ModTime) {
				return in[i].ModTime.After(in[j].ModTime)
			}
			return in[i].Path < in[j].Path
		})
	default: // alphabetical
		sort.SliceStable(in, func(i, j int) bool {
			return in[i].Path < in[j].Path
		})
	}
}