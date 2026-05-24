// Package config loads and validates .agentsmdrc.yaml.
//
// Defaults are applied so a totally empty file still produces a working
// agents-toc setup: scan ./skills/**/SKILL.md (Anthropic-style frontmatter)
// and ./knowledge/**/*.md (id/summary frontmatter), write into AGENTS.md.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SchemaVersion is the highest .agentsmdrc.yaml `version:` value this binary
// understands. Bump when introducing a breaking config change.
const SchemaVersion = 1

// DefaultConfigName is the filename agents-toc looks for.
const DefaultConfigName = ".agentsmdrc.yaml"

// AltConfigName is an accepted shortened filename.
const AltConfigName = ".agentsmdrc.yml"

// Config is the parsed .agentsmdrc.yaml.
type Config struct {
	// Version of the config schema. Required for forward compatibility.
	Version int `yaml:"version"`

	// Sources is the ordered list of file groups to index.
	Sources []SourceConfig `yaml:"sources"`

	// Target describes the AGENTS.md file and its marker pair.
	Target TargetConfig `yaml:"target"`

	// Render controls the appearance of the generated block.
	Render RenderConfig `yaml:"render"`

	// Ignore is a list of doublestar globs evaluated against every candidate
	// path. A match excludes the file from indexing.
	Ignore []string `yaml:"ignore"`

	// path is the absolute path the config was loaded from; populated by Load.
	path string `yaml:"-"`
}

// SourceConfig describes one indexable file group.
type SourceConfig struct {
	// Glob is a doublestar pattern, relative to the project root.
	Glob string `yaml:"glob"`

	// Label is the short tag the renderer uses to group entries.
	Label string `yaml:"label"`

	// Frontmatter tells the scanner how to extract name/summary from this
	// source's markdown files.
	Frontmatter FrontmatterConfig `yaml:"frontmatter"`
}

// FrontmatterConfig pins which YAML keys map to the generic name/summary slots.
type FrontmatterConfig struct {
	// NameField is the YAML key that holds the entry name. If empty, scanner
	// falls back to the file's basename.
	NameField string `yaml:"name_field"`

	// SummaryField is the YAML key that holds the one-line description.
	SummaryField string `yaml:"summary_field"`

	// Require is the list of frontmatter fields that must be present and
	// non-empty. `validate` reports missing fields as errors.
	Require []string `yaml:"require"`

	// RespectDisabled, when true, skips files whose frontmatter contains
	// `enabled: false`. Default true.
	RespectDisabled *bool `yaml:"respect_disabled"`
}

// TargetConfig is the AGENTS.md file plus its fenced markers.
type TargetConfig struct {
	File            string `yaml:"file"`
	MarkerStart     string `yaml:"marker_start"`
	MarkerEnd       string `yaml:"marker_end"`
	CreateIfMissing *bool  `yaml:"create_if_missing"`
}

// RenderConfig controls how entries are rendered between the markers.
type RenderConfig struct {
	// GroupBy is "label" (group entries by source label) or "none" (flat list).
	GroupBy string `yaml:"group_by"`

	// Sort is "alphabetical" (by path) or "mtime" (newest first).
	Sort string `yaml:"sort"`

	// LineTemplate is the per-entry template. Placeholders:
	//   {path}     project-relative path
	//   {name}     resolved entry name
	//   {summary}  one-line summary
	//   {label}    source label
	LineTemplate string `yaml:"line_template"`

	// GroupTemplate wraps lines in a group when GroupBy=="label". Placeholders:
	//   {label}    source label
	//   {lines}    rendered lines joined with newline
	//   {count}    number of entries in the group
	GroupTemplate string `yaml:"group_template"`

	// EmptyMessage replaces the body when nothing matches.
	EmptyMessage string `yaml:"empty_message"`

	// IncludeDisabled forces disabled entries to be rendered. CLI
	// --include-disabled overrides this at runtime.
	IncludeDisabled bool `yaml:"include_disabled"`
}

// boolPtr returns a pointer to b. Used to fill *bool defaults.
func boolPtr(b bool) *bool { return &b }

// Default returns a Config populated with the sensible default behavior:
// scan ./skills/**/SKILL.md (name/description) and ./knowledge/**/*.md
// (id/summary), respect `enabled: false`, group by label.
func Default() Config {
	return Config{
		Version: SchemaVersion,
		Sources: []SourceConfig{
			{
				Glob:  "skills/**/SKILL.md",
				Label: "skill",
				Frontmatter: FrontmatterConfig{
					NameField:       "name",
					SummaryField:    "description",
					RespectDisabled: boolPtr(true),
				},
			},
			{
				Glob:  "knowledge/**/*.md",
				Label: "knowledge",
				Frontmatter: FrontmatterConfig{
					NameField:       "id",
					SummaryField:    "summary",
					Require:         []string{"id", "summary"},
					RespectDisabled: boolPtr(true),
				},
			},
		},
		Target: TargetConfig{
			File:            "AGENTS.md",
			MarkerStart:     "<!-- INDEX:START -->",
			MarkerEnd:       "<!-- INDEX:END -->",
			CreateIfMissing: boolPtr(true),
		},
		Render: RenderConfig{
			GroupBy:       "label",
			Sort:          "alphabetical",
			LineTemplate:  "- `{path}` — {summary}",
			GroupTemplate: "### {label}/\n\n{lines}\n",
			EmptyMessage:  "_no entries_",
		},
		Ignore: []string{
			"**/node_modules/**",
			"**/.git/**",
			"**/vendor/**",
			"**/.venv/**",
		},
	}
}

