package cache

const (
	Key = "github.com/10hin/cache-eks-creds/pkg/cache.CredentialCache"
)

type CredentialCache interface {
	Check(profile, clusterName string) (string, error)
	Update(profile, clusterName, credential string) error
	Clear(profile, clusterName string) error
}
