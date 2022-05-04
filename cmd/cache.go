package cmd

import (
	"fmt"
	"github.com/10hin/cache-eks-creds/pkg/cache"
	"github.com/10hin/cache-eks-creds/pkg/profile_resolver"
	"github.com/spf13/cobra"
)

var (
	deleteCacheCmd = &cobra.Command{
		Use: "cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteCache(cmd, args[0])
		},
		Args: cobra.ExactArgs(1),
	}
)

func deleteCache(cmd *cobra.Command, clusterName string) error {
	var err error

	cacheStore := cmd.Context().Value(cache.Key).(cache.CredentialCache)

	profileResolver := cmd.Context().Value(profile_resolver.Key).(*profile_resolver.ProfileResolver)
	var profile string
	profile, err = profileResolver.Profile()
	if err != nil {
		panic(err)
	}

	return cacheStore.Clear(profile, clusterName)
}

var (
	getCacheCmd = &cobra.Command{
		Use: "cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return listCache(cmd)
			}
			return getCache(cmd, args[0])
		},
		Args: cobra.MaximumNArgs(1),
	}
)

func init() {
	getCacheCmd.PersistentFlags().String("cluster-name", "", "Specify the name of the Amazon EKS cluster to delete cache for.")
	_ = getCacheCmd.MarkPersistentFlagRequired("cluster-name")
}

// TODO implement
func listCache(cmd *cobra.Command) error {

	//cacheStrore := cmd.Context().Value(cache.Key).(cache.CredentialCache)

	return fmt.Errorf("currently listing cache entry is not supported") // TODO implement
}

// TODO complete implementation
func getCache(cmd *cobra.Command, clusterName string) error {
	var err error

	cacheStore := cmd.Context().Value(cache.Key).(cache.CredentialCache)

	profileResolver := cmd.Context().Value(profile_resolver.Key).(*profile_resolver.ProfileResolver)
	var profile string
	profile, err = profileResolver.Profile()
	if err != nil {
		panic(err)
	}

	// TODO implement printing cached content
	_, err = cacheStore.Check(profile, clusterName)
	return err // TODO implement
}