// Locate searches upward from start for a config file (.agentsmdrc.yaml or
// .agentsmdrc.yml), stopping at the filesystem root. Returns the absolute
// path of the first match, or empty string if none found.
func Locate(start string) string {
	dir, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	for {
		for _, name := range []string{DefaultConfigName, AltConfigName} {
			candidate := filepath.Join(dir, name)
			if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
				return candidate
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Load reads and parses the config at path, applies defaults to zero fields,
// and validates schema version. The returned Config has its Path() populated.
func Load(path string) (*Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", abs, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", abs, err)
	}
	if cfg.Version == 0 {
		return nil, fmt.Errorf("%s: missing required field `version`", abs)
	}
	if cfg.Version > SchemaVersion {
		return nil, fmt.Errorf("%s: schema version %d is newer than this binary supports (%d); run `agents-toc migrate` after upgrading", abs, cfg.Version, SchemaVersion)
	}
	cfg.applyDefaults()
	cfg.path = abs
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", abs, err)
	}
	return &cfg, nil
}

// Path returns the absolute path the config was loaded from.
func (c *Config) Path() string { return c.path }

// Root returns the project root, defined as the directory containing the
// config file.
func (c *Config) Root() string { return filepath.Dir(c.path) }

// Validate reports structural errors. Defaults must already be applied.
func (c *Config) Validate() error {
	if len(c.Sources) == 0 {
		return errors.New("at least one `sources` entry is required")
	}
	for i, src := range c.Sources {
		if src.Glob == "" {
			return fmt.Errorf("sources[%d]: `glob` is required", i)
		}
		if src.Label == "" {
			return fmt.Errorf("sources[%d]: `label` is required", i)
		}
	}
	if c.Target.File == "" {
		return errors.New("target.file is required")
	}
	if c.Target.MarkerStart == "" || c.Target.MarkerEnd == "" {
		return errors.New("target.marker_start and target.marker_end are required")
	}
	if c.Target.MarkerStart == c.Target.MarkerEnd {
		return errors.New("target.marker_start and target.marker_end must differ")
	}
	switch c.Render.GroupBy {
	case "label", "none":
	default:
		return fmt.Errorf("render.group_by must be `label` or `none`, got %q", c.Render.GroupBy)
	}
	switch c.Render.Sort {
	case "alphabetical", "mtime":
	default:
		return fmt.Errorf("render.sort must be `alphabetical` or `mtime`, got %q", c.Render.Sort)
	}
	return nil
}

// applyDefaults fills zero-valued fields from Default(). Slices and maps that
// are nil are filled wholesale; primitives use their zero value as the
// "unset" sentinel.
func (c *Config) applyDefaults() {
	d := Default()
	if c.Target.File == "" {
		c.Target.File = d.Target.File
	}
	if c.Target.MarkerStart == "" {
		c.Target.MarkerStart = d.Target.MarkerStart
	}
	if c.Target.MarkerEnd == "" {
		c.Target.MarkerEnd = d.Target.MarkerEnd
	}
	if c.Target.CreateIfMissing == nil {
		c.Target.CreateIfMissing = d.Target.CreateIfMissing
	}
	if c.Render.GroupBy == "" {
		c.Render.GroupBy = d.Render.GroupBy
	}
	if c.Render.Sort == "" {
		c.Render.Sort = d.Render.Sort
	}
	if c.Render.LineTemplate == "" {
		c.Render.LineTemplate = d.Render.LineTemplate
	}
	if c.Render.GroupTemplate == "" {
		c.Render.GroupTemplate = d.Render.GroupTemplate
	}
	if c.Render.EmptyMessage == "" {
		c.Render.EmptyMessage = d.Render.EmptyMessage
	}
	if len(c.Ignore) == 0 {
		c.Ignore = d.Ignore
	}
	for i := range c.Sources {
		fm := &c.Sources[i].Frontmatter
		if fm.RespectDisabled == nil {
			fm.RespectDisabled = boolPtr(true)
		}
	}
}

// Marshal returns a YAML serialization of the config suitable for `init` to
// write to disk. The output is canonical: stable key order, two-space indent.
func (c *Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}
