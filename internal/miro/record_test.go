package miro

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordCreatesRelativePath(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	mustMkdirAll(t, testDir)

	got := withWorkingDir(t, root, func() string {
		path, err := Record(filepath.Join("a", "b", "c") + string(os.PathSeparator))
		if err != nil {
			t.Fatalf("Record() error = %v", err)
		}
		return path
	})

	want := filepath.Join(testDir, "a", "b", "c")
	if got != want {
		t.Fatalf("Record() = %q, want %q", got, want)
	}

	info, err := os.Stat(want)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("created path %q is not a directory", want)
	}
}

func TestRecordStripsExplicitTestDirPrefix(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	mustMkdirAll(t, testDir)

	got := withWorkingDir(t, root, func() string {
		path, err := Record(filepath.Join("e2e", "a", "b", "c"))
		if err != nil {
			t.Fatalf("Record() error = %v", err)
		}
		return path
	})

	want := filepath.Join(testDir, "a", "b", "c")
	if got != want {
		t.Fatalf("Record() = %q, want %q", got, want)
	}

	if _, err := os.Stat(want); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
}

func TestRecordRejectsAbsolutePathOutsideTestDir(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "e2e"))
	outside := filepath.Join(root, "outside", "a", "b", "c")

	err := withWorkingDir(t, root, func() error {
		_, err := Record(outside)
		return err
	})

	if err == nil {
		t.Fatal("Record() error = nil, want error")
	}
	if _, statErr := os.Stat(outside); !os.IsNotExist(statErr) {
		t.Fatalf("Stat(%q) error = %v, want not exists", outside, statErr)
	}
}
