package cmd

import "github.com/spf13/cobra"

var (
	deleteCmd = &cobra.Command{
		Use:  "delete",
		RunE: showHelpE,
	}
)

func init() {
	deleteCmd.AddCommand(deleteCacheCmd)
}
