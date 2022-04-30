package cmd

import "github.com/spf13/cobra"

var (
	eksCmd = &cobra.Command{
		Use: "eks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	eksCmd.AddCommand(getTokenCmd)
}
