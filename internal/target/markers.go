// Package target finds the AGENTS.md fenced marker block and replaces only
// its interior, never touching the bytes outside.
package target

import (
	"errors"
	"fmt"
	"strings"
)

// MarkerError describes a failure to locate a marker pair.
type MarkerError struct {
	Reason string
}

func (e *MarkerError) Error() string { return e.Reason }

// ErrMissingMarkers is returned when neither marker is present in the file.
// Callers may then decide to insert markers (e.g. on `init`).
var ErrMissingMarkers = &MarkerError{Reason: "neither start nor end marker found"}

// ErrUnbalancedMarkers is returned when only one of the two markers is found
// or when the end marker precedes the start marker.
var ErrUnbalancedMarkers = &MarkerError{Reason: "marker pair is unbalanced or out of order"}

// Locate returns the byte spans of the start and end marker lines within
// content. Returns ErrMissingMarkers when neither is present, or
// ErrUnbalancedMarkers when only one is present.
//
// Markers are matched as substrings of a line; surrounding whitespace within
// the line is not significant. The first occurrence of each marker wins.
func Locate(content, startMarker, endMarker string) (startStart, startEnd, endStart, endEnd int, err error) {
	si := strings.Index(content, startMarker)
	ei := strings.Index(content, endMarker)
	switch {
	case si < 0 && ei < 0:
		return 0, 0, 0, 0, ErrMissingMarkers
	case si < 0 || ei < 0 || ei <= si:
		return 0, 0, 0, 0, ErrUnbalancedMarkers
	}
	startStart = si
	startEnd = si + len(startMarker)
	endStart = ei
	endEnd = ei + len(endMarker)
	return
}

// Replace returns content with the bytes between (and including) the marker
// pair replaced by block. The block string MUST already begin with the start
// marker and end with the end marker; this function does NOT add them.
//
// Returns the new content and ErrMissingMarkers / ErrUnbalancedMarkers as in
// Locate.
func Replace(content, startMarker, endMarker, block string) (string, error) {
	ss, _, _, ee, err := Locate(content, startMarker, endMarker)
	if err != nil {
		return "", err
	}
	return content[:ss] + block + content[ee:], nil
}

// Insert appends a marker block to content, separated by a blank line if the
// file does not already end with one. Used when the file exists but has no
// markers yet.
func Insert(content, block string) string {
	if content == "" {
		return block + "\n"
	}
	var b strings.Builder
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteByte('\n')
	}
	if !strings.HasSuffix(content, "\n\n") {
		b.WriteByte('\n')
	}
	b.WriteString(block)
	b.WriteByte('\n')
	return b.String()
}

// ScaffoldFile is the minimal AGENTS.md body produced by `agents-toc init`
// when the file does not exist at all.
func ScaffoldFile(block, projectName string) string {
	name := projectName
	if name == "" {
		name = "this project"
	}
	return fmt.Sprintf(`# AGENTS.md

This file orients AI agents working on %s. The index block below is regenerated
by [agents-toc](https://github.com/noamsiegel/agents-toc); everything else is
yours to edit.

%s
`, name, block)
}

// ValidateBlock ensures block starts with startMarker and ends with endMarker.
// Useful as a sanity check before calling Replace.
func ValidateBlock(block, startMarker, endMarker string) error {
	if !strings.HasPrefix(block, startMarker) {
		return errors.New("block must begin with the configured start marker")
	}
	if !strings.HasSuffix(block, endMarker) {
		return errors.New("block must end with the configured end marker")
	}
	return nil
}
