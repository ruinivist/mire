package mire

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mire/internal/testutil"
)

func TestRecordCreatesRelativePath(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	testutil.MustMkdirAll(t, testDir)
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("e2e"))
	mustWriteRecordShell(t, testDir)
	testutil.AddFakeRecordDependencies(t, "script")

	got := testutil.WithWorkingDir(t, root, func() string {
		testutil.WithStdin(t, "y\n", func() {})
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

	if got := testutil.ReadFile(t, filepath.Join(want, "in")); got != "fake recorded input\n" {
		t.Fatalf("saved in = %q, want %q", got, "fake recorded input\n")
	}
	if got := testutil.ReadFile(t, filepath.Join(want, "out")); got != "fake recorded output\n" {
		t.Fatalf("saved out = %q, want %q", got, "fake recorded output\n")
	}
}

func TestRecordReturnsDiscardedErrorWhenSaveDeclined(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	testutil.MustMkdirAll(t, testDir)
	mustWriteRecordShell(t, testDir)
	testutil.AddFakeRecordDependencies(t, "script")

	err := testutil.WithWorkingDir(t, root, func() error {
		target := filepath.Join(testDir, "a", "b", "c")
		testutil.MustMkdirAll(t, target)
		return withRecordStreams(t, "n\n", func(rio recordIO) error {
			return recordScenario(target, recordShellPath(testDir), rio, defaultSandboxConfig(), nil)
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
	testutil.MustMkdirAll(t, target)
	mustWriteRecordShell(t, testDir)
	testutil.AddFakeRecordDependencies(t, "script")
	testutil.WriteFile(t, filepath.Join(target, "in"), "existing in\n")
	testutil.WriteFile(t, filepath.Join(target, "out"), "existing out\n")

	err := testutil.WithWorkingDir(t, root, func() error {
		return withRecordStreams(t, "n\n", func(rio recordIO) error {
			return recordScenario(target, recordShellPath(testDir), rio, defaultSandboxConfig(), nil)
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

func TestBuildRecordShellScriptUsesExpectedCommands(t *testing.T) {
	body := buildRecordShellScript()

	for _, want := range []string{
		"host_home=${MIRE_HOST_HOME:?}",
		"host_tmp=${MIRE_HOST_TMP:?}",
		"path_env=${MIRE_PATH_ENV:?}",
		"visible_home=${MIRE_VISIBLE_HOME:?}",
		"bootstrap_rc=\"$host_home/.mire-shell-rc\"",
		"setup_scripts_dir='/tmp/mire-setup-scripts'",
		"visible_bootstrap_rc=\"$visible_home/.mire-shell-rc\"",
		"for path in /tmp/mire-setup-scripts/*.sh; do",
		"source \"$path\"",
		`if [ -n "${MIRE_SETUP_SCRIPTS:-}" ]; then`,
		"i=1",
		`while IFS= read -r host_path || [ -n "$host_path" ]; do`,
		`visible_path=$(printf '%s/%03d.sh' "$setup_scripts_dir" "$i")`,
		`set -- "$@" --ro-bind "$host_path" "$visible_path"`,
		"${MIRE_SETUP_SCRIPTS-}",
		`if [ "${MIRE_COMPARE_MARKER:-0}" = "1" ]; then`,
		"printf '__MIRE_E2E_BEGIN__\\n'",
		"--bind \"$host_home\" \"$visible_home\"",
		"--bind \"$host_tmp\" '/tmp'",
		"--setenv HOME \"$visible_home\"",
		"--setenv PATH \"$path_env\"",
		"--setenv PS1 '$ '",
		"--setenv TERM 'xterm-256color'",
		"--setenv TZ 'UTC'",
		"--chdir \"$visible_home\"",
		"exec bwrap \"$@\" bash --noprofile --rcfile \"$visible_bootstrap_rc\" -i",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("wrapper = %q, want substring %q", body, want)
		}
	}
}

func TestRecordSessionEnvIncludesConfiguredSandboxEnv(t *testing.T) {
	env := recordSessionEnv(recordSandbox{
		hostHome: "/tmp/host-home",
		hostTmp:  "/tmp/host-tmp",
		pathEnv:  "/tmp/bin",
	}, map[string]string{
		"visible_home": "/sandbox/home",
		"key_word":     "value",
	}, []string{"/repo/e2e/setup.sh", "/repo/e2e/suite/setup.sh"})

	for _, want := range []string{
		"MIRE_HOST_HOME=/tmp/host-home",
		"MIRE_HOST_TMP=/tmp/host-tmp",
		"MIRE_PATH_ENV=/tmp/bin",
		"MIRE_KEY_WORD=value",
		"MIRE_VISIBLE_HOME=/sandbox/home",
		"MIRE_SETUP_SCRIPTS=/repo/e2e/setup.sh\n/repo/e2e/suite/setup.sh",
	} {
		if !containsEnvEntry(env, want) {
			t.Fatalf("env = %#v, want entry %q", env, want)
		}
	}
	if containsEnvKey(env, "MIRE_SETUP_SCRIPT_BINDS") {
		t.Fatalf("env = %#v, want MIRE_SETUP_SCRIPT_BINDS omitted", env)
	}
}

func TestRecordSessionEnvWithExtraIncludesAdditionalEntries(t *testing.T) {
	env := recordSessionEnvWithExtra(recordSandbox{
		hostHome: "/tmp/host-home",
		hostTmp:  "/tmp/host-tmp",
		pathEnv:  "/tmp/bin",
	}, map[string]string{
		"visible_home": "/sandbox/home",
	}, []string{"/repo/e2e/setup.sh"}, map[string]string{
		compareMarkerEnvName: compareMarkerEnabledValue,
	})

	for _, want := range []string{
		"MIRE_HOST_HOME=/tmp/host-home",
		"MIRE_HOST_TMP=/tmp/host-tmp",
		"MIRE_PATH_ENV=/tmp/bin",
		"MIRE_VISIBLE_HOME=/sandbox/home",
		"MIRE_SETUP_SCRIPTS=/repo/e2e/setup.sh",
		"MIRE_COMPARE_MARKER=1",
	} {
		if !containsEnvEntry(env, want) {
			t.Fatalf("env = %#v, want entry %q", env, want)
		}
	}
	if containsEnvKey(env, "MIRE_SETUP_SCRIPT_BINDS") {
		t.Fatalf("env = %#v, want MIRE_SETUP_SCRIPT_BINDS omitted", env)
	}
}

func TestRunRecordSessionUsesSandboxedScriptCommand(t *testing.T) {
	testDir := filepath.Join(t.TempDir(), "e2e")
	mustWriteRecordShell(t, testDir)
	testutil.AddFakeRecordDependencies(t, "script")

	argsPath := filepath.Join(t.TempDir(), "script.args")
	commandBodyPath := filepath.Join(t.TempDir(), "script.command")
	t.Setenv("FAKE_SCRIPT_ARGS_FILE", argsPath)
	t.Setenv("FAKE_SCRIPT_COMMAND_BODY_FILE", commandBodyPath)

	sandbox, cleanup, err := newRecordSandboxForPathEnv(os.Getenv("PATH"))
	if err != nil {
		t.Fatalf("newRecordSandboxForPathEnv() error = %v", err)
	}
	defer cleanup()

	shellPath := recordShellPath(testDir)
	err = withRecordStreams(t, "", func(rio recordIO) error {
		return runRecordSession(t.TempDir(), filepath.Join(t.TempDir(), "raw.in"), filepath.Join(t.TempDir(), "raw.out"), shellPath, sandbox, rio, defaultSandboxConfig(), []string{"/repo/e2e/setup.sh"})
	})
	if err != nil {
		t.Fatalf("runRecordSession() error = %v", err)
	}

	args := strings.Split(strings.TrimSpace(testutil.ReadFile(t, argsPath)), "\n")
	if len(args) != 9 {
		t.Fatalf("script args = %q, want 9 args", args)
	}
	if got := args[:4]; strings.Join(got, "\n") != strings.Join([]string{"-q", "-E", "always", "-I"}, "\n") {
		t.Fatalf("script args prefix = %q, want %q", got, []string{"-q", "-E", "always", "-I"})
	}
	if args[5] != "-O" {
		t.Fatalf("script args[5] = %q, want %q", args[5], "-O")
	}
	if args[7] != "-c" {
		t.Fatalf("script args[7] = %q, want %q", args[7], "-c")
	}
	if args[8] != shellPath {
		t.Fatalf("script args[8] = %q, want %q", args[8], shellPath)
	}

	body := testutil.ReadFile(t, commandBodyPath)
	for _, want := range []string{
		"host_home=${MIRE_HOST_HOME:?}",
		"visible_home=${MIRE_VISIBLE_HOME:?}",
		"bootstrap_rc=\"$host_home/.mire-shell-rc\"",
		"visible_bootstrap_rc=\"$visible_home/.mire-shell-rc\"",
		"for path in /tmp/mire-setup-scripts/*.sh; do",
		"source \"$path\"",
		`if [ -n "${MIRE_SETUP_SCRIPTS:-}" ]; then`,
		`set -- "$@" --ro-bind "$host_path" "$visible_path"`,
		`if [ "${MIRE_COMPARE_MARKER:-0}" = "1" ]; then`,
		"printf '__MIRE_E2E_BEGIN__\\n'",
		"--ro-bind / /",
		"--tmpfs /home",
		"--setenv HOME \"$visible_home\"",
		"--setenv TMPDIR '/tmp'",
		"exec bwrap \"$@\" bash --noprofile --rcfile \"$visible_bootstrap_rc\" -i",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("wrapper = %q, want substring %q", body, want)
		}
	}
}

func TestRecordScenarioUsesDeterministicSandbox(t *testing.T) {
	testutil.RequireCommands(t, "script", "bwrap", "bash")

	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	target := filepath.Join(testDir, "suite", "spec")
	testutil.MustMkdirAll(t, target)
	mustWriteRecordShell(t, testDir)
	sandboxConfig := map[string]string{
		"visible_home": "/home/test",
		"key_word":     "value",
	}
	visibleHome := sandboxConfig["visible_home"]

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close pipe reader: %v", err)
		}
	})

	writeDone := make(chan error, 1)
	go func() {
		defer close(writeDone)
		defer writer.Close()

		if _, err := writer.Write([]byte("pwd\necho \"$HOME\"\necho \"$MIRE_KEY_WORD\"\nif [ -e \"$HOME/repo\" ]; then echo FOUND; else echo MISSING; fi\npwd\nexit\n")); err != nil {
			writeDone <- err
			return
		}

		time.Sleep(300 * time.Millisecond)

		if _, err := writer.Write([]byte("y\n")); err != nil {
			writeDone <- err
			return
		}

		writeDone <- nil
	}()

	err = testutil.WithWorkingDir(t, root, func() error {
		return recordScenario(target, recordShellPath(testDir), recordIO{
			in:  reader,
			out: ioDiscard{},
			err: &bytes.Buffer{},
		}, sandboxConfig, nil)
	})
	if err != nil {
		t.Fatalf("recordScenario() error = %v", err)
	}
	if err := <-writeDone; err != nil {
		t.Fatalf("write session input: %v", err)
	}

	recordedIn := testutil.ReadFile(t, filepath.Join(target, "in"))
	if strings.Contains(recordedIn, "Script started on ") {
		t.Fatalf("saved in = %q, want stripped script wrapper", recordedIn)
	}
	for _, want := range []string{"pwd\n", "echo \"$HOME\"\n", "echo \"$MIRE_KEY_WORD\"\n", "if [ -e \"$HOME/repo\" ]; then echo FOUND; else echo MISSING; fi\n", "exit\n"} {
		if !strings.Contains(recordedIn, want) {
			t.Fatalf("saved in = %q, want substring %q", recordedIn, want)
		}
	}

	recordedOut := testutil.ReadFile(t, filepath.Join(target, "out"))
	if strings.Contains(recordedOut, "Script started on ") {
		t.Fatalf("saved out = %q, want stripped script wrapper", recordedOut)
	}
	for _, want := range []string{visibleHome, "value"} {
		if !strings.Contains(recordedOut, want) {
			t.Fatalf("saved out = %q, want substring %q", recordedOut, want)
		}
	}
	if !strings.Contains(recordedOut, "MISSING") {
		t.Fatalf("saved out = %q, want missing repo confirmation", recordedOut)
	}
	if strings.Contains(recordedOut, "\r\nFOUND\r\n") {
		t.Fatalf("saved out = %q, want repo to stay unavailable", recordedOut)
	}
}

func TestRecordFailsWhenRecorderShellMissing(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "e2e")
	testutil.WriteFile(t, filepath.Join(root, "mire.toml"), testutil.ValidConfigContent("e2e"))
	testutil.MustMkdirAll(t, testDir)
	testutil.AddFakeRecordDependencies(t, "script")

	target := filepath.Join(testDir, "suite", "spec")
	err := testutil.WithWorkingDir(t, root, func() error {
		_, err := Record(filepath.Join("suite", "spec"))
		return err
	})
	if err == nil {
		t.Fatal("Record() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "rerun `mire init`") {
		t.Fatalf("Record() error = %q, want rerun init hint", err.Error())
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("Stat(%q) error = %v, want not exists", target, statErr)
	}
}
