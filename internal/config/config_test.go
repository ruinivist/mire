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
		wantErr     string
		wantMissing bool
	}{
		{
			name:    "with test dir",
			content: "[miro]\ntest_dir = \"custom/suite\"\n",
			wantDir: "custom/suite",
		},
		{
			name:    "legacy top level key",
			content: "test_dir = \"custom/suite\"\n",
			wantErr: "missing [miro] config",
		},
		{
			name:    "without miro table",
			content: "",
			wantErr: "missing [miro] config",
		},
		{
			name:    "without test dir",
			content: "[miro]\n",
			wantErr: "missing required miro.test_dir",
		},
		{
			name:    "empty test dir",
			content: "[miro]\ntest_dir = \"\"\n",
			wantErr: "empty miro.test_dir",
		},
		{
			name:    "invalid toml",
			content: "[miro]\ntest_dir = [\n",
			wantErr: "failed to read",
		},
		{
			name:        "missing file",
			wantMissing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "miro.toml")
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
		})
	}
}

func TestWriteConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "miro.toml")

	if err := WriteConfig(path, Config{TestDir: "e2e"}); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "[miro]\n  test_dir = \"e2e\"\n" {
		t.Fatalf("config = %q, want %q", string(got), "[miro]\n  test_dir = \"e2e\"\n")
	}
}

func TestWriteConfigEmptyTestDirFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "miro.toml")

	err := WriteConfig(path, Config{})
	if err == nil {
		t.Fatal("WriteConfig() error = nil, want error")
	}
	if err.Error() != "empty miro.test_dir" {
		t.Fatalf("WriteConfig() error = %q, want %q", err.Error(), "empty miro.test_dir")
	}
}
