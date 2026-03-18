package miro

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"miro/internal/output"
)

var ()

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

	overwrite, err := confirmRecordOverwrite(target, rio)
	if err != nil {
		return err
	}
	if !overwrite {
		return ErrRecordingDiscarded
	}

	output.Fprintln(rio.err, "Run commands in the recorder shell, then type exit to finish.")

	if err := runRecordSession(target, rawIn, rawOut, rio); err != nil {
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

func runRecordSession(dir, rawIn, rawOut string, rio recordIO) error {
	cmd := exec.Command("script", "-q", "-I", rawIn, "-O", rawOut)
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
	recordedIn, err := os.ReadFile(rawIn)
	if err != nil {
		return nil, nil, err
	}

	recordedOut, err := loadRecordedOutput(rawOut)
	if err != nil {
		return nil, nil, err
	}

	return recordedIn, recordedOut, nil
}
