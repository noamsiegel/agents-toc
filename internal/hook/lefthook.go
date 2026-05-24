package hook

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// installLefthook adds (or refreshes) a pre-commit > commands > agents-toc-sync
// block in lefthook.yml.
//
// We parse to a generic map (not a typed struct) because lefthook's schema is
// rich and we want to leave every other field untouched. Yes, this drops YAML
// comments — that's the documented behavior of yaml.v3 round-trip. We pay
// that cost because the alternative (string-splicing a stranger's YAML) is
// worse.
func installLefthook(projectRoot, command string, force bool) (*Result, error) {
	path := filepath.Join(projectRoot, "lefthook.yml")
	if _, err := os.Stat(path); err != nil {
		// Fall back to .yaml suffix.
		alt := filepath.Join(projectRoot, "lefthook.yaml")
		if _, err2 := os.Stat(alt); err2 == nil {
			path = alt
		} else {
			return nil, fmt.Errorf("lefthook config not found at %s or %s", path, alt)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if doc == nil {
		doc = map[string]any{}
	}

	preCommit, _ := doc["pre-commit"].(map[string]any)
	if preCommit == nil {
		preCommit = map[string]any{}
		doc["pre-commit"] = preCommit
	}
	commands, _ := preCommit["commands"].(map[string]any)
	if commands == nil {
		commands = map[string]any{}
		preCommit["commands"] = commands
	}
	existing, ok := commands[JobName].(map[string]any)
	if ok {
		if existing["run"] == command && !force {
			return &Result{Manager: ManagerLefthook, Path: path, AlreadyPresent: true}, nil
		}
	}
	commands[JobName] = map[string]any{
		"run":         command,
		"stage_fixed": true,
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return nil, err
	}
	return &Result{Manager: ManagerLefthook, Path: path, Modified: true}, nil
}
