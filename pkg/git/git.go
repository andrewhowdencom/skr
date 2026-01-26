package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ChangedFiles returns a list of files changed between HEAD and baseRef
func ChangedFiles(baseRef string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", baseRef)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff failed: %s (%w)", out.String(), err)
	}

	lines := strings.Split(out.String(), "\n")
	var files []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			files = append(files, strings.TrimSpace(line))
		}
	}
	return files, nil
}
