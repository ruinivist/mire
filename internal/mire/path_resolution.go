package mire

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveRecordTarget(testDir, path string) (string, error) {
	return resolvePathWithinTestDir(testDir, path, "record")
}

func resolvePathWithinTestDir(testDir, path, pathType string) (string, error) {
	absTestDir, err := filepath.Abs(testDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve test directory path: %v", err)
	}

	ensureWithinTestDir := func(candidate string) (string, bool) {
		relToTestDir, relErr := filepath.Rel(absTestDir, candidate)
		if relErr != nil || !isWithinBase(relToTestDir) {
			return "", false
		}

		return candidate, true
	}

	if filepath.IsAbs(path) {
		absPath := filepath.Clean(path)
		if candidate, ok := ensureWithinTestDir(absPath); ok {
			return candidate, nil
		}

		return "", fmt.Errorf("%s path %q must be inside test directory %q", pathType, path, absTestDir)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s path: %v", pathType, err)
	}

	if candidate, ok := ensureWithinTestDir(absPath); ok {
		return candidate, nil
	}

	candidate := filepath.Clean(filepath.Join(absTestDir, path))
	if candidate, ok := ensureWithinTestDir(candidate); ok {
		return candidate, nil
	}

	return "", fmt.Errorf("%s path %q must be inside test directory %q", pathType, path, absTestDir)
}

func isWithinBase(rel string) bool {
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}
