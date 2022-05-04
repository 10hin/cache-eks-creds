package cmd

import "github.com/spf13/cobra"

var (
	kubeconfigCmd = &cobra.Command{
		Use: "kubeconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil // TODO implement
		},
	}
)
