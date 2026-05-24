package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/noamsiegel/agents-toc/internal/buildinfo"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the agents-toc version, commit, and build date.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "agents-toc %s (commit %s, built %s)\n",
			buildinfo.Version, buildinfo.Commit, buildinfo.Date)
		return err
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
