package mire

import (
	"bytes"
	"strings"
	"testing"
)

func TestFirstMismatchingLineReturnsFirstChangedLine(t *testing.T) {
	got := firstMismatchingLine(
		[]byte("$ echo x\nx\n$ \nexit\n"),
		[]byte("$ echo a\na\n$ \nexit\n"),
	)

	if got.lineNumber != 1 {
		t.Fatalf("lineNumber = %d, want 1", got.lineNumber)
	}
	if got.expected != (mismatchLine{text: "$ echo x", present: true}) {
		t.Fatalf("expected line = %#v, want first changed expected line", got.expected)
	}
	if got.actual != (mismatchLine{text: "$ echo a", present: true}) {
		t.Fatalf("actual line = %#v, want first changed actual line", got.actual)
	}
}

func TestFirstMismatchingLineMarksMissingLine(t *testing.T) {
	got := firstMismatchingLine(
		[]byte("alpha\nbeta\n"),
		[]byte("alpha\n"),
	)

	if got.lineNumber != 2 {
		t.Fatalf("lineNumber = %d, want 2", got.lineNumber)
	}
	if got.expected != (mismatchLine{text: "beta", present: true}) {
		t.Fatalf("expected line = %#v, want missing expected line content", got.expected)
	}
	if got.actual != (mismatchLine{}) {
		t.Fatalf("actual line = %#v, want missing line", got.actual)
	}
}

func TestWriteScenarioMismatchPrintsOnlyFirstMismatch(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer
	writeScenarioMismatch(&buf, &testMismatchError{
		expected: []byte("$ echo x\nx\n$ \nexit\n"),
		actual:   []byte("$ echo a\na\n$ \nexit\n"),
	})

	got := buf.String()
	for _, want := range []string{
		"Expected\n",
		"$ echo x\n",
		"Actual\n",
		"$ echo a\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("writeScenarioMismatch() output = %q, want substring %q", got, want)
		}
	}
	if strings.Contains(got, "mire › Expected:\n") || strings.Contains(got, "mire › Actual:\n") {
		t.Fatalf("writeScenarioMismatch() output = %q, want mismatch lines without mire prefix", got)
	}

	for _, unwanted := range []string{"\nx\n", "\na\n", "\nexit\n"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("writeScenarioMismatch() output = %q, want only first mismatching line", got)
		}
	}
}

func TestWriteScenarioMismatchFormatsMissingLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer
	writeScenarioMismatch(&buf, &testMismatchError{
		expected: []byte("alpha\nbeta\n"),
		actual:   []byte("alpha\n"),
	})

	got := buf.String()
	if !strings.Contains(got, "beta\n") {
		t.Fatalf("writeScenarioMismatch() output = %q, want expected mismatching line", got)
	}
	if !strings.Contains(got, missingMismatchLine+"\n") {
		t.Fatalf("writeScenarioMismatch() output = %q, want missing-line placeholder", got)
	}
}

func TestWriteScenarioMismatchItalicizesLabels(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	var buf bytes.Buffer
	writeScenarioMismatch(&buf, &testMismatchError{
		expected: []byte("alpha\n"),
		actual:   []byte("beta\n"),
	})

	got := buf.String()
	if !strings.Contains(got, "\x1b[3m") {
		t.Fatalf("writeScenarioMismatch() output = %q, want italic styling", got)
	}
}

func TestTestMismatchErrorIncludesLineNumber(t *testing.T) {
	err := &testMismatchError{lineNumber: 7}
	if got := err.Error(); got != "output differed at line 7" {
		t.Fatalf("Error() = %q, want %q", got, "output differed at line 7")
	}
}
