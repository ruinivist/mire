package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDir     string
		wantHasDir  bool
		wantReadErr bool
	}{
		{
			name:       "with test dir",
			content:    "[miro]\ntest_dir = \"custom/suite\"\n",
			wantDir:    "custom/suite",
			wantHasDir: true,
		},
		{
			name:       "without test dir",
			content:    "[miro]\n",
			wantHasDir: false,
		},
		{
			name:        "invalid toml",
			content:     "[miro]\ntest_dir = [\n",
			wantReadErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "miro.toml")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			got, err := ReadConfig(path)
			if tt.wantReadErr {
				if err == nil {
					t.Fatal("ReadConfig() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadConfig() error = %v", err)
			}
			if got.TestDir != tt.wantDir {
				t.Fatalf("ReadConfig() TestDir = %q, want %q", got.TestDir, tt.wantDir)
			}
			if got.HasTestDir != tt.wantHasDir {
				t.Fatalf("ReadConfig() HasTestDir = %v, want %v", got.HasTestDir, tt.wantHasDir)
			}
		})
	}
}
