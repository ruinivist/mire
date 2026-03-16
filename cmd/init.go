package cmd

import (
	"strings"

	"miro/internal/miro"
	"miro/internal/output"
)

func runInit(args []string) int {
	if len(args) != 0 {
		output.Printf("unknown init option: %s\n", strings.Join(args, " "))
		return 1
	}

	if err := miro.Init(); err != nil {
		output.Printf("%v\n", err)
		return 1
	}

	output.Println("Done initialising...")
	return 0
}
