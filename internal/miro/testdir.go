package miro

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	miroconfig "miro/internal/config"
)

// ResolveTestDir resolves the project test directory from config.
func ResolveTestDir() (string, error) {
	root, err := currentProjectRoot()
	if err != nil {
		return "", err
	}

	return resolveTestDirFromRoot(root)
}

func resolveTestDirFromRoot(root string) (string, error) {
	cfg, err := readConfigFromRoot(root)
	if err != nil {
		return "", err
	}

	return resolveTestDirFromConfig(root, cfg)
}

func readConfigFromRoot(root string) (miroconfig.Config, error) {
	configPath := filepath.Join(root, "miro.toml")
	return miroconfig.ReadConfig(configPath)
}

func resolveTestDirFromConfig(root string, cfg miroconfig.Config) (string, error) {
	testDir := filepath.Join(root, cfg.TestDir)
	info, err := os.Stat(testDir)
	if err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("configured test_dir is not a directory: %s", testDir)
		}
		return testDir, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to check test directory %s: %v", testDir, err)
	}

	return testDir, nil
}

func currentProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %v", err)
	}

	return projectRoot(cwd), nil
}

// returns the git root if in a git repo else current pwd
func projectRoot(cwd string) string {
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return cwd
	}

	root := strings.TrimSpace(string(out))
	if root == "" {
		return cwd
	}
	return filepath.Clean(root)
}
