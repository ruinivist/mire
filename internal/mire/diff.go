package mire

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"mire/internal/output"
)

const missingMismatchLine = "<no line>"

type mismatchLine struct {
	text    string
	present bool
}

type mismatchDetails struct {
	lineNumber int
	expected   mismatchLine
	actual     mismatchLine
}

func writeScenarioMismatch(w io.Writer, mismatch *testMismatchError) {
	details := mismatch.details

	fmt.Fprintf(w, "%s\n%s\n", italicLabel("Expected", output.Pass), formatMismatchLine(details.expected))
	fmt.Fprintf(w, "%s\n%s\n", italicLabel("Actual", output.Fail), formatMismatchLine(details.actual))
}

func firstMismatchingLine(expected, actual []byte) mismatchDetails {
	_, details := compareOutput(expected, actual, nil)
	return details
}

func compareOutput(expected, actual []byte, ignoreDiffs []string) (bool, mismatchDetails) {
	expectedLines := splitMismatchLines(expected)
	actualLines := splitMismatchLines(actual)
	patterns := compileIgnoreDiffs(ignoreDiffs)

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		expectedLine := mismatchLineAt(expectedLines, i)
		actualLine := mismatchLineAt(actualLines, i)
		if expectedLine == actualLine {
			continue
		}
		if shouldIgnoreMismatch(expectedLine, actualLine, patterns) {
			continue
		}
		return false, mismatchDetails{
			lineNumber: i + 1,
			expected:   expectedLine,
			actual:     actualLine,
		}
	}

	return true, mismatchDetails{
		expected: mismatchLineAt(expectedLines, len(expectedLines)-1),
		actual:   mismatchLineAt(actualLines, len(actualLines)-1),
	}
}

func shouldIgnoreMismatch(expected, actual mismatchLine, patterns []*regexp.Regexp) bool {
	if !expected.present || !actual.present || len(patterns) == 0 {
		return false
	}

	return lineMatchesAnyPattern(expected.text, patterns) && lineMatchesAnyPattern(actual.text, patterns)
}

func compileIgnoreDiffs(ignoreDiffs []string) []*regexp.Regexp {
	if len(ignoreDiffs) == 0 {
		return nil
	}

	patterns := make([]*regexp.Regexp, 0, len(ignoreDiffs))
	for _, ignoreDiff := range ignoreDiffs {
		patterns = append(patterns, regexp.MustCompile(ignoreDiff))
	}

	return patterns
}

func lineMatchesAnyPattern(line string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(line) {
			return true
		}
	}

	return false
}

func splitMismatchLines(data []byte) []string {
	if len(data) == 0 {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	if data[len(data)-1] == '\n' {
		lines = lines[:len(lines)-1]
	}

	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}

	return lines
}

func mismatchLineAt(lines []string, index int) mismatchLine {
	if index < 0 || index >= len(lines) {
		return mismatchLine{}
	}

	return mismatchLine{
		text:    lines[index],
		present: true,
	}
}

func formatMismatchLine(line mismatchLine) string {
	if !line.present {
		return missingMismatchLine
	}

	return line.text
}

func italicLabel(text string, color output.Color) string {
	return output.NewStyle().Italic().Apply(output.Label(text, color))
}
