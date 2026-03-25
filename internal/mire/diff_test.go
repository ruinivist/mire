package mire

import (
	"bytes"
	"strings"
	"testing"

	"mire/internal/testutil"
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

func TestCompareOutputMatchesEqualOutput(t *testing.T) {
	matched, details := compareOutput([]byte("alpha\n"), []byte("alpha\n"), nil)
	if !matched {
		t.Fatalf("compareOutput() matched = false, want true with details %#v", details)
	}
}

func TestCompareOutputSkipsIgnoredMismatches(t *testing.T) {
	matched, details := compareOutput(
		[]byte("ts=111\nid=foo\nstable\n"),
		[]byte("ts=222\nid=bar\nstable\n"),
		[]string{`^ts=\d+$`, `^id=\w+$`},
	)
	if !matched {
		t.Fatalf("compareOutput() matched = false, want true with details %#v", details)
	}
}

func TestCompareOutputReturnsFirstNonIgnoredMismatch(t *testing.T) {
	matched, details := compareOutput(
		[]byte("ts=111\nstable\nfinal=a\n"),
		[]byte("ts=222\nstable\nfinal=b\n"),
		[]string{`^ts=\d+$`},
	)
	if matched {
		t.Fatal("compareOutput() matched = true, want false")
	}
	if details.lineNumber != 3 {
		t.Fatalf("lineNumber = %d, want 3", details.lineNumber)
	}
	if details.expected != (mismatchLine{text: "final=a", present: true}) {
		t.Fatalf("expected line = %#v, want first non-ignored mismatch", details.expected)
	}
	if details.actual != (mismatchLine{text: "final=b", present: true}) {
		t.Fatalf("actual line = %#v, want first non-ignored mismatch", details.actual)
	}
}

func TestCompareOutputDoesNotIgnoreOneSidedMatch(t *testing.T) {
	matched, details := compareOutput(
		[]byte("ts=111\n"),
		[]byte("value=222\n"),
		[]string{`^ts=\d+$`},
	)
	if matched {
		t.Fatal("compareOutput() matched = true, want false")
	}
	if details.lineNumber != 1 {
		t.Fatalf("lineNumber = %d, want 1", details.lineNumber)
	}
}

func TestCompareOutputDoesNotIgnoreMissingLine(t *testing.T) {
	matched, details := compareOutput(
		[]byte("ts=111\n"),
		[]byte(""),
		[]string{`^ts=\d+$`},
	)
	if matched {
		t.Fatal("compareOutput() matched = true, want false")
	}
	if details.lineNumber != 1 {
		t.Fatalf("lineNumber = %d, want 1", details.lineNumber)
	}
	if details.actual.present {
		t.Fatalf("actual line = %#v, want missing line", details.actual)
	}
}

func TestWriteScenarioMismatchPrintsOnlyFirstMismatch(t *testing.T) {
	var buf bytes.Buffer
	writeScenarioMismatch(&buf, &testMismatchError{
		expected: []byte("$ echo x\nx\n$ \nexit\n"),
		actual:   []byte("$ echo a\na\n$ \nexit\n"),
		details: mismatchDetails{
			lineNumber: 1,
			expected:   mismatchLine{text: "$ echo x", present: true},
			actual:     mismatchLine{text: "$ echo a", present: true},
		},
	})

	got := testutil.StripANSI(buf.String())
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
	var buf bytes.Buffer
	writeScenarioMismatch(&buf, &testMismatchError{
		expected: []byte("alpha\nbeta\n"),
		actual:   []byte("alpha\n"),
		details: mismatchDetails{
			lineNumber: 2,
			expected:   mismatchLine{text: "beta", present: true},
			actual:     mismatchLine{},
		},
	})

	got := testutil.StripANSI(buf.String())
	if !strings.Contains(got, "beta\n") {
		t.Fatalf("writeScenarioMismatch() output = %q, want expected mismatching line", got)
	}
	if !strings.Contains(got, missingMismatchLine+"\n") {
		t.Fatalf("writeScenarioMismatch() output = %q, want missing-line placeholder", got)
	}
}

func TestWriteScenarioMismatchItalicizesLabels(t *testing.T) {
	var buf bytes.Buffer
	writeScenarioMismatch(&buf, &testMismatchError{
		expected: []byte("alpha\n"),
		actual:   []byte("beta\n"),
		details: mismatchDetails{
			lineNumber: 1,
			expected:   mismatchLine{text: "alpha", present: true},
			actual:     mismatchLine{text: "beta", present: true},
		},
	})

	got := buf.String()
	if !strings.Contains(got, "\x1b[3m") {
		t.Fatalf("writeScenarioMismatch() output = %q, want italic styling", got)
	}
}

func TestTestMismatchErrorIncludesLineNumber(t *testing.T) {
	err := &testMismatchError{details: mismatchDetails{lineNumber: 7}}
	if got := err.Error(); got != "output differed at line 7" {
		t.Fatalf("Error() = %q, want %q", got, "output differed at line 7")
	}
}
