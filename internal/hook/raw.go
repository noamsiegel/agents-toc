package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// installRaw writes (or extends) .git/hooks/pre-commit with a fenced managed
// block. The fence lets `uninstall` (future) remove our lines cleanly without
// destroying anything else.
func installRaw(projectRoot, command string, force bool) (*Result, error) {
	path := filepath.Join(projectRoot, ".git", "hooks", "pre-commit")

	created := false
	body := ""
	if data, err := os.ReadFile(path); err == nil {
		body = string(data)
	} else if os.IsNotExist(err) {
		created = true
	} else {
		return nil, err
	}

	managed := rawManagedBlock(command)
	if strings.Contains(body, "# >>> agents-toc-sync >>>") {
		if !force {
			return &Result{Manager: ManagerRaw, Path: path, AlreadyPresent: true}, nil
		}
		body = stripManaged(body, "# >>> agents-toc-sync >>>", "# <<< agents-toc-sync <<<")
	}

	if created {
		body = "#!/usr/bin/env sh\nset -e\n\n" + managed + "\n"
	} else {
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		body += managed + "\n"
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir hooks: %w", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		return nil, err
	}
	_ = os.Chmod(path, 0o755)

	return &Result{
		Manager:  ManagerRaw,
		Path:     path,
		Created:  created,
		Modified: !created,
	}, nil
}

func rawManagedBlock(command string) string {
	return fmt.Sprintf(`# >>> agents-toc-sync >>>
%s
# <<< agents-toc-sync <<<`, command)
}
