package cache

import (
	"encoding/json"
	"fmt"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthn "k8s.io/client-go/pkg/apis/clientauthentication"
	"os"
	"path/filepath"
	"strings"
)

const (
	appCacheDir = "cache-eks-creds"
)

type fileCache struct{}

func NewFileCache() CredentialCache {
	return &fileCache{}
}

func buildCachePath(profile, clusterName string) (string, error) {
	var err error

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Clean(fmt.Sprintf("%s/%s/%s/%s", userCacheDir, appCacheDir, profile, clusterName)), nil

}

func (c *fileCache) Check(profile, clusterName string) (string, error) {
	cachePath, err := buildCachePath(profile, clusterName)
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
	cachePath, err := buildCachePath(profile, clusterName)
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cachePath)
	err = os.MkdirAll(cacheDir, 0700)
	if err != nil {
		return err
	}

	var f *os.File
	f, err = os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer func(f1 *os.File) {
		err1 := f1.Close()
		if err1 != nil {
			fmt.Printf("error when cache file closing: %#v\n", err)
		}
	}(f)

	_, err = f.WriteString(result)
	return err
}

func (c *fileCache) Clear(profile, clusterName string) error {
	cachePath, err := buildCachePath(profile, clusterName)
	if err != nil {
		return err
	}

	return os.Remove(cachePath)
}
