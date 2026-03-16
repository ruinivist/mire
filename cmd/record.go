package cmd

import (
	"path/filepath"

	"miro/internal/miro"
	"miro/internal/output"
)

func runRecord(args []string) int {
	if len(args) != 1 {
		output.Println("record requires exactly one path")
		return 1
	}

	path := filepath.Clean(args[0])
	createdPath, err := miro.Record(path)
	if err != nil {
		output.Printf("%v\n", err)
		return 1
	}

	output.Println(createdPath)
	return 0
}
