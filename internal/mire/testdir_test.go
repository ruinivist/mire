package mire

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mire/internal/testutil"
)

func TestInitCreatesConfigAtProjectRoot(t *testing.T) {
	root := t.TempDir()

	testutil.WithWorkingDir(t, root, func() struct{} {
		if err := Init(); err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		return struct{}{}
	})

	got := testutil.ReadFile(t, filepath.Join(root, "mire.toml"))
	if got != testutil.DefaultWrittenConfig("e2e") {
		t.Fatalf("config = %q, want %q", got, testutil.DefaultWrittenConfig("e2e"))
	}
	assertRecordShell(t, filepath.Join(root, "e2e", recordShellName))
}

func TestInitUsesGitRoot(t *testing.T) {
	root := t.TempDir()
	testutil.MustGitInit(t, root)
	subdir := filepath.Join(root, "nested", "dir")
	testutil.MustMkdirAll(t, subdir)

	testutil.WithWorkingDir(t, subdir, func() struct{} {
		if err := Init(); err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		return struct{}{}
	})

	if _, err := os.Stat(filepath.Join(root, "mire.toml")); err != nil {
		t.Fatalf("Stat(%q) error = %v", filepath.Join(root, "mire.toml"), err)
	}
	assertRecordShell(t, filepath.Join(root, "e2e", recordShellName))
}

func TestInitLeavesExistingValidConfigUntouchedAndRefreshesShell(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("custom/suite"))
	testutil.WriteFile(t, filepath.Join(root, "custom", "suite", recordShellName), "outdated\n")

	testutil.WithWorkingDir(t, root, func() struct{} {
		if err := Init(); err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		return struct{}{}
	})

	got := testutil.ReadFile(t, filepath.Join(root, "mire.toml"))
	if got != testutil.ValidConfigContent("custom/suite") {
		t.Fatalf("config = %q, want %q", got, testutil.ValidConfigContent("custom/suite"))
	}
	if got := testutil.ReadFile(t, filepath.Join(root, "custom", "suite", recordShellName)); got != buildRecordShellScript() {
		t.Fatalf("shell = %q, want refreshed recorder shell", got)
	}
}

func TestInitCreatesMissingConfiguredTestDir(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("custom/suite"))

	testutil.WithWorkingDir(t, root, func() struct{} {
		if err := Init(); err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		return struct{}{}
	})

	assertRecordShell(t, filepath.Join(root, "custom", "suite", recordShellName))
}

func TestInitFailsWhenExistingConfigInvalid(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), "test_dir = \"e2e\"\n")

	err := testutil.WithWorkingDir(t, root, func() error {
		return Init()
	})
	if err == nil {
		t.Fatal("Init() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "missing [mire] config") {
		t.Fatalf("Init() error = %q, want missing [mire] config error", err.Error())
	}
}

func TestResolveTestDirFromConfig(t *testing.T) {
	root := t.TempDir()
	configured := filepath.Join(root, "custom", "suite")
	testutil.MustMkdirAll(t, configured)
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("custom/suite"))

	got := testutil.WithWorkingDir(t, root, func() string {
		path, err := ResolveTestDir()
		if err != nil {
			t.Fatalf("ResolveTestDir() error = %v", err)
		}
		return path
	})

	if got != configured {
		t.Fatalf("ResolveTestDir() = %q, want %q", got, configured)
	}
}

func TestResolveTestDirMissingConfigFails(t *testing.T) {
	root := t.TempDir()

	err := testutil.WithWorkingDir(t, root, func() error {
		_, err := ResolveTestDir()
		return err
	})

	if err == nil {
		t.Fatal("ResolveTestDir() error = nil, want error")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("ResolveTestDir() error = %v, want os.ErrNotExist", err)
	}
}

func TestResolveTestDirMissingTestDirFails(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), "[mire]\n")

	err := testutil.WithWorkingDir(t, root, func() error {
		_, err := ResolveTestDir()
		return err
	})

	if err == nil {
		t.Fatal("ResolveTestDir() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "missing required mire.test_dir") {
		t.Fatalf("ResolveTestDir() error = %q, want missing required mire.test_dir", err.Error())
	}
}

func TestResolveTestDirEmptyTestDirFails(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), "[mire]\ntest_dir = \"\"\n")

	err := testutil.WithWorkingDir(t, root, func() error {
		_, err := ResolveTestDir()
		return err
	})

	if err == nil {
		t.Fatal("ResolveTestDir() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "empty mire.test_dir") {
		t.Fatalf("ResolveTestDir() error = %q, want empty mire.test_dir", err.Error())
	}
}

func TestResolveTestDirMalformedConfigFails(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), "[mire]\ntest_dir = [\n")

	err := testutil.WithWorkingDir(t, root, func() error {
		_, err := ResolveTestDir()
		return err
	})

	if err == nil {
		t.Fatal("ResolveTestDir() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Fatalf("ResolveTestDir() error = %q, want read failure", err.Error())
	}
}

func TestResolveTestDirConfiguredMissingDirectoryReturnsConfiguredPath(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "missing")
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("missing"))

	got := testutil.WithWorkingDir(t, root, func() string {
		path, err := ResolveTestDir()
		if err != nil {
			t.Fatalf("ResolveTestDir() error = %v", err)
		}
		return path
	})

	if got != want {
		t.Fatalf("ResolveTestDir() = %q, want %q", got, want)
	}
}

func TestResolveTestDirConfiguredFileFails(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "case.txt"), "hello\n")
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("case.txt"))

	err := testutil.WithWorkingDir(t, root, func() error {
		_, err := ResolveTestDir()
		return err
	})

	if err == nil {
		t.Fatal("ResolveTestDir() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "configured test_dir is not a directory") {
		t.Fatalf("ResolveTestDir() error = %q, want file path error", err.Error())
	}
}

func TestResolveTestDirUsesGitRoot(t *testing.T) {
	root := t.TempDir()
	testutil.MustGitInit(t, root)
	want := filepath.Join(root, "e2e")
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("e2e"))
	subdir := filepath.Join(root, "nested", "dir")
	testutil.MustMkdirAll(t, subdir)

	got := testutil.WithWorkingDir(t, subdir, func() string {
		path, err := ResolveTestDir()
		if err != nil {
			t.Fatalf("ResolveTestDir() error = %v", err)
		}
		return path
	})

	if got != want {
		t.Fatalf("ResolveTestDir() = %q, want %q", got, want)
	}
}
