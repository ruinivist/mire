package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDir     string
		wantSandbox map[string]string
		wantErr     string
		wantMissing bool
	}{
		{
			name:    "with test dir and sandbox",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nvisible_home = \"/home/test\"\nkey_word = \"value\"\n",
			wantDir: "custom/suite",
			wantSandbox: map[string]string{
				"visible_home": "/home/test",
				"key_word":     "value",
			},
		},
		{
			name:    "legacy top level key",
			content: "test_dir = \"custom/suite\"\n",
			wantErr: "missing [mire] config",
		},
		{
			name:    "without mire table",
			content: "",
			wantErr: "missing [mire] config",
		},
		{
			name:    "without test dir",
			content: "[mire]\n",
			wantErr: "missing required mire.test_dir",
		},
		{
			name:    "empty test dir",
			content: "[mire]\ntest_dir = \"\"\n",
			wantErr: "empty mire.test_dir",
		},
		{
			name:    "without sandbox table",
			content: "[mire]\ntest_dir = \"custom/suite\"\n",
			wantErr: "missing [sandbox] config",
		},
		{
			name:    "without required visible home",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\n",
			wantErr: "missing required sandbox.visible_home",
		},
		{
			name:    "empty required visible home",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nvisible_home = \"\"\n",
			wantErr: "empty sandbox.visible_home",
		},
		{
			name:    "relative visible home",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nvisible_home = \"home/test\"\n",
			wantErr: "sandbox.visible_home must be an absolute path",
		},
		{
			name:    "invalid sandbox key",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nvisible_home = \"/home/test\"\nKeyWord = \"value\"\n",
			wantErr: "invalid sandbox key",
		},
		{
			name:    "invalid toml",
			content: "[mire]\ntest_dir = [\n",
			wantErr: "failed to read",
		},
		{
			name:        "missing file",
			wantMissing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "mire.toml")
			if !tt.wantMissing {
				if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			}

			got, err := ReadConfig(path)
			if tt.wantMissing {
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("ReadConfig() error = %v, want os.ErrNotExist", err)
				}
				return
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("ReadConfig() error = nil, want error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ReadConfig() error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadConfig() error = %v", err)
			}
			if got.TestDir != tt.wantDir {
				t.Fatalf("ReadConfig() TestDir = %q, want %q", got.TestDir, tt.wantDir)
			}
			if len(got.Sandbox) != len(tt.wantSandbox) {
				t.Fatalf("ReadConfig() Sandbox = %#v, want %#v", got.Sandbox, tt.wantSandbox)
			}
			for key, want := range tt.wantSandbox {
				if got.Sandbox[key] != want {
					t.Fatalf("ReadConfig() Sandbox[%q] = %q, want %q", key, got.Sandbox[key], want)
				}
			}
		})
	}
}

func TestWriteConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mire.toml")

	if err := WriteConfig(path, Config{
		TestDir: "e2e",
		Sandbox: map[string]string{
			"visible_home": "/home/test",
			"alpha_key":    "a",
			"zulu_key":     "z",
		},
	}); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	want := "[mire]\n  test_dir = \"e2e\"\n\n[sandbox]\n  alpha_key = \"a\"\n  visible_home = \"/home/test\"\n  zulu_key = \"z\"\n"
	if string(got) != want {
		t.Fatalf("config = %q, want %q", string(got), want)
	}
}

func TestWriteConfigEmptyTestDirFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mire.toml")

	err := WriteConfig(path, Config{})
	if err == nil {
		t.Fatal("WriteConfig() error = nil, want error")
	}
	if err.Error() != "empty mire.test_dir" {
		t.Fatalf("WriteConfig() error = %q, want %q", err.Error(), "empty mire.test_dir")
	}
}

func TestWriteConfigMissingRequiredSandboxFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mire.toml")

	err := WriteConfig(path, Config{
		TestDir: "e2e",
		Sandbox: map[string]string{},
	})
	if err == nil {
		t.Fatal("WriteConfig() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "missing required sandbox.visible_home") {
		t.Fatalf("WriteConfig() error = %q, want visible_home error", err.Error())
	}
}
