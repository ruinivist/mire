package config

import (
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

	_, err := toml.DecodeFile(path, &raw)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		return Config{}, fmt.Errorf("failed to read %s: %v", path, err)
	}

	return Config{
		TestDir: raw.Miro.TestDir,
	}, nil
}
