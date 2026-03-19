package miro

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

const recordShellName = "shell.sh"

//go:embed record_shell.sh.tmpl
var recordShellTemplateFS embed.FS

var recordShellTemplate = template.Must(
	template.New("record_shell.sh.tmpl").ParseFS(recordShellTemplateFS, "record_shell.sh.tmpl"),
)

func recordShellPath(testDir string) string {
	return filepath.Join(testDir, recordShellName)
}

func ensureRecordShell(testDir string) error {
	path := recordShellPath(testDir)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check recorder shell %q: %v", path, err)
	}

	return writeRecordShell(testDir)
}

func writeRecordShell(testDir string) error {
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		return fmt.Errorf("failed to create test directory %q: %v", testDir, err)
	}

	path := recordShellPath(testDir)
	if err := os.WriteFile(path, []byte(buildRecordShellScript()), 0o644); err != nil {
		return fmt.Errorf("failed to write recorder shell %q: %v", path, err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return fmt.Errorf("failed to make recorder shell executable %q: %v", path, err)
	}

	return nil
}

func resolveRecordShell(testDir string) (string, error) {
	path := recordShellPath(testDir)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("missing recorder shell %q; rerun `miro init` or restore the file", path)
		}
		return "", fmt.Errorf("failed to check recorder shell %q: %v", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("recorder shell %q is a directory; rerun `miro init` or restore the file", path)
	}

	return path, nil
}

func buildRecordShellScript() string {
	var body bytes.Buffer
	if err := recordShellTemplate.Execute(&body, struct {
		GitDate string
	}{
		GitDate: shQuote(recordGitDate),
	}); err != nil {
		panic(fmt.Sprintf("render record shell template: %v", err))
	}

	return body.String()
}

func recordSessionEnv(sandbox recordSandbox, sandboxConfig map[string]string) []string {
	env := append([]string{}, os.Environ()...)
	env = append(env,
		"MIRO_HOST_HOME="+sandbox.hostHome,
		"MIRO_HOST_TMP="+sandbox.hostTmp,
		"MIRO_PATH_ENV="+sandbox.pathEnv,
	)
	for _, key := range sortedSandboxKeys(sandboxConfig) {
		env = append(env, sandboxEnvName(key)+"="+sandboxConfig[key])
	}

	return env
}

func sandboxEnvName(key string) string {
	return "MIRO_" + strings.ToUpper(key)
}

func sortedSandboxKeys(sandboxConfig map[string]string) []string {
	keys := make([]string, 0, len(sandboxConfig))
	for key := range sandboxConfig {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func shQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
