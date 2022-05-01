package cache

import (
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newFileCacheForTest() *fileCache {
	cacheRoot, err := os.MkdirTemp("", "cache-eks-creds-test-")
	if err != nil {
		panic(err)
	}
	return &fileCache{
		CacheRootPath: cacheRoot,
	}
}

func TestFileCache_UpdateCreatesCacheFile(t *testing.T) {
	cacheStore := newFileCacheForTest()

	profile := "test-profile"
	clusterName := "test-cluster"
	result := prepareCredentialContent(metav1.NewTime(metav1.Now().Add(15*time.Minute)), "")

	err := cacheStore.Update(profile, clusterName, result)
	if err != nil {
		t.Fatalf("fileCache.Update failed")
	}

	assertFileContent(t, filepath.Clean(fmt.Sprintf("%s/%s/%s", cacheStore.CacheRootPath, profile, clusterName)), result)

}

func TestFileCache_CheckFailsWithNoCacheFile(t *testing.T) {
	cacheStore := newFileCacheForTest()

	profile := "profile-without-dir"
	clusterName := "test-cluster"

	_, err := cacheStore.Check(profile, clusterName)
	if err == nil {
		t.Fatal("expected Check fails, but in actual, succeeds")
	}
}

func TestFileCache_CheckSucceedsWithCacheFile(t *testing.T) {
	cacheStore := newFileCacheForTest()

	profile := "test-profile"
	clusterName := "test-cluster"
	cacheFilePath := filepath.Clean(fmt.Sprintf(
		"%s/%s/%s",
		cacheStore.CacheRootPath,
		profile,
		clusterName,
	))
	result := prepareCredentialContent(metav1.NewTime(metav1.Now().Add(15*time.Minute)), "")
	prepareFileWithContent(
		cacheFilePath,
		result,
		0755,
	)

	_, err := cacheStore.Check(profile, clusterName)
	if err != nil {
		t.Fatal("expected Check succeeds, but in actual, failed:", err)
	}
}

func prepareCredentialContent(expires metav1.Time, token string) string {
	expiresBytes, err := expires.MarshalText()
	if err != nil {
		panic(err)
	}
	var encoded []byte
	encoded, err = json.Marshal(map[string]interface{}{
		"apiVersion": "client.authentication.k8s.io/v1alpha1",
		"kind":       "ExecCredential",
		"spec":       map[string]interface{}{},
		"status": map[string]interface{}{
			"expirationTimestamp": (string)(expiresBytes),
			"token":               token,
		},
	})
	if err != nil {
		panic(err)
	}

	return (string)(encoded)
}

func prepareFileWithContent(path, content string, basemode os.FileMode) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, basemode)
	if err != nil {
		panic(err)
	}

	var f *os.File
	f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, basemode&0666)
	if err != nil {
		panic(err)
	}
	defer func(f1 *os.File) { _ = f1.Close() }(f)

	_, err = f.WriteString(content)
	if err != nil {
		panic(err)
	}
}

func assertFileContent(t *testing.T, expectedPath string, expectedContent string) {
	f, err := os.Open(expectedPath)
	if err != nil {
		t.Fatalf("expected path %s failed to open", expectedPath)
	}
	defer func(f1 *os.File) { _ = f1.Close() }(f)

	var contentBytes []byte
	contentBytes, err = ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("expected path %s failed to read all content", expectedPath)
	}

	actualContent := (string)(contentBytes)
	if actualContent != expectedContent {
		t.Fatalf("expected content %s different from actual content %s", expectedContent, actualContent)
	}
}
