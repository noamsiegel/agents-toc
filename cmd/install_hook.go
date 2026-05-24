package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/noamsiegel/agents-toc/internal/hook"
)

var (
	hookManager string
	hookForce   bool
)

var installHookCmd = &cobra.Command{
	Use:   "install-hook",
	Short: "Wire pre-commit to run `agents-toc sync` automatically.",
	Long: `install-hook detects your hook manager (lefthook > husky > raw .git/hooks)
and installs a single pre-commit entry that runs ` + "`agents-toc sync`" + ` and
restages AGENTS.md. Running it twice is a no-op. Use --manager to force a
specific adapter.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := resolveConfig()
		if err != nil {
			return err
		}
		root := cfg.Root()
		if _, err := os.Stat(root); err != nil {
			return fmt.Errorf("project root %s: %w", root, err)
		}
		mgr := hook.Manager(hookManager)
		if mgr == "" {
			mgr = hook.ManagerAuto
		}
		res, err := hook.Install(root, mgr, cfg.Target.File, hookForce)
		if err != nil {
			if errors.Is(err, hook.ErrAlreadyInstalled) {
				fmt.Fprintf(cmd.OutOrStdout(), "agents-toc hook already installed via %s\n", res.Manager)
				return nil
			}
			return err
		}
		out := cmd.OutOrStdout()
		switch {
		case res.AlreadyPresent:
			fmt.Fprintf(out, "agents-toc hook already present in %s\n", res.Path)
		case res.Created:
			fmt.Fprintf(out, "installed %s hook (created %s)\n", res.Manager, res.Path)
		case res.Modified:
			fmt.Fprintf(out, "installed %s hook (updated %s)\n", res.Manager, res.Path)
		}
		return nil
	},
}

func init() {
	installHookCmd.Flags().StringVar(&hookManager, "manager", "auto", "hook manager: auto | lefthook | husky | raw")
	installHookCmd.Flags().BoolVar(&hookForce, "force", false, "overwrite an existing agents-toc hook entry")
	rootCmd.AddCommand(installHookCmd)
}
