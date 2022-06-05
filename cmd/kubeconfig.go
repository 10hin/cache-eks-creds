package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/10hin/cache-eks-creds/pkg/kubeconfig_resolver"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
	apiV1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

//
// kubeconfig Command
//

var (
	kubeconfigCmd = &cobra.Command{
		Use:  "kubeconfig",
		RunE: showHelpE,
	}
)

func init() {
	kubeconfigCmd.AddCommand(kubeconfigListCmd)
	kubeconfigCmd.AddCommand(kubeconfigUpdateCmd)
}

//
// list Command
//

var (
	kubeconfigListCmd = &cobra.Command{
		Use: "list",
		Aliases: []string{
			"ls",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listKubeconfigUser(cmd)
		},
	}
)

func init() {
	kubeconfigListCmd.Flags().BoolP("all", "a", false, "List all users in kubeconfig.")
}

func listKubeconfigUser(cmd *cobra.Command) error {
	var err error

	kubeconfigResolver := cmd.Context().Value(kubeconfig_resolver.Key).(kubeconfig_resolver.KubeconfigResolver)

	var kubeconfigPath string
	kubeconfigPath, err = kubeconfigResolver.Kubeconfig()
	if err != nil {
		return err
	}

	var kubeconfig *apiV1.Config
	kubeconfig, err = readKubeconfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to decode kubeconfig: %w", err)
	}

	optionAll, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	candidates := make([]*User, 0)
	for _, u := range kubeconfig.AuthInfos {
		user := &User{
			Name:     u.Name,
			IsExec:   u.AuthInfo.Exec != nil,
			Command:  Command{},
			IsCached: CachedUnknown,
		}
		if user.IsExec {
			command := Command{u.AuthInfo.Exec.Command}
			command = append(command, u.AuthInfo.Exec.Args...)
			user.Command = command
			switch filepath.Base(u.AuthInfo.Exec.Command) {
			case "cache-eks-creds":
				user.IsCached = CachedYes
			case "aws":
				user.IsCached = CachedNo
			default:
				user.IsCached = CachedUnknown
			}
		}

		if !optionAll {
			if user.IsCached == CachedUnknown {
				continue
			}
		}
		candidates = append(candidates, user)
	}

	// output
	if len(candidates) <= 0 {
		if optionAll {
			fmt.Println("no user entry found.")
			return nil
		}
		fmt.Println("no user entry found for condition. try '--all' option.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	defer func(w1 *tabwriter.Writer) {
		err1 := w.Flush()
		if err1 != nil {
			_, _ = fmt.Fprintln(os.Stderr, err1)
		}
	}(w)
	_, err = fmt.Fprintln(w, "NAME\tIS EXEC\tCOMMAND\tCACHED")
	if err != nil {
		return err
	}
	for _, c := range candidates {

		_, err = fmt.Fprintf(w, "%s\t%t\t%s\t%s\n", c.Name, c.IsExec, c.Command, c.IsCached)
	}

	return nil
}

type User struct {
	Name     string
	IsExec   bool
	Command  Command
	IsCached Cached
}

type Cached string

const (
	CachedYes     Cached = "Yes"
	CachedNo      Cached = "No"
	CachedUnknown Cached = "Unknown"
)

func (c Cached) String() string {
	return (string)(c)
}

type Command []string

func (c Command) String() string {
	if len(c) <= 0 {
		return ""
	}

	return fmt.Sprintf("'%s'", strings.Join(c, "' '"))
}

func readKubeconfig(kubeconfigPath string) (*apiV1.Config, error) {
	var err error

	var yamlBuffer bytes.Buffer
	func() {
		var f *os.File
		f, err = os.Open(kubeconfigPath)
		if err != nil {
			err = fmt.Errorf("failed to open file: %s: %w", kubeconfigPath, err)
			return
		}
		defer func(f1 *os.File) {
			err1 := f1.Close()
			if err1 != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to close file: %#v", err1)
			}
		}(f)

		_, err = io.Copy(&yamlBuffer, f)
	}()
	if err != nil {
		return nil, err
	}

	var jsonBytes []byte
	jsonBytes, err = k8sYAML.ToJSON(yamlBuffer.Bytes())

	kubeconfig := new(apiV1.Config)
	err = json.NewDecoder(bytes.NewReader(jsonBytes)).Decode(kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubeconfig, nil

}

//
// update Command
//

var (
	kubeconfigUpdateCmd = &cobra.Command{
		Use:  "update",
		RunE: updateKubeconfig,
		Args: cobra.MinimumNArgs(1),
	}
)

func init() {
	kubeconfigUpdateCmd.Flags().Bool("disable", false, "")
}

func updateKubeconfig(cmd *cobra.Command, names []string) error {
	var err error

	var kubeconfigPath string
	kubeconfigPath, err = cmd.Context().Value(kubeconfig_resolver.Key).(kubeconfig_resolver.KubeconfigResolver).Kubeconfig()
	if err != nil {
		panic(err)
	}

	var kubeconfig *apiV1.Config
	kubeconfig, err = readKubeconfig(kubeconfigPath)

	var isDisable bool
	isDisable, err = cmd.Flags().GetBool("disable")
	if err != nil {
		panic(err)
	}

	commandUpdateTo := "cache-eks-creds"
	commandUpdateFrom := "aws"
	if isDisable {
		commandUpdateTo = "aws"
		commandUpdateFrom = "cache-eks-creds"
	}

	var nameSet map[string]struct{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}

	for _, authInfo := range kubeconfig.AuthInfos {
		_, exists := nameSet[authInfo.Name]
		if !exists {
			continue
		}

		user := authInfo.AuthInfo
		execConfig := user.Exec
		if execConfig == nil {
			return fmt.Errorf(
				"specified name user %s exists,"+
					" but no \"exec\" entry. abort",
				authInfo.Name,
			)
		}

		if execConfig.Command == commandUpdateTo {
			continue
		}

		if execConfig.Command != commandUpdateFrom {
			return fmt.Errorf(
				"specified name user %s exists,"+
					" but not using \"%s\" command. abort",
				authInfo.Name,
				commandUpdateFrom,
			)
		}

		execConfig.Command = commandUpdateTo

	}

	err = writeKubeconfig(kubeconfigPath, kubeconfig)
	if err != nil {
		panic(fmt.Errorf(
			"failed to write updated kubeconfig content: %w",
			err,
		))
	}

	return nil

}

func writeKubeconfig(path string, kubeconfig *apiV1.Config) error {
	var err error

	var jsonBuf bytes.Buffer
	err = json.NewEncoder(&jsonBuf).Encode(kubeconfig)
	if err != nil {
		return err
	}

	var tree interface{}
	err = yaml.NewDecoder(&jsonBuf).Decode(&tree)
	if err != nil {
		return err
	}

	var yamlBuf bytes.Buffer
	err = yaml.NewEncoder(&yamlBuf).Encode(tree)
	if err != nil {
		return err
	}

	// To be atomic while writing kubeconfig,
	// write temporary file and rename it to path
	var f *os.File
	f, err = os.CreateTemp(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return err
	}
	defer func(f1 *os.File) {
		err1 := f1.Close()
		if err1 != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close temp file: %#v\n", err)
		}
	}(f)

	_, err = io.Copy(f, &yamlBuf)
	if err != nil {
		return err
	}

	err = os.Rename(f.Name(), path)

	return err

}
