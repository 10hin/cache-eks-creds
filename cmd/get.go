package cmd

import "github.com/spf13/cobra"

var (
	getCmd = &cobra.Command{
		Use: "cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	getCmd.AddCommand(getCacheCmd)
}
