package cmd

import (
	"errors"
	"path/filepath"

	"miro/internal/miro"
	"miro/internal/output"

	"github.com/spf13/cobra"
)

func newRecordCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "record <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := filepath.Clean(args[0])
			createdPath, err := miro.Record(path)
			if err != nil {
				if errors.Is(err, miro.ErrRecordingDiscarded) {
					return nil
				}
				return err
			}

			output.Println(createdPath)
			return nil
		},
	}
}
