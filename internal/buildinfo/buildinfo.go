// Package buildinfo carries CLI build metadata stamped at link time.
package buildinfo

// Version of the agents-toc binary. Overridden via -ldflags by goreleaser.
var Version = "dev"

// Commit hash of the build. Overridden via -ldflags.
var Commit = "none"

// Date of the build. Overridden via -ldflags.
var Date = "unknown"
