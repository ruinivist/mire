package output

import (
	"fmt"
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
	fmt.Println(Format(msg))
}

func Printf(format string, args ...any) {
	fmt.Print(Format(fmt.Sprintf(format, args...)))
}
