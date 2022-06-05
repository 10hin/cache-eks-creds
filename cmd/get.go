package cmd

import "github.com/spf13/cobra"

var (
	getCmd = &cobra.Command{
		Use:  "cache",
		RunE: showHelpE,
	}
)

func init() {
	getCmd.AddCommand(getCacheCmd)
}
