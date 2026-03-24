package config

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	TestDir     string
	IgnoreDiffs []string
	Sandbox     map[string]string
	Mounts      []string
	Paths       []string
}

type tomlConfig struct {
	Mire    tomlMireConfig `toml:"mire"`
	Sandbox toml.Primitive `toml:"sandbox"`
}

type tomlMireConfig struct {
	TestDir     string   `toml:"test_dir"`
	IgnoreDiffs []string `toml:"ignore_diffs"`
}

type tomlSandboxConfig struct {
	Home   string   `toml:"home"`
	Mounts []string `toml:"mounts"`
	Paths  []string `toml:"paths"`
}

//go:embed mire.toml
var defaultConfigFS embed.FS

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
	if !meta.IsDefined("mire", "ignore_diffs") {
		return Config{}, fmt.Errorf("failed to read %s: missing required mire.ignore_diffs", path)
	}
	ignoreDiffs, err := validateIgnoreDiffs(path, raw.Mire.IgnoreDiffs)
	if err != nil {
		return Config{}, err
	}
	if !meta.IsDefined("sandbox") {
		return Config{}, fmt.Errorf("failed to read %s: missing [sandbox] config", path)
	}
	if !meta.IsDefined("sandbox", "home") {
		return Config{}, fmt.Errorf("failed to read %s: missing required sandbox.home", path)
	}
	if !meta.IsDefined("sandbox", "mounts") {
		return Config{}, fmt.Errorf("failed to read %s: missing required sandbox.mounts", path)
	}
	if !meta.IsDefined("sandbox", "paths") {
		return Config{}, fmt.Errorf("failed to read %s: missing required sandbox.paths", path)
	}

	sandbox, mounts, paths, err := decodeSandbox(meta, path, raw.Sandbox)
	if err != nil {
		return Config{}, err
	}

	return Config{
		TestDir:     raw.Mire.TestDir,
		IgnoreDiffs: ignoreDiffs,
		Sandbox:     sandbox,
		Mounts:      mounts,
		Paths:       paths,
	}, nil
}

// WriteDefaultConfig writes the embedded default mire.toml.
func WriteDefaultConfig(path string) error {
	body, err := defaultConfigFS.ReadFile("mire.toml")
	if err != nil {
		return fmt.Errorf("read embedded default mire.toml: %v", err)
	}

	return os.WriteFile(path, body, 0o644)
}

func decodeSandbox(meta toml.MetaData, path string, raw toml.Primitive) (map[string]string, []string, []string, error) {
	var typed tomlSandboxConfig
	if err := meta.PrimitiveDecode(raw, &typed); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read %s: %v", path, err)
	}

	var sandboxTable map[string]any
	if err := meta.PrimitiveDecode(raw, &sandboxTable); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read %s: %v", path, err)
	}

	sandbox := map[string]string{
		"home": typed.Home,
	}
	for key, value := range sandboxTable {
		if key == "home" || key == "mounts" || key == "paths" {
			continue
		}

		str, ok := value.(string)
		if !ok {
			return nil, nil, nil, fmt.Errorf("failed to read %s: sandbox.%s must be a string", path, key)
		}
		sandbox[key] = str
	}

	return validateSandbox(path, sandbox, typed.Mounts, typed.Paths)
}

func validateSandbox(path string, sandbox map[string]string, mounts, paths []string) (map[string]string, []string, []string, error) {
	validated := cloneSandbox(sandbox)

	for key := range validated {
		if !lowerSnakeCasePattern.MatchString(key) {
			return nil, nil, nil, fmt.Errorf("failed to read %s: invalid sandbox key %q: must be lower_snake_case", path, key)
		}
	}

	for key := range requiredSandboxDefaults {
		value, ok := validated[key]
		if !ok {
			return nil, nil, nil, fmt.Errorf("failed to read %s: missing required sandbox.%s", path, key)
		}
		if value == "" {
			return nil, nil, nil, fmt.Errorf("failed to read %s: empty sandbox.%s", path, key)
		}
	}

	if !filepath.IsAbs(validated["home"]) {
		return nil, nil, nil, fmt.Errorf("failed to read %s: sandbox.home must be an absolute path", path)
	}

	validatedMounts, err := normalizeMounts(path, mounts)
	if err != nil {
		return nil, nil, nil, err
	}
	validatedPaths, err := normalizePaths(path, paths)
	if err != nil {
		return nil, nil, nil, err
	}

	return validated, validatedMounts, validatedPaths, nil
}

func validateIgnoreDiffs(path string, ignoreDiffs []string) ([]string, error) {
	for i, pattern := range ignoreDiffs {
		if _, err := regexp.Compile(pattern); err != nil {
			return nil, fmt.Errorf("failed to read %s: invalid mire.ignore_diffs[%d] regex %q: %v", path, i, pattern, err)
		}
	}

	return cloneStrings(ignoreDiffs), nil
}

func normalizeMounts(configPath string, mounts []string) ([]string, error) {
	normalized := make([]string, 0, len(mounts))
	for _, mount := range mounts {
		hostPath, sandboxPath, ok := strings.Cut(mount, ":")
		if !ok {
			normalized = append(normalized, mount)
			continue
		}
		hostPath, err := normalizePath(configPath, hostPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: sandbox mount host path %q does not exist", configPath, hostPath)
		}
		normalized = append(normalized, hostPath+":"+sandboxPath)
	}

	return normalized, nil
}

func normalizePaths(configPath string, paths []string) ([]string, error) {
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		hostPath, err := normalizePath(configPath, path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: sandbox path host path %q does not exist", configPath, hostPath)
		}
		normalized = append(normalized, hostPath)
	}

	return normalized, nil
}

func normalizePath(configPath, path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Clean(filepath.Join(filepath.Dir(configPath), path))
	}
	if _, err := os.Stat(path); err != nil {
		return path, err
	}

	return path, nil
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

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	cloned := make([]string, len(values))
	copy(cloned, values)

	return cloned
}
