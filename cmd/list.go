package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/noamsiegel/agents-toc/internal/pipeline"
	"github.com/noamsiegel/agents-toc/internal/scan"
)

var (
	listJSON     bool
	listShowAll  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print every file that would be indexed and why.",
	Long: `list runs the same scan as sync but does not write anything. By default
disabled entries are reported separately at the end. Use --json for machine
consumption.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := resolveConfig()
		if err != nil {
			return err
		}
		res, err := pipeline.List(cfg)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if listJSON {
			payload := map[string]any{
				"entries": shrinkEntries(res.Entries),
			}
			if listShowAll {
				payload["skipped"] = shrinkEntries(res.Skipped)
			}
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			return enc.Encode(payload)
		}
		for _, e := range res.Entries {
			fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", e.Label, e.Path, e.Name, e.Summary)
		}
		if listShowAll && len(res.Skipped) > 0 {
			fmt.Fprintln(out, "--- skipped (disabled) ---")
			for _, e := range res.Skipped {
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", e.Label, e.Path, e.Name, e.Summary)
			}
		}
		return nil
	},
}

// shrinkEntries returns a JSON-friendly view of entries (drops parsed
// frontmatter blob and absolute paths).
func shrinkEntries(in []scan.Entry) []map[string]any {
	out := make([]map[string]any, len(in))
	for i, e := range in {
		out[i] = map[string]any{
			"path":     e.Path,
			"label":    e.Label,
			"name":     e.Name,
			"summary":  e.Summary,
			"disabled": e.Disabled,
		}
	}
	return out
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "emit machine-readable JSON")
	listCmd.Flags().BoolVar(&listShowAll, "include-disabled", false, "also report files hidden by enabled: false")
	rootCmd.AddCommand(listCmd)
}
