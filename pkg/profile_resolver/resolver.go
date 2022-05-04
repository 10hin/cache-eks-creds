package profile_resolver

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const (
	Key                          = "github.com/10hin/cache-eks-creds/pkg/profile_resolver.ProfileResolver"
	cliEnvKeyDefaultProfile      = "AWS_DEFAULT_PROFILE"
	cliEnvKeyProfile             = "AWS_PROFILE"
	cliFactoryDefaultProfileName = "default"
)

type ProfileResolver struct {
	resolved   bool
	profile    string
	flagHolder *cobra.Command
}

func NewProfileResolver() *ProfileResolver {
	return &ProfileResolver{
		resolved:   false,
		profile:    "",
		flagHolder: nil,
	}
}

func (r *ProfileResolver) Profile() (string, error) {
	if !r.resolved {
		return "", fmt.Errorf("not resolved yet")
	}
	return r.profile, nil
}

func (r *ProfileResolver) SetFlagHolder(cmd *cobra.Command) {
	if r.resolved {
		panic(fmt.Errorf("ProfileResolver cannot modify flagHolder after Resolve()"))
	}
	r.flagHolder = cmd
}

func (r *ProfileResolver) Resolve() error {
	if r.resolved {
		return nil
	}
	r.resolved = true

	var err error
	r.profile, err = r.flagHolder.Root().PersistentFlags().GetString("profile")
	if err != nil {
		return err
	}
	if r.profile != "" {
		return nil
	}

	r.profile = os.Getenv(cliEnvKeyDefaultProfile)
	if r.profile != "" {
		return nil
	}

	r.profile = os.Getenv(cliEnvKeyProfile)
	if r.profile != "" {
		return nil
	}

	r.profile = cliFactoryDefaultProfileName
	return nil
}
