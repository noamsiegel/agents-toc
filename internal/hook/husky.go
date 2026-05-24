package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// installHusky adds the Command to .husky/pre-commit. Husky v9+ format: a
// plain shell script, no shebang required but POSIX-friendly.
//
// If the file does not exist we create it. If it exists but lacks the
// Command, we append a managed block.
func installHusky(projectRoot, command string, force bool) (*Result, error) {
	dir := filepath.Join(projectRoot, ".husky")
	path := filepath.Join(dir, "pre-commit")

	created := false
	body := ""
	if data, err := os.ReadFile(path); err == nil {
		body = string(data)
	} else if os.IsNotExist(err) {
		created = true
	} else {
		return nil, err
	}

	managed := huskyManagedBlock(command)
	if strings.Contains(body, "# >>> agents-toc-sync >>>") {
		if !force {
			return &Result{Manager: ManagerHusky, Path: path, AlreadyPresent: true}, nil
		}
		body = stripManaged(body, "# >>> agents-toc-sync >>>", "# <<< agents-toc-sync <<<")
	}

	if created {
		body = "#!/usr/bin/env sh\n" + managed + "\n"
	} else {
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		body += managed + "\n"
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		return nil, err
	}
	// Ensure exec bit even when the file pre-existed without it.
	_ = os.Chmod(path, 0o755)

	return &Result{
		Manager:  ManagerHusky,
		Path:     path,
		Created:  created,
		Modified: !created,
	}, nil
}

// huskyManagedBlock is the shell snippet wrapped in fence markers for clean
// removal/refresh.
func huskyManagedBlock(command string) string {
	return fmt.Sprintf(`# >>> agents-toc-sync >>>
%s
# <<< agents-toc-sync <<<`, command)
}

// stripManaged removes a [start, end] line-fence (inclusive) from body. If
// either fence is absent, body is returned unchanged.
func stripManaged(body, start, end string) string {
	si := strings.Index(body, start)
	if si < 0 {
		return body
	}
	ei := strings.Index(body[si:], end)
	if ei < 0 {
		return body
	}
	ei += si + len(end)
	// Also swallow the trailing newline after the end marker.
	if ei < len(body) && body[ei] == '\n' {
		ei++
	}
	// And the leading newline before the start marker, if any.
	if si > 0 && body[si-1] == '\n' {
		si--
	}
	return body[:si] + body[ei:]
}
