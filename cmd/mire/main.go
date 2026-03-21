package main

import (
	"os"

	"mire/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
