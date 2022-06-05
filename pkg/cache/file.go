package cache

import (
	"encoding/json"
	"fmt"
	"github.com/10hin/cache-eks-creds/pkg/write_rename"
	"io"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthn "k8s.io/client-go/pkg/apis/clientauthentication"
)

const (
	appCacheDir = "cache-eks-creds"
)

type fileCache struct {
	CacheRootPath string
}

func NewFileCache() CredentialCache {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	return &fileCache{
		CacheRootPath: filepath.Clean(userCacheDir + "/" + appCacheDir),
	}
}

func (c *fileCache) buildCachePath(profile, clusterName string) (string, error) {
	return filepath.Clean(fmt.Sprintf("%s/%s/%s", c.CacheRootPath, profile, clusterName)), nil

}

func (c *fileCache) Check(profile, clusterName string) (string, error) {
	cachePath, err := c.buildCachePath(profile, clusterName)
	if err != nil {
		return "", err
	}

	var f *os.File
	f, err = os.Open(cachePath)
	if err != nil {
		return "", err
	}
	defer func(f1 *os.File) {
		err1 := f1.Close()
		if err1 != nil {
			fmt.Printf("error when cache file closing: %#v\n", err)
		}
	}(f)

	// read only to check expiration.
	// use buf to returning.
	// because re-encoding will apply unintended case conversion.
	var buf strings.Builder
	read := io.TeeReader(f, &buf)

	var execCred clientauthn.ExecCredential
	err = json.NewDecoder(read).Decode(&execCred)
	if err != nil {
		return "", err
	}

	now := metav1.Now()
	expire := execCred.Status.ExpirationTimestamp
	if expire.Before(&now) {
		return "", fmt.Errorf("cached credential already expired")
	}

	return buf.String(), nil
}

func (c *fileCache) Update(profile, clusterName string, result string) error {
	cachePath, err := c.buildCachePath(profile, clusterName)
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cachePath)
	err = os.MkdirAll(cacheDir, 0700)
	if err != nil {
		return err
	}

	return write_rename.WriteRename(cachePath, strings.NewReader(result))
}

func (c *fileCache) Clear(profile, clusterName string) error {
	cachePath, err := c.buildCachePath(profile, clusterName)
	if err != nil {
		return err
	}

	return os.Remove(cachePath)
}
