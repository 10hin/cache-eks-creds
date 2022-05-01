package cache

type CredentialCache interface {
	Check(profile, clusterName string) (string, error)
	Update(profile, clusterName, credential string) error
	Clear(profile, clusterName string) error
}
