package miro

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordCreatesRelativePath(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	mustMkdirAll(t, testDir)
	addFakeRecordDependencies(t, "script")

	got := withWorkingDir(t, root, func() string {
		withStdin(t, "y\n", func() {})
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

	for _, name := range []string{"in", "out"} {
		if _, err := os.Stat(filepath.Join(want, name)); err != nil {
			t.Fatalf("Stat(%q) error = %v", filepath.Join(want, name), err)
		}
	}
}

func TestRecordStripsExplicitTestDirPrefix(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	mustMkdirAll(t, testDir)
	addFakeRecordDependencies(t, "script")

	got := withWorkingDir(t, root, func() string {
		withStdin(t, "y\n", func() {})
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

	for _, name := range []string{"in", "out"} {
		if _, err := os.Stat(filepath.Join(want, name)); err != nil {
			t.Fatalf("Stat(%q) error = %v", filepath.Join(want, name), err)
		}
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

func TestRecordReturnsDiscardedErrorWhenSaveDeclined(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	mustMkdirAll(t, testDir)
	addFakeRecordDependencies(t, "script")

	err := withWorkingDir(t, root, func() error {
		target := filepath.Join(testDir, "a", "b", "c")
		mustMkdirAll(t, target)
		return withRecordStreams(t, "n\n", func(rio recordIO) error {
			return recordScenario(target, rio)
		})
	})

	if !errors.Is(err, ErrRecordingDiscarded) {
		t.Fatalf("Record() error = %v, want ErrRecordingDiscarded", err)
	}

	target := filepath.Join(testDir, "a", "b", "c")
	for _, name := range []string{"in", "out"} {
		if _, err := os.Stat(filepath.Join(target, name)); !os.IsNotExist(err) {
			t.Fatalf("Stat(%q) error = %v, want not exists", filepath.Join(target, name), err)
		}
	}
}

func TestRecordReturnsDiscardedErrorWhenOverwriteDeclined(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	target := filepath.Join(testDir, "a", "b", "c")
	mustMkdirAll(t, target)
	addFakeRecordDependencies(t, "script")
	writeFile(t, filepath.Join(target, "in"), "existing in\n")
	writeFile(t, filepath.Join(target, "out"), "existing out\n")

	err := withWorkingDir(t, root, func() error {
		return withRecordStreams(t, "n\n", func(rio recordIO) error {
			return recordScenario(target, rio)
		})
	})

	if !errors.Is(err, ErrRecordingDiscarded) {
		t.Fatalf("recordScenario() error = %v, want ErrRecordingDiscarded", err)
	}

	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "in", want: "existing in\n"},
		{name: "out", want: "existing out\n"},
	} {
		got, readErr := os.ReadFile(filepath.Join(target, tc.name))
		if readErr != nil {
			t.Fatalf("ReadFile(%q) error = %v", filepath.Join(target, tc.name), readErr)
		}
		if string(got) != tc.want {
			t.Fatalf("%s = %q, want %q", tc.name, string(got), tc.want)
		}
	}
}

func withRecordStreams[T any](t *testing.T, input string, fn func(recordIO) T) T {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

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

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}

	oldStdin := os.Stdin
	os.Stdin = file
	t.Cleanup(func() {
		os.Stdin = oldStdin
		if err := file.Close(); err != nil {
			t.Fatalf("close stdin file: %v", err)
		}
	})

	fn()
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func addFakeRecordDependencies(t *testing.T, names ...string) {
	t.Helper()

	binDir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(binDir, name)
		body := "#!/bin/sh\nexit 0\n"
		if name == "script" {
			body = "#!/bin/sh\nin=''\nout=''\nwhile [ \"$#\" -gt 0 ]; do\n  case \"$1\" in\n    -I)\n      in=\"$2\"\n      shift 2\n      ;;\n    -O)\n      out=\"$2\"\n      shift 2\n      ;;\n    *)\n      shift\n      ;;\n  esac\ndone\nprintf 'fake recorded input\\n' > \"$in\"\nprintf 'fake recorded output\\n' > \"$out\"\nexit 0\n"
		}
		if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
