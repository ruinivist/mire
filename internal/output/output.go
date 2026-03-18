package output

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	bold     = "\x1b[1m"
	italic   = "\x1b[3m"
	wordmark = "\x1b[38;2;112;224;0m"
	chevron  = "\x1b[38;2;29;211;176m"
	reset    = "\x1b[0m"
	prefix   = wordmark + bold + italic + "miro" + reset + " " + chevron + bold + italic + "›" + reset + " "
)

func Prefix() string {
	return prefix
}

func Format(msg string) string {
	body := strings.TrimRight(msg, "\n")
	suffix := msg[len(body):]
	return Prefix() + body + suffix
}

func Println(msg string) {
	Fprintln(os.Stdout, msg)
}

func Printf(format string, args ...any) {
	Fprintf(os.Stdout, format, args...)
}

func Fprintln(w io.Writer, msg string) {
	fmt.Fprintln(w, Format(msg))
}

func Fprintf(w io.Writer, format string, args ...any) {
	fmt.Fprint(w, Format(fmt.Sprintf(format, args...)))
}
