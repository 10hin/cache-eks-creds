package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/10hin/cache-eks-creds/pkg/kubeconfig_resolver"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	"io"
	apiV1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

var (
	kubeconfigCmd = &cobra.Command{
		Use: "kubeconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	kubeconfigCmd.AddCommand(kubeconfigListCmd)
}

var (
	kubeconfigListCmd = &cobra.Command{
		Use: "list",
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

	var f *os.File
	f, err = os.Open(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s: %w", kubeconfigPath, err)
	}
	defer func(f1 *os.File) {
		err1 := f1.Close()
		if err1 != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close file: %#v", err1)
		}
	}(f)

	var obj interface{}
	err = yaml.NewDecoder(f).Decode(&obj)
	if err != nil {
		return nil, err
	}

	pipeOut, pipeIn := io.Pipe()
	eg := &errgroup.Group{}
	eg.Go(func() error {
		return json.NewEncoder(pipeIn).Encode(&obj)
	})
	res := make(chan *apiV1.Config, 1)
	eg.Go(func() error {
		dec := new(apiV1.Config)
		var err1 error
		err1 = json.NewDecoder(pipeOut).Decode(dec)
		res <- dec
		return err1
	})

	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	return <-res, nil
}
