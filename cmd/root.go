package cmd

import (
	"context"
	"github.com/10hin/cache-eks-creds/pkg/kubeconfig_resolver"
	"github.com/10hin/cache-eks-creds/pkg/profile_resolver"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "cache-eks-creds",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			profileResolver := cmd.Context().Value(profile_resolver.Key).(*profile_resolver.ProfileResolver)
			profileResolver.SetFlagHolder(cmd)
			err = profileResolver.Resolve()
			if err != nil {
				return err
			}

			kubeconfigResolver := cmd.Context().Value(kubeconfig_resolver.Key).(kubeconfig_resolver.KubeconfigResolver)
			kubeconfigResolver.SetFlagHolder(cmd)
			err = kubeconfigResolver.Resolve()
			if err != nil {
				return err
			}

			return nil
		},
		RunE: showHelpE,
	}
)

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// flags compatible to aws-cli
	rootCmd.PersistentFlags().Bool("debug", false, "Turn on debug logging.")
	rootCmd.PersistentFlags().String("endpoint-url", "", "Override command's default URL with the given URL.")
	rootCmd.PersistentFlags().Bool("no-verify-ssl", false, "By default, the AWS CLI uses SSL when communicating with AWS services. For each SSL connection, the AWS CLI will verify SSL certificates. This option overrides the default behavior of verifying SSL certificates.")
	rootCmd.PersistentFlags().Bool("no-paginate", false, "Disable automatic pagination.")
	rootCmd.PersistentFlags().String("output", "", "The formatting style for command output.")
	rootCmd.PersistentFlags().String("query", "", "A JMESPath query to use in filtering the response data.")
	rootCmd.PersistentFlags().String("profile", "", "Use a specific profile from your credential file.")
	rootCmd.PersistentFlags().String("region", "", "The region to use. Overrides config/env settings.")
	rootCmd.PersistentFlags().String("version", "", "Display the version of this tool.")
	rootCmd.PersistentFlags().String("color", "", "Turn on/off color output.")
	rootCmd.PersistentFlags().Bool("no-sign-request", false, "Do not sign requests. Credentials will not be loaded if this argument is provided.")
	rootCmd.PersistentFlags().String("ca-bandle", "", "The CA certificate bundle to use when verifying SSL certificates. Overrides config/env settings.")
	rootCmd.PersistentFlags().Int("cli-read-timeout", -1, "The maximum socket read time in seconds. If the value is set to 0, the socket read will be blocking and not timeout. The default value is 60 seconds.")
	rootCmd.PersistentFlags().Int("cli-connect-timeout", -1, "The maximum socket connect time in seconds. If the value is set to 0, the socket connect will be blocking and not timeout. The default value is 60 seconds.")
	// flags compatible to kubectl
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")

	rootCmd.AddCommand(eksCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(kubeconfigCmd)
}

func showHelpE(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
