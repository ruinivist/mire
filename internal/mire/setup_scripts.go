package mire

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	setupScriptName     = "setup.sh"
	setupScriptsEnvName = "MIRE_SETUP_SCRIPTS"
)

func discoverSetupScripts(testDir, scenarioDir string) ([]string, error) {
	absTestDir, err := filepath.Abs(testDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve test directory path: %v", err)
	}

	absScenarioDir, err := filepath.Abs(scenarioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scenario directory path: %v", err)
	}

	relToTestDir, err := filepath.Rel(absTestDir, absScenarioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scenario directory %q: %v", absScenarioDir, err)
	}
	if !isWithinBase(relToTestDir) {
		return nil, fmt.Errorf("scenario directory %q must be inside test directory %q", absScenarioDir, absTestDir)
	}

	dirs := []string{absTestDir}
	if relToTestDir != "." {
		current := absTestDir
		for _, part := range strings.Split(relToTestDir, string(os.PathSeparator)) {
			current = filepath.Join(current, part)
			dirs = append(dirs, current)
		}
	}

	scripts := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		path := filepath.Join(dir, setupScriptName)
		info, err := os.Stat(path)
		if err == nil {
			if info.IsDir() {
				return nil, fmt.Errorf("setup fixture %q is a directory", path)
			}
			scripts = append(scripts, path)
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to check setup fixture %q: %v", path, err)
		}
	}

	return scripts, nil
}
