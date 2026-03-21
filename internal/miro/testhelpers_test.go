package miro

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"miro/internal/testutil"
)

func assertRecordShell(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("%q mode = %o, want executable", path, info.Mode().Perm())
	}

	if got := testutil.ReadFile(t, path); got != buildRecordShellScript() {
		t.Fatalf("shell = %q, want generated recorder shell", got)
	}
}

func withRecordStreams[T any](t *testing.T, input string, fn func(recordIO) T) T {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stdin.txt")
	testutil.WriteFile(t, path, input)

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}

	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close record input: %v", err)
		}
	})

	return fn(recordIO{
		in:  file,
		out: ioDiscard{},
		err: &bytes.Buffer{},
	})
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func mustWriteRecordShell(t *testing.T, testDir string) {
	t.Helper()

	if err := writeRecordShell(testDir); err != nil {
		t.Fatalf("writeRecordShell(%q) error = %v", testDir, err)
	}
}

func defaultSandboxConfig() map[string]string {
	return map[string]string{
		"visible_home": "/home/test",
	}
}

func containsEnvEntry(env []string, want string) bool {
	for _, entry := range env {
		if entry == want {
			return true
		}
	}

	return false
}

func containsEnvKey(env []string, key string) bool {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return true
		}
	}

	return false
}
