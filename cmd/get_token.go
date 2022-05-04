package cmd

import (
	"bytes"
	"fmt"
	"github.com/10hin/cache-eks-creds/pkg/cache"
	"github.com/10hin/cache-eks-creds/pkg/profile_resolver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"os/exec"
)

const (
	appCacheDir       = "cache-eks-creds"
	cliCommandName    = "aws"
	cliServiceCmdName = "eks"
	cliActionCmdName  = "get-token"
)

var (
	getTokenCmd = &cobra.Command{
		Use: "get-token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getToken(cmd)
		},
	}
)

func init() {
	getTokenCmd.PersistentFlags().String("cluster-name", "", "Specify the name of the Amazon EKS cluster to create a token for.")
	_ = getTokenCmd.MarkPersistentFlagRequired("cluster-name")
	getTokenCmd.PersistentFlags().String("role-arn", "", "Assume this role for credentials when signing the token.")
}

func getToken(cmd *cobra.Command) error {
	var err error

	cacheStore := cmd.Context().Value(cache.Key).(cache.CredentialCache)

	profileResolver := cmd.Context().Value(profile_resolver.Key).(*profile_resolver.ProfileResolver)
	profile, err := profileResolver.Profile()
	if err != nil {
		panic(err)
	}

	var clusterName string
	clusterName, err = cmd.PersistentFlags().GetString("cluster-name")
	if err != nil {
		return err
	}

	var cacheContent string
	cacheContent, err = cacheStore.Check(profile, clusterName)
	if err == nil {
		fmt.Println(cacheContent)
		return nil
	}

	// ignore error while checking cache

	var execResult string
	execResult, err = executeAWSCLI(cmd)

	fmt.Println(execResult)

	err = cacheStore.Update(profile, clusterName, execResult)
	if err != nil {
		fmt.Printf("Warning: error happend whild updating cache: %#v\n", err)
		// do not return error
	}

	return nil
}

func executeAWSCLI(cmd *cobra.Command) (string, error) {
	var err error

	rFlags := make([]string, 0)
	rCmd := cmd.Root()
	rCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		//fmt.Println("rCmd:PersistentFlags:", flag.Name, flag.Value.String())
		switch flag.Value.Type() {
		case "bool":
			if flag.Value.String() == "true" {
				rFlags = append(rFlags, "--"+flag.Name)
			}
		case "string":
			if flag.Value.String() != "" {
				rFlags = append(rFlags, "--"+flag.Name, flag.Value.String())
			}
		case "int":
			if flag.Value.String() != "-1" {
				rFlags = append(rFlags, "--"+flag.Name, flag.Value.String())
			}
		}
	})

	pFlags := make([]string, 0)
	pCmd := cmd.Parent()
	pCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		//fmt.Println("pCmd:PersistentFlags:", flag.Name, flag.Value.String())
		switch flag.Value.Type() {
		case "bool":
			if flag.Value.String() == "true" {
				pFlags = append(pFlags, "--"+flag.Name)
			}
		case "string":
			if flag.Value.String() != "" {
				pFlags = append(pFlags, "--"+flag.Name, flag.Value.String())
			}
		case "int":
			if flag.Value.String() != "-1" {
				pFlags = append(pFlags, "--"+flag.Name, flag.Value.String())
			}
		}
	})

	tFlags := make([]string, 0)
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		//fmt.Println("cmd:PersistentFlags:", flag.Name, flag.Value.String())
		switch flag.Value.Type() {
		case "bool":
			if flag.Value.String() == "true" {
				tFlags = append(tFlags, "--"+flag.Name)
			}
		case "string":
			if flag.Value.String() != "" {
				tFlags = append(tFlags, "--"+flag.Name, flag.Value.String())
			}
		case "int":
			if flag.Value.String() != "-1" {
				tFlags = append(tFlags, "--"+flag.Name, flag.Value.String())
			}
		}
	})

	fmt.Println("[DEBUG] -- before exec.Command(string, ...string)")

	args := make([]string, 0, 2+len(rFlags)+len(pFlags)+len(tFlags))
	args = append(args, rFlags...)
	args = append(args, cliServiceCmdName)
	args = append(args, pFlags...)
	args = append(args, cliActionCmdName)
	args = append(args, tFlags...)
	fmt.Printf("args: %v\n", args)
	cliCmd := exec.Command(cliCommandName, args...)

	var cliRawStdout io.ReadCloser
	cliRawStdout, err = cliCmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	var cliStdout bytes.Buffer
	go func(w io.Writer, r io.Reader) {
		_, err1 := io.Copy(w, r)
		if err1 != nil && err1 != io.EOF {
			fmt.Println("cliStdout copying routine encounter error other than io.EOF")
		}
	}(&cliStdout, cliRawStdout)

	var cliRawStderr io.ReadCloser
	cliRawStderr, err = cliCmd.StderrPipe()
	if err != nil {
		return "", err
	}
	var cliStderr bytes.Buffer
	go func(w io.Writer, r io.Reader) {
		_, err1 := io.Copy(w, r)
		if err1 != nil && err1 != io.EOF {
			fmt.Println("cliStderr copying routine encounter error other than io.EOF")
		}
	}(&cliStderr, cliRawStderr)

	fmt.Println("[DEBUG] -- before Command.Run()")

	err = cliCmd.Run()
	if err != nil {
		fmt.Println("-- stdout --")
		fmt.Println(cliStdout.String())
		fmt.Println("------------")
		fmt.Println("-- stderr --")
		fmt.Println(cliStderr.String())
		fmt.Println("------------")
		return "", err
	}

	return cliStdout.String(), nil
}
