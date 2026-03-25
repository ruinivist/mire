package output

import (
	"strings"
	"testing"
)

func TestFormatIncludesANSI(t *testing.T) {
	got := Format("hello\n")
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("Format() = %q, want ANSI styling", got)
	}
	if !strings.Contains(got, "hello\n") {
		t.Fatalf("Format() = %q, want message body", got)
	}
}

func TestLabelIncludesANSI(t *testing.T) {
	for _, tc := range []struct {
		name  string
		text  string
		color Color
	}{
		{name: "info", text: "RUN", color: Info},
		{name: "pass", text: "PASS", color: Pass},
		{name: "fail", text: "FAIL", color: Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Label(tc.text, tc.color)
			if !strings.Contains(got, "\x1b[") {
				t.Fatalf("Label() = %q, want ANSI styling", got)
			}
			if !strings.Contains(got, tc.text) {
				t.Fatalf("Label() = %q, want text %q", got, tc.text)
			}
		})
	}
}
