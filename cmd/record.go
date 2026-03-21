package cmd

import (
	"errors"
	"path/filepath"

	"mire/internal/mire"
	"mire/internal/output"

	"github.com/spf13/cobra"
)

func newRecordCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "record <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := filepath.Clean(args[0])
			createdPath, err := mire.Record(path)
			if err != nil {
				if errors.Is(err, mire.ErrRecordingDiscarded) {
					return nil
				}
				return err
			}

			output.Println(createdPath)
			return nil
		},
	}
}
