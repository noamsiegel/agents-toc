// Package pipeline orchestrates the end-to-end sync: scan → filter → sort →
// render → splice into target. It is the single place CLI subcommands call.
package pipeline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/noamsiegel/agents-toc/internal/config"
	"github.com/noamsiegel/agents-toc/internal/render"
	"github.com/noamsiegel/agents-toc/internal/scan"
	"github.com/noamsiegel/agents-toc/internal/target"
)

// SyncResult reports what happened during a Sync call.
type SyncResult struct {
	// TargetPath is the absolute path of the AGENTS.md file written.
	TargetPath string

	// EntryCount is the number of entries rendered (after filtering).
	EntryCount int

	// SkippedDisabled is the count of entries hidden by `enabled: false`.
	SkippedDisabled int

	// Changed reports whether the file's content actually changed. False
	// means the sync was a no-op (idempotent re-run).
	Changed bool

	// Created reports whether the AGENTS.md file was created from scratch
	// by this call. Implies Changed=true.
	Created bool

	// MarkersInserted reports whether the markers were inserted into a file
	// that previously had none. Implies Changed=true.
	MarkersInserted bool
}

// Options modify a Sync call's behavior.
type Options struct {
	// IncludeDisabled forces files with `enabled: false` to be rendered.
	IncludeDisabled bool

	// Version is the agents-toc version string stamped into the generated
	// header comment.
	Version string

	// Check, when true, computes what Sync would do without writing.
	// SyncResult.Changed reports whether the file would have changed.
	// Suitable for CI gates like `agents-toc sync --check` in a workflow.
	Check bool
}

// Sync executes one full sync pass.
func Sync(cfg *config.Config, opt Options) (*SyncResult, error) {
	root := cfg.Root()
	entries, err := scan.Scan(root, cfg)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	skippedDisabled := 0
	for _, e := range entries {
		if e.Disabled {
			skippedDisabled++
		}
	}
	includeDisabled := opt.IncludeDisabled || cfg.Render.IncludeDisabled
	filtered := scan.EntryFilter(entries, includeDisabled)
	scan.SortEntries(filtered, cfg.Render.Sort)

	block := render.Block(filtered, cfg, opt.Version)

	targetPath := filepath.Join(root, cfg.Target.File)
	res := &SyncResult{
		TargetPath:      targetPath,
		EntryCount:      len(filtered),
		SkippedDisabled: skippedDisabled,
	}

	existing, err := os.ReadFile(targetPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if cfg.Target.CreateIfMissing != nil && !*cfg.Target.CreateIfMissing {
			return nil, fmt.Errorf("%s does not exist and target.create_if_missing is false", targetPath)
		}
		if opt.Check {
			res.Created = true
			res.Changed = true
			return res, nil
		}
		body := target.ScaffoldFile(block, filepath.Base(root))
		if err := target.WriteAtomic(targetPath, []byte(body), 0o644); err != nil {
			return nil, err
		}
		res.Created = true
		res.Changed = true
		return res, nil
	case err != nil:
		return nil, fmt.Errorf("read target: %w", err)
	}

	updated, err := target.Replace(string(existing), cfg.Target.MarkerStart, cfg.Target.MarkerEnd, block)
	if err != nil {
		if errors.Is(err, target.ErrMissingMarkers) {
			if cfg.Target.CreateIfMissing != nil && !*cfg.Target.CreateIfMissing {
				return nil, fmt.Errorf("%s has no marker pair and target.create_if_missing is false", targetPath)
			}
			updated = target.Insert(string(existing), block)
			res.MarkersInserted = true
		} else {
			return nil, fmt.Errorf("locate markers in %s: %w", targetPath, err)
		}
	}

	if updated == string(existing) {
		return res, nil
	}
	if opt.Check {
		res.Changed = true
		return res, nil
	}
	if err := target.WriteAtomic(targetPath, []byte(updated), 0o644); err != nil {
		return nil, err
	}
	res.Changed = true
	return res, nil
}

// ListResult is the structured output of `agents-toc list`.
type ListResult struct {
	Entries []scan.Entry
	Skipped []scan.Entry
}

// List returns the full unfiltered entry set plus the subset hidden by
// `enabled: false`. It performs the same scan + sort as Sync but does not
// touch the target file.
func List(cfg *config.Config) (*ListResult, error) {
	entries, err := scan.Scan(cfg.Root(), cfg)
	if err != nil {
		return nil, err
	}
	scan.SortEntries(entries, cfg.Render.Sort)
	var visible, skipped []scan.Entry
	for _, e := range entries {
		if e.Disabled {
			skipped = append(skipped, e)
			continue
		}
		visible = append(visible, e)
	}
	return &ListResult{Entries: visible, Skipped: skipped}, nil
}

// ValidationIssue is one frontmatter problem reported by `agents-toc validate`.
type ValidationIssue struct {
	Path    string
	Kind    string // "missing-field", "missing-frontmatter", "duplicate-name", "broken-path"
	Field   string
	Message string
}

// Validate runs structural checks on the scanned entries.
//
// Checks:
//   - every source's `require` fields are present and non-empty.
//   - entries without a frontmatter block at all are flagged when the source
//     declares any required fields.
//   - duplicate `name` values within the same label are flagged.
func Validate(cfg *config.Config) ([]ValidationIssue, error) {
	entries, err := scan.Scan(cfg.Root(), cfg)
	if err != nil {
		return nil, err
	}
	var issues []ValidationIssue

	for _, e := range entries {
		src := cfg.Sources[e.SourceIndex]
		required := src.Frontmatter.Require
		if len(required) == 0 {
			continue
		}
		if !e.HasFrontmatter {
			issues = append(issues, ValidationIssue{
				Path:    e.Path,
				Kind:    "missing-frontmatter",
				Message: "source requires frontmatter but file has none",
			})
			continue
		}
		for _, field := range required {
			if e.Frontmatter.String(field) == "" {
				issues = append(issues, ValidationIssue{
					Path:    e.Path,
					Kind:    "missing-field",
					Field:   field,
					Message: fmt.Sprintf("required field %q is missing or empty", field),
				})
			}
		}
	}

	// Duplicate name within label.
	seen := map[string]string{}
	for _, e := range entries {
		key := e.Label + "\x00" + e.Name
		if prev, ok := seen[key]; ok {
			issues = append(issues, ValidationIssue{
				Path:    e.Path,
				Kind:    "duplicate-name",
				Field:   "name",
				Message: fmt.Sprintf("duplicate name %q within label %q (also at %s)", e.Name, e.Label, prev),
			})
			continue
		}
		seen[key] = e.Path
	}
	return issues, nil
}
