package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	TestDir string
}

type tomlConfig struct {
	Miro tomlMiroConfig `toml:"miro"`
}

type tomlMiroConfig struct {
	TestDir string `toml:"test_dir"`
}

// ReadConfig reads miro.toml.
func ReadConfig(path string) (Config, error) {
	var raw tomlConfig

	meta, err := toml.DecodeFile(path, &raw)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		return Config{}, fmt.Errorf("failed to read %s: %v", path, err)
	}
	if !meta.IsDefined("miro") {
		return Config{}, fmt.Errorf("failed to read %s: missing [miro] config", path)
	}
	if !meta.IsDefined("miro", "test_dir") {
		return Config{}, fmt.Errorf("failed to read %s: missing required miro.test_dir", path)
	}
	if raw.Miro.TestDir == "" {
		return Config{}, fmt.Errorf("failed to read %s: empty miro.test_dir", path)
	}

	return Config{
		TestDir: raw.Miro.TestDir,
	}, nil
}

// WriteConfig writes miro.toml.
func WriteConfig(path string, cfg Config) error {
	if cfg.TestDir == "" {
		return errors.New("empty miro.test_dir")
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(tomlConfig{
		Miro: tomlMiroConfig{
			TestDir: cfg.TestDir,
		},
	}); err != nil {
		return fmt.Errorf("failed to encode %s: %v", path, err)
	}

	return os.WriteFile(path, buf.Bytes(), 0o644)
}
