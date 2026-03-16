package cmd

import "github.com/spf13/cobra"

// Run executes the miro CLI and returns a process exit code.
func Run(args []string) int {
	rootCmd := newRootCommand()
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		return 1
	}

	return 0
}

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "miro",
		Short: "A lean CLI E2E testing framework.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(newInitCommand(), newRecordCommand())

	return rootCmd
}
