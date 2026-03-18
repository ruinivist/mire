package miro

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"miro/internal/output"
)

const (
	recordVisibleHome = "/home/test"
	recordVisibleRepo = "/home/test/repo"
	recordVisibleTmp  = "/tmp"
	recordGitDate     = "2024-01-01T00:00:00Z"
)

type recordIO struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func recordScenario(target string, rio recordIO) error {
	rawIn, rawOut, cleanup, err := newRecordFiles()
	if err != nil {
		return err
	}
	defer cleanup()

	sandbox, cleanupSandbox, err := newRecordSandbox()
	if err != nil {
		return err
	}
	defer cleanupSandbox()

	overwrite, err := confirmRecordOverwrite(target, rio)
	if err != nil {
		return err
	}
	if !overwrite {
		return ErrRecordingDiscarded
	}

	output.Fprintln(rio.err, "Run commands in the recorder shell, then type exit to finish.")

	if err := runRecordSession(target, rawIn, rawOut, sandbox, rio); err != nil {
		return err
	}

	save, err := confirmRecordSave(rio)
	if err != nil {
		return err
	}
	if !save {
		return ErrRecordingDiscarded
	}

	recordedIn, recordedOut, err := loadRecordedFixtures(rawIn, rawOut)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(target, "in"), recordedIn, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(target, "out"), recordedOut, 0o644); err != nil {
		return err
	}

	return nil
}

