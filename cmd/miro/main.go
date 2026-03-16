package main

import (
	"os"

	"miro/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
