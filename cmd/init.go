package cmd

import (
	"miro/internal/miro"
	"miro/internal/output"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "init",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := miro.Init(); err != nil {
				return err
			}

			output.Println("Done initialising...")
			return nil
		},
	}
}
