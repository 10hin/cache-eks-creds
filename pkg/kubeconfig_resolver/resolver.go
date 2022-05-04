package kubeconfig_resolver

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	Key                     = "github.com/10hin/cache-eks-creds/pkg/kubeconfig_resolver.KubeconfigResolver"
	kubectlEnvKeyKubeconfig = "KUBECONFIG"
	kubectlDefaultSubPath   = ".kube/config"
)

type KubeconfigResolver interface {
	Kubeconfig() (string, error)
	SetFlagHolder(cmd *cobra.Command)
	Resolve() error
}

type kubeconfigResolver struct {
	resolved   bool
	kubeconfig string
	flagHolder *cobra.Command
}

func NewKubeconfigResolver() KubeconfigResolver {
	return &kubeconfigResolver{
		resolved:   false,
		kubeconfig: "",
		flagHolder: nil,
	}
}

func (r *kubeconfigResolver) Kubeconfig() (string, error) {
	if !r.resolved {
		return "", fmt.Errorf("not resolved yet")
	}
	return r.kubeconfig, nil
}

func (r *kubeconfigResolver) SetFlagHolder(cmd *cobra.Command) {
	if r.resolved {
		panic(fmt.Errorf("ProfileResolver cannot modify flagHolder after Resolve()"))
	}
	r.flagHolder = cmd
}

func (r *kubeconfigResolver) Resolve() error {
	if r.resolved {
		return nil
	}
	r.resolved = true

	var err error
	r.kubeconfig, err = r.flagHolder.Root().PersistentFlags().GetString("kubeconfig")
	if err != nil {
		return err
	}
	if r.kubeconfig != "" {
		return nil
	}

	r.kubeconfig = os.Getenv(kubectlEnvKeyKubeconfig)
	if r.kubeconfig != "" {
		return nil
	}

	r.kubeconfig = kubectlKubeconfigDefault()
	return nil
}

func kubectlKubeconfigDefault() string {
	var err error
	var homePath string
	homePath, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Clean(homePath + "/" + kubectlDefaultSubPath)
}
