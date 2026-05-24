// Package hook installs a pre-commit hook that runs `agents-toc sync` and
// re-stages the target file. It supports three managers, in priority order:
// lefthook, husky, raw .git/hooks.
package hook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Manager identifies a supported hook manager.
type Manager string

const (
	ManagerAuto     Manager = "auto"
	ManagerLefthook Manager = "lefthook"
	ManagerHusky    Manager = "husky"
	ManagerRaw      Manager = "raw"
)

// BuildCommand returns the shell line every adapter installs. It includes the
// configured target file so a project that points `target.file` at something
// other than `AGENTS.md` still re-stages the right file post-sync.
func BuildCommand(targetFile string) string {
	if targetFile == "" {
		targetFile = "AGENTS.md"
	}
	return fmt.Sprintf("agents-toc sync && git add %s", targetFile)
}

// JobName is the identifier each adapter uses for its agents-toc entry. It's
// also used to detect "already installed" so install is idempotent.
const JobName = "agents-toc-sync"

// Result reports what happened.
type Result struct {
	Manager        Manager // which adapter was used
	Path           string  // path of the file the adapter wrote to
	Created        bool    // file did not exist before
	Modified       bool    // file existed and we changed it
	AlreadyPresent bool    // entry was already there; we did nothing
}

// Detect picks the best-fit manager for projectRoot. Order: lefthook > husky
// > raw. Returns ManagerRaw if no manager is detected (raw mode works on any
// git repo).
func Detect(projectRoot string) (Manager, error) {
	if _, err := os.Stat(filepath.Join(projectRoot, "lefthook.yml")); err == nil {
		return ManagerLefthook, nil
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "lefthook.yaml")); err == nil {
		return ManagerLefthook, nil
	}
	if st, err := os.Stat(filepath.Join(projectRoot, ".husky")); err == nil && st.IsDir() {
		return ManagerHusky, nil
	}
	// Fall back to raw .git/hooks. Verify we're inside a git repo.
	gitDir := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return "", fmt.Errorf("not a git repository: no .git/ under %s", projectRoot)
	}
	return ManagerRaw, nil
}

// Install wires the pre-commit hook into the chosen manager. Pass
// ManagerAuto to let Detect pick. The command shell-line is built from
// targetFile so non-default `target.file` configurations work correctly.
func Install(projectRoot string, manager Manager, targetFile string, force bool) (*Result, error) {
	if manager == "" || manager == ManagerAuto {
		m, err := Detect(projectRoot)
		if err != nil {
			return nil, err
		}
		manager = m
	}
	command := BuildCommand(targetFile)
	switch manager {
	case ManagerLefthook:
		return installLefthook(projectRoot, command, force)
	case ManagerHusky:
		return installHusky(projectRoot, command, force)
	case ManagerRaw:
		return installRaw(projectRoot, command, force)
	default:
		return nil, fmt.Errorf("unknown hook manager: %q", manager)
	}
}

// ErrAlreadyInstalled is the sentinel an adapter may wrap when force=false
// and the entry is already present. Install treats this as a successful
// no-op, returning AlreadyPresent=true on the Result.
var ErrAlreadyInstalled = errors.New("agents-toc hook already installed")
