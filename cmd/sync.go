package cmd

import (
	"errors"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/noamsiegel/agents-toc/internal/buildinfo"
	"github.com/noamsiegel/agents-toc/internal/config"
	"github.com/noamsiegel/agents-toc/internal/pipeline"
)

// errCheckDrift is returned by `sync --check` when the target file would
// change. cobra prints the error and the process exits non-zero, suitable
// for CI gates.
var errCheckDrift = errors.New("AGENTS.md is stale; run `agents-toc sync`")

// writeJSON emits any value as JSON to w with stable indentation.
func writeJSON(w interface{ Write([]byte) (int, error) }, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

var (
	syncIncludeDisabled bool
	syncCheck           bool
	syncJSON            bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Regenerate the AGENTS.md index block from current sources.",
	Long: `sync rescans the configured sources and rewrites the fenced index block
inside AGENTS.md. Running it twice in a row is a no-op. If AGENTS.md does not
exist (and target.create_if_missing is true), it is scaffolded.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := resolveConfig()
		if err != nil {
			return err
		}
		res, err := pipeline.Sync(cfg, pipeline.Options{
			IncludeDisabled: syncIncludeDisabled,
			Version:         buildinfo.Version,
			Check:           syncCheck,
		})
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if syncJSON {
			return writeJSON(out, res)
		}
		if syncCheck {
			if res.Changed {
				fmt.Fprintf(out, "would change %s (%d entries)\n", res.TargetPath, res.EntryCount)
				return errCheckDrift
			}
			fmt.Fprintf(out, "%s already up-to-date (%d entries)\n", res.TargetPath, res.EntryCount)
			return nil
		}
		switch {
		case res.Created:
			fmt.Fprintf(out, "created %s (%d entries)\n", res.TargetPath, res.EntryCount)
		case res.MarkersInserted:
			fmt.Fprintf(out, "inserted markers in %s (%d entries)\n", res.TargetPath, res.EntryCount)
		case res.Changed:
			fmt.Fprintf(out, "updated %s (%d entries)\n", res.TargetPath, res.EntryCount)
		default:
			fmt.Fprintf(out, "%s already up-to-date (%d entries)\n", res.TargetPath, res.EntryCount)
		}
		if res.SkippedDisabled > 0 && !syncIncludeDisabled {
			fmt.Fprintf(out, "(skipped %d disabled)\n", res.SkippedDisabled)
		}
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncIncludeDisabled, "include-disabled", false, "render files marked enabled: false")
	syncCmd.Flags().BoolVar(&syncCheck, "check", false, "compute what sync would do; exit non-zero if the target file would change")
	syncCmd.Flags().BoolVar(&syncJSON, "json", false, "emit machine-readable JSON result")
	rootCmd.AddCommand(syncCmd)
}

// resolveConfig finds and loads the project config. It honors --config when
// set, otherwise searches upward from CWD.
func resolveConfig() (*config.Config, error) {
	path := configPathFlag
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = config.Locate(cwd)
		if path == "" {
			return nil, errors.New("no .agentsmdrc.yaml found in this directory or any parent — run `agents-toc init` to create one")
		}
	}
	return config.Load(path)
}
