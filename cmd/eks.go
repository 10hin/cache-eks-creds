package cmd

import "github.com/spf13/cobra"

var (
	eksCmd = &cobra.Command{
		Use:  "eks",
		RunE: showHelpE,
	}
)

func init() {
	eksCmd.AddCommand(getTokenCmd)
}
