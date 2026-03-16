package miro

import (
	"fmt"
	"os"
	"path/filepath"
)

// Record creates the requested directory path under the resolved test directory.
func Record(path string) (string, error) {
	testDir, err := ResolveTestDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve test directory: %v", err)
	}

	target := filepath.Join(testDir, path)
	if err := os.MkdirAll(target, 0o755); err != nil {
		return "", err
	}

	return target, nil
}
