package miro

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Record creates the requested directory path under the resolved test directory.
func Record(path string) (string, error) {
	testDir, err := ResolveTestDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve test directory: %v", err)
	}

	target, err := resolveRecordTarget(testDir, path)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(target, 0o755); err != nil {
		return "", err
	}

	return target, nil
}

func resolveRecordTarget(testDir, path string) (string, error) {
	absTestDir, err := filepath.Abs(testDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve test directory path: %v", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve record path: %v", err)
	}

	// path rel to testDir
	relToTestDir, err := filepath.Rel(absTestDir, absPath)
	if err == nil && isWithinBase(relToTestDir) {
		return filepath.Join(absTestDir, relToTestDir), nil
	}

	// not rel to test dir and absolute path given
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("record path %q must be inside test directory %q", path, absTestDir)
	}

	return filepath.Join(absTestDir, path), nil
}

// is the rel path going "outside"
func isWithinBase(rel string) bool {
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}
