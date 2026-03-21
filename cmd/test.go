package cmd

import (
	"mire/internal/mire"

	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "test [path]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}

			return mire.RunTests(path)
		},
	}
}
