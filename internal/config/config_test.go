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
		wantMounts  []string
		wantErr     string
		wantMissing bool
	}{
		{
			name:    "with test dir and sandbox",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nhome = \"/home/test\"\nmounts = [\"/host/data:/sandbox/data\", \"/host/cache:/sandbox/cache\"]\nkey_word = \"value\"\n",
			wantDir: "custom/suite",
			wantSandbox: map[string]string{
				"home":     "/home/test",
				"key_word": "value",
			},
			wantMounts: []string{"/host/data:/sandbox/data", "/host/cache:/sandbox/cache"},
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
			name:    "without required home",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nmounts = []\n",
			wantErr: "missing required sandbox.home",
		},
		{
			name:    "without required mounts",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nhome = \"/home/test\"\n",
			wantErr: "missing required sandbox.mounts",
		},
		{
			name:    "empty required home",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nhome = \"\"\nmounts = []\n",
			wantErr: "empty sandbox.home",
		},
		{
			name:    "relative home",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nhome = \"home/test\"\nmounts = []\n",
			wantErr: "sandbox.home must be an absolute path",
		},
		{
			name:    "invalid sandbox key",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\nKeyWord = \"value\"\n",
			wantErr: "invalid sandbox key",
		},
		{
			name:    "non string sandbox value",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nhome = \"/home/test\"\nmounts = []\nkey_word = 1\n",
			wantErr: "sandbox.key_word must be a string",
		},
		{
			name:    "mounts wrong type",
			content: "[mire]\ntest_dir = \"custom/suite\"\n\n[sandbox]\nhome = \"/home/test\"\nmounts = \"oops\"\n",
			wantErr: "failed to read",
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
			if len(got.Mounts) != len(tt.wantMounts) {
				t.Fatalf("ReadConfig() Mounts = %#v, want %#v", got.Mounts, tt.wantMounts)
			}
			for i, want := range tt.wantMounts {
				if got.Mounts[i] != want {
					t.Fatalf("ReadConfig() Mounts[%d] = %q, want %q", i, got.Mounts[i], want)
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
	if got.Sandbox["home"] != DefaultVisibleHome {
		t.Fatalf("ReadConfig() Sandbox[home] = %q, want %q", got.Sandbox["home"], DefaultVisibleHome)
	}
	if len(got.Mounts) != 0 {
		t.Fatalf("ReadConfig() Mounts = %#v, want empty", got.Mounts)
	}
}
