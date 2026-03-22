package mire

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	mireconfig "mire/internal/config"
)

const defaultTestDir = "e2e"

// Init creates the default config when missing and refreshes the recorder shell.
func Init() error {
	root, err := currentProjectRoot()
	if err != nil {
		return err
	}

	configPath := filepath.Join(root, "mire.toml")
	if _, err := os.Stat(configPath); err == nil {
		if _, err := mireconfig.ReadConfig(configPath); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check %s: %v", configPath, err)
	} else {
		if err := mireconfig.WriteDefaultConfig(configPath); err != nil {
			return fmt.Errorf("failed to write %s: %v", configPath, err)
		}
	}

	testDir, err := resolveTestDirFromRoot(root)
	if err != nil {
		return err
	}

	return ensureRecordShell(testDir)
}