func newRecordFiles() (string, string, func(), error) {
	dir, err := os.MkdirTemp("", "miro-record-")
	if err != nil {
		return "", "", nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return filepath.Join(dir, "in"), filepath.Join(dir, "out"), cleanup, nil
}

type recordSandbox struct {
	hostHome    string
	hostTmp     string
	projectRoot string
	wrapperPath string
	pathEnv     string
}

func newRecordSandbox() (recordSandbox, func(), error) {
	cwd, err := os.Getwd()
	if err != nil {
		return recordSandbox{}, nil, err
	}

	return newRecordSandboxForProjectRoot(projectRoot(cwd), os.Getenv("PATH"))
}

func newRecordSandboxForProjectRoot(root, pathEnv string) (recordSandbox, func(), error) {
	dir, err := os.MkdirTemp("", "miro-record-sandbox-")
	if err != nil {
		return recordSandbox{}, nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	sandbox := recordSandbox{
		hostHome:    filepath.Join(dir, "home"),
		hostTmp:     filepath.Join(dir, "tmp"),
		projectRoot: root,
		wrapperPath: filepath.Join(dir, "sandbox.sh"),
		pathEnv:     pathEnv,
	}

	for _, path := range []string{
		sandbox.hostHome,
		filepath.Join(sandbox.hostHome, "repo"),
		sandbox.hostTmp,
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			cleanup()
			return recordSandbox{}, nil, err
		}
	}

	body := buildRecordWrapperScript(sandbox)
	if err := os.WriteFile(sandbox.wrapperPath, []byte(body), 0o755); err != nil {
		cleanup()
		return recordSandbox{}, nil, err
	}

	return sandbox, cleanup, nil
}

func buildRecordWrapperScript(sandbox recordSandbox) string {
	var body strings.Builder

	body.WriteString("#!/bin/sh\n")
	body.WriteString("set -eu\n\n")
	body.WriteString("if command -v git >/dev/null 2>&1; then\n")
	body.WriteString(fmt.Sprintf("  HOME=%s GIT_CONFIG_NOSYSTEM=1 git config --global user.name 'Miro Test' >/dev/null 2>&1 || :\n", shQuote(sandbox.hostHome)))
	body.WriteString(fmt.Sprintf("  HOME=%s GIT_CONFIG_NOSYSTEM=1 git config --global user.email 'miro-test@example.com' >/dev/null 2>&1 || :\n", shQuote(sandbox.hostHome)))
	body.WriteString(fmt.Sprintf("  HOME=%s GIT_CONFIG_NOSYSTEM=1 git config --global init.defaultBranch main >/dev/null 2>&1 || :\n", shQuote(sandbox.hostHome)))
	body.WriteString(fmt.Sprintf("  HOME=%s GIT_CONFIG_NOSYSTEM=1 git config --global advice.defaultBranchName false >/dev/null 2>&1 || :\n", shQuote(sandbox.hostHome)))
	body.WriteString("fi\n\n")
	body.WriteString("exec bwrap \\\n")
	body.WriteString("  --ro-bind / / \\\n")
	body.WriteString("  --tmpfs /home \\\n")
	body.WriteString(fmt.Sprintf("  --bind %s %s \\\n", shQuote(sandbox.hostHome), shQuote(recordVisibleHome)))
	body.WriteString(fmt.Sprintf("  --ro-bind %s %s \\\n", shQuote(sandbox.projectRoot), shQuote(recordVisibleRepo)))
	body.WriteString(fmt.Sprintf("  --bind %s %s \\\n", shQuote(sandbox.hostTmp), shQuote(recordVisibleTmp)))
	body.WriteString("  --dev /dev \\\n")
	body.WriteString("  --proc /proc \\\n")
	body.WriteString("  --unshare-pid \\\n")
	body.WriteString("  --die-with-parent \\\n")
	body.WriteString(fmt.Sprintf("  --setenv GIT_AUTHOR_DATE %s \\\n", shQuote(recordGitDate)))
	body.WriteString(fmt.Sprintf("  --setenv GIT_COMMITTER_DATE %s \\\n", shQuote(recordGitDate)))
	body.WriteString("  --setenv GIT_CONFIG_NOSYSTEM '1' \\\n")
	body.WriteString("  --setenv GIT_PAGER 'cat' \\\n")
	body.WriteString("  --setenv HISTFILE '/dev/null' \\\n")
	body.WriteString(fmt.Sprintf("  --setenv HOME %s \\\n", shQuote(recordVisibleHome)))
	body.WriteString("  --setenv LANG 'C' \\\n")
	body.WriteString("  --setenv LC_ALL 'C' \\\n")
	body.WriteString("  --setenv PAGER 'cat' \\\n")
	body.WriteString(fmt.Sprintf("  --setenv PATH %s \\\n", shQuote(sandbox.pathEnv)))
	body.WriteString("  --setenv PS1 '$ ' \\\n")
	body.WriteString("  --setenv TERM 'xterm-256color' \\\n")
	body.WriteString(fmt.Sprintf("  --setenv TMPDIR %s \\\n", shQuote(recordVisibleTmp)))
	body.WriteString("  --setenv TZ 'UTC' \\\n")
	body.WriteString(fmt.Sprintf("  --chdir %s \\\n", shQuote(recordVisibleHome)))
	body.WriteString("  bash --noprofile --norc -i\n")

	return body.String()
}

func shQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func runRecordSession(dir, rawIn, rawOut string, sandbox recordSandbox, rio recordIO) error {
	cmd := exec.Command("script", "-q", "-E", "always", "-I", rawIn, "-O", rawOut, "-c", sandbox.wrapperPath)
	cmd.Dir = dir
	cmd.Stdin = rio.in
	cmd.Stdout = rio.out
	cmd.Stderr = rio.err
	return cmd.Run()
}

func confirmRecordOverwrite(target string, rio recordIO) (bool, error) {
	exists, err := recordFixturesExist(target)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}

	output.Fprintf(rio.err, "Overwrite existing recording? [y/N] ")
	return readRecordConfirmation(rio)
}

func confirmRecordSave(rio recordIO) (bool, error) {
	output.Fprintf(rio.err, "Save recording? [y/N] ")

	return readRecordConfirmation(rio)
}

func readRecordConfirmation(rio recordIO) (bool, error) {
	reply, err := bufio.NewReader(rio.in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	reply = strings.TrimSpace(reply)
	switch strings.ToLower(reply) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func recordFixturesExist(target string) (bool, error) {
	for _, path := range []string{
		filepath.Join(target, "in"),
		filepath.Join(target, "out"),
	} {
		_, err := os.Stat(path)
		if err == nil {
			return true, nil
		}
		if !os.IsNotExist(err) {
			return false, err
		}
	}

	return false, nil
}

func loadRecordedFixtures(rawIn, rawOut string) ([]byte, []byte, error) {
	recordedIn, err := loadRecordedInput(rawIn)
	if err != nil {
		return nil, nil, err
	}

	recordedOut, err := loadRecordedOutput(rawOut)
	if err != nil {
		return nil, nil, err
	}

	return recordedIn, recordedOut, nil
}
