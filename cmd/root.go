// Package cmd wires the agents-toc CLI surface using cobra.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/noamsiegel/agents-toc/internal/buildinfo"
)

var rootCmd = &cobra.Command{
	Use:   "agents-toc",
	Short: "Keep a one-line lazy-load index inside AGENTS.md in sync.",
	Version: buildinfo.Version,
	Long: `agents-toc scans configurable sources (SKILL.md bundles, knowledge
markdown, anything you point it at) and maintains a fenced index block inside
AGENTS.md. It owns only the bytes between its markers; everything else in the
file is yours.

Typical use:

  agents-toc init           # scaffold .agentsmdrc.yaml and ensure markers
  agents-toc sync           # regenerate the index block (idempotent)
  agents-toc install-hook   # wire pre-commit to run sync automatically
`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// configPathFlag is the resolved config path, shared across subcommands.
var configPathFlag string

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPathFlag, "config", "c", "", "path to .agentsmdrc.yaml (default: search from CWD upward)")
	rootCmd.SetVersionTemplate("agents-toc {{.Version}}\n")
}

// Execute runs the root command and returns any error.
func Execute() error {
	return rootCmd.Execute()
}
