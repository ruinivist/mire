package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"mire/internal/mire"
	"mire/internal/output"

	"github.com/spf13/cobra"
)

func newRecordCommand() *cobra.Command {
	var save bool

	cmd := &cobra.Command{
		Use:   "record <path>",
		Short: "Record a new CLI scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := filepath.Clean(args[0])
			createdPath, err := mire.Record(path, mire.RecordOptions{Save: save})
			if err != nil {
				if errors.Is(err, mire.ErrRecordingDiscarded) {
					return nil
				}
				cmd.SilenceUsage = true
				return err
			}

			displayPath := createdPath
			if cwd, err := os.Getwd(); err == nil {
				if relPath, err := filepath.Rel(cwd, createdPath); err == nil {
					displayPath = relPath
				}
			}

			output.Printf("Saved at %s\n", displayPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&save, "save", false, "Save the recording without prompting")

	return cmd
}
