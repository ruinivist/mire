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
		name            string
		content         string
		wantDir         string
		wantIgnoreDiffs []string
		wantSandbox     map[string]string
		wantMounts      []string
		wantPaths       []string
		wantErr         string
		wantMissing     bool
		setup           func(t *testing.T, dir string) (string, []string, []string, []string)
	}{
		{
			name: "with test dir and sandbox",
			setup: func(t *testing.T, dir string) (string, []string, []string, []string) {
				hostData := filepath.Join(dir, "host-data")
				hostCache := filepath.Join(dir, "host-cache")
				for _, path := range []string{hostData, hostCache} {
					if err := os.MkdirAll(path, 0o755); err != nil {
						t.Fatalf("MkdirAll(%q) error = %v", path, err)
					}
				}
				return "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = [\"^ts=.*$\", \"^id=.*$\"]\n\n[sandbox]\nhome = \"/home/test\"\nmounts = [\"" + hostData + ":/sandbox/data\", \"" + hostCache + ":/sandbox/cache\"]\npaths = []\nkey_word = \"value\"\n",
					[]string{hostData + ":/sandbox/data", hostCache + ":/sandbox/cache"},
					nil,
					[]string{"^ts=.*$", "^id=.*$"}
			},
			wantDir: "custom/suite",
			wantIgnoreDiffs: []string{
				"^ts=.*$",
				"^id=.*$",
			},
			wantSandbox: map[string]string{
				"home":     "/home/test",
				"key_word": "value",
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
			name:    "without ignore diffs",
			content: "[mire]\ntest_dir = \"custom/suite\"\n",
			wantErr: "missing required mire.ignore_diffs",
		},
		{
			name:    "empty test dir",
			content: "[mire]\ntest_dir = \"\"\n",
			wantErr: "empty mire.test_dir",
		},
		{
			name:    "without sandbox table",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n",
			wantErr: "missing [sandbox] config",
		},
		{
			name:    "without required home",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nmounts = []\npaths = []\n",
			wantErr: "missing required sandbox.home",
		},
		{
			name:    "without required mounts",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"/home/test\"\npaths = []\n",
			wantErr: "missing required sandbox.mounts",
		},
		{
			name:    "empty required home",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"\"\nmounts = []\npaths = []\n",
			wantErr: "empty sandbox.home",
		},
		{
			name:    "relative home",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"home/test\"\nmounts = []\npaths = []\n",
			wantErr: "sandbox.home must be an absolute path",
		},
		{
			name:    "invalid sandbox key",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\npaths = []\nKeyWord = \"value\"\n",
			wantErr: "invalid sandbox key",
		},
		{
			name:    "non string sandbox value",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\npaths = []\nkey_word = 1\n",
			wantErr: "sandbox.key_word must be a string",
		},
		{
			name:    "mounts wrong type",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"/home/test\"\nmounts = \"oops\"\npaths = []\n",
			wantErr: "failed to read",
		},
		{
			name:    "ignore diffs wrong type",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = \"oops\"\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\npaths = []\n",
			wantErr: "failed to read",
		},
		{
			name:    "ignore diff entry not string",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = [1]\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\npaths = []\n",
			wantErr: "failed to read",
		},
		{
			name:    "ignore diff regex invalid",
			content: "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = [\"(\"]\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\npaths = []\n",
			wantErr: "invalid mire.ignore_diffs[0] regex",
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
		{
			name: "normalizes relative mount host path",
			setup: func(t *testing.T, dir string) (string, []string, []string, []string) {
				hostBuild := filepath.Join(dir, "build")
				if err := os.MkdirAll(hostBuild, 0o755); err != nil {
					t.Fatalf("MkdirAll(%q) error = %v", hostBuild, err)
				}
				return "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"/home/test\"\nmounts = [\"./build:/sandbox/build\"]\npaths = []\n",
					[]string{hostBuild + ":/sandbox/build"},
					nil,
					[]string{}
			},
			wantDir: "custom/suite",
			wantSandbox: map[string]string{
				"home": "/home/test",
			},
		},
		{
			name: "normalizes relative paths host path",
			setup: func(t *testing.T, dir string) (string, []string, []string, []string) {
				hostTool := filepath.Join(dir, "build", "mend")
				if err := os.MkdirAll(filepath.Dir(hostTool), 0o755); err != nil {
					t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(hostTool), err)
				}
				if err := os.WriteFile(hostTool, []byte("tool"), 0o644); err != nil {
					t.Fatalf("WriteFile(%q) error = %v", hostTool, err)
				}
				return "[mire]\ntest_dir = \"custom/suite\"\nignore_diffs = []\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\npaths = [\"./build/mend\"]\n",
					nil,
					[]string{hostTool},
					[]string{}
			},
			wantDir: "custom/suite",
			wantSandbox: map[string]string{
				"home": "/home/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "mire.toml")
			wantMounts := tt.wantMounts
			wantPaths := tt.wantPaths
			wantIgnoreDiffs := tt.wantIgnoreDiffs
			content := tt.content
			if tt.setup != nil {
				content, wantMounts, wantPaths, wantIgnoreDiffs = tt.setup(t, dir)
			}
			if !tt.wantMissing {
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
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
			if len(got.IgnoreDiffs) != len(wantIgnoreDiffs) {
				t.Fatalf("ReadConfig() IgnoreDiffs = %#v, want %#v", got.IgnoreDiffs, wantIgnoreDiffs)
			}
			for i, want := range wantIgnoreDiffs {
				if got.IgnoreDiffs[i] != want {
					t.Fatalf("ReadConfig() IgnoreDiffs[%d] = %q, want %q", i, got.IgnoreDiffs[i], want)
				}
			}
			if len(got.Sandbox) != len(tt.wantSandbox) {
				t.Fatalf("ReadConfig() Sandbox = %#v, want %#v", got.Sandbox, tt.wantSandbox)
			}
			for key, want := range tt.wantSandbox {
				if got.Sandbox[key] != want {
					t.Fatalf("ReadConfig() Sandbox[%q] = %q, want %q", key, got.Sandbox[key], want)
				}
			}
			if len(got.Mounts) != len(wantMounts) {
				t.Fatalf("ReadConfig() Mounts = %#v, want %#v", got.Mounts, wantMounts)
			}
			for i, want := range wantMounts {
				if got.Mounts[i] != want {
					t.Fatalf("ReadConfig() Mounts[%d] = %q, want %q", i, got.Mounts[i], want)
				}
			}
			if len(got.Paths) != len(wantPaths) {
				t.Fatalf("ReadConfig() Paths = %#v, want %#v", got.Paths, wantPaths)
			}
			for i, want := range wantPaths {
				if got.Paths[i] != want {
					t.Fatalf("ReadConfig() Paths[%d] = %q, want %q", i, got.Paths[i], want)
				}
			}
		})
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mire.toml")

	if err := WriteDefaultConfig(path); err != nil {
		t.Fatalf("WriteDefaultConfig() error = %v", err)
	}

	got, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("ReadConfig() error = %v", err)
	}
	if got.TestDir != "e2e" {
		t.Fatalf("ReadConfig() TestDir = %q, want %q", got.TestDir, "e2e")
	}
	if len(got.IgnoreDiffs) != 0 {
		t.Fatalf("ReadConfig() IgnoreDiffs = %#v, want empty", got.IgnoreDiffs)
	}
	if got.Sandbox["home"] != DefaultVisibleHome {
		t.Fatalf("ReadConfig() Sandbox[home] = %q, want %q", got.Sandbox["home"], DefaultVisibleHome)
	}
	if len(got.Mounts) != 0 {
		t.Fatalf("ReadConfig() Mounts = %#v, want empty", got.Mounts)
	}
	if len(got.Paths) != 0 {
		t.Fatalf("ReadConfig() Paths = %#v, want empty", got.Paths)
	}
}
