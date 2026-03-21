package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/BurntSushi/toml"
)

const DefaultVisibleHome = "/home/test"

var (
	lowerSnakeCasePattern   = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
	requiredSandboxDefaults = map[string]string{
		"home": DefaultVisibleHome,
	}
)

type Config struct {
	TestDir string
	Sandbox map[string]string
}

type tomlConfig struct {
	Mire    tomlMireConfig    `toml:"mire"`
	Sandbox map[string]string `toml:"sandbox"`
}

type tomlMireConfig struct {
	TestDir string `toml:"test_dir"`
}

func DefaultSandboxConfig() map[string]string {
	return cloneSandbox(requiredSandboxDefaults)
}

// ReadConfig reads mire.toml.
func ReadConfig(path string) (Config, error) {
	var raw tomlConfig

	meta, err := toml.DecodeFile(path, &raw)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		return Config{}, fmt.Errorf("failed to read %s: %v", path, err)
	}
	if !meta.IsDefined("mire") {
		return Config{}, fmt.Errorf("failed to read %s: missing [mire] config", path)
	}
	if !meta.IsDefined("mire", "test_dir") {
		return Config{}, fmt.Errorf("failed to read %s: missing required mire.test_dir", path)
	}
	if raw.Mire.TestDir == "" {
		return Config{}, fmt.Errorf("failed to read %s: empty mire.test_dir", path)
	}
	if !meta.IsDefined("sandbox") {
		return Config{}, fmt.Errorf("failed to read %s: missing [sandbox] config", path)
	}

	sandbox, err := validateSandbox(path, raw.Sandbox)
	if err != nil {
		return Config{}, err
	}

	return Config{
		TestDir: raw.Mire.TestDir,
		Sandbox: sandbox,
	}, nil
}

// WriteConfig writes mire.toml.
func WriteConfig(path string, cfg Config) error {
	if cfg.TestDir == "" {
		return errors.New("empty mire.test_dir")
	}
	sandbox, err := validateSandbox(path, cfg.Sandbox)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[mire]\n  test_dir = %q\n\n[sandbox]\n", cfg.TestDir)

	keys := make([]string, 0, len(sandbox))
	for key := range sandbox {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(&buf, "  %s = %q\n", key, sandbox[key])
	}

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func validateSandbox(path string, sandbox map[string]string) (map[string]string, error) {
	validated := cloneSandbox(sandbox)

	for key := range validated {
		if !lowerSnakeCasePattern.MatchString(key) {
			return nil, fmt.Errorf("failed to read %s: invalid sandbox key %q: must be lower_snake_case", path, key)
		}
	}

	for key := range requiredSandboxDefaults {
		value, ok := validated[key]
		if !ok {
			return nil, fmt.Errorf("failed to read %s: missing required sandbox.%s", path, key)
		}
		if value == "" {
			return nil, fmt.Errorf("failed to read %s: empty sandbox.%s", path, key)
		}
	}

	if !filepath.IsAbs(validated["home"]) {
		return nil, fmt.Errorf("failed to read %s: sandbox.home must be an absolute path", path)
	}

	return validated, nil
}

func cloneSandbox(sandbox map[string]string) map[string]string {
	if len(sandbox) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(sandbox))
	for key, value := range sandbox {
		cloned[key] = value
	}

	return cloned
}
