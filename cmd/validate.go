package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/noamsiegel/agents-toc/internal/pipeline"
)

var validateJSON bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check frontmatter for missing required fields and duplicate names.",
	Long: `validate runs the same scan as sync and reports any structural problems:
missing required frontmatter fields, files without frontmatter where the source
requires it, and duplicate names within a label. Exits non-zero if any issues
are found, suitable for pre-commit gating.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := resolveConfig()
		if err != nil {
			return err
		}
		issues, err := pipeline.Validate(cfg)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if validateJSON {
			if err := writeJSON(out, map[string]any{
				"ok":     len(issues) == 0,
				"issues": issues,
			}); err != nil {
				return err
			}
			if len(issues) > 0 {
				return fmt.Errorf("%d validation issue(s)", len(issues))
			}
			return nil
		}
		if len(issues) == 0 {
			fmt.Fprintln(out, "ok: no issues")
			return nil
		}
		for _, i := range issues {
			if i.Field != "" {
				fmt.Fprintf(out, "%s: [%s] %s — %s\n", i.Path, i.Kind, i.Field, i.Message)
			} else {
				fmt.Fprintf(out, "%s: [%s] %s\n", i.Path, i.Kind, i.Message)
			}
		}
		return fmt.Errorf("%d validation issue(s)", len(issues))
	},
}

func init() {
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "emit machine-readable JSON")
	rootCmd.AddCommand(validateCmd)
}
