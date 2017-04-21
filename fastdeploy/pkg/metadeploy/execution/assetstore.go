package execution

import (
	"crypto/sha256"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/utils"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Asset struct {
	Key       string
	LocalPath string
	AssetPath string
}

type AssetStore struct {
	assetDir string
	assets   []*Asset
}

func NewAssetStore(assetDir string) *AssetStore {
	a := &AssetStore{
		assetDir: assetDir,
	}
	return a
}
func (a *AssetStore) Find(key string, assetPath string) (*Asset, error) {
	var matches []*Asset
	for _, asset := range a.assets {
		if asset.Key != key {
			continue
		}

		if assetPath != "" {
			if !strings.HasSuffix(asset.AssetPath, assetPath) {
				continue
			}
		}

		matches = append(matches, asset)
	}

	if len(matches) == 0 {
		return nil, nil
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return nil, fmt.Errorf("found multiple matching assets for key: %q", key)
}

func hashFromHttpHeader(url string) (string, error) {
	glog.Infof("Doing HTTP HEAD on %q", url)
	response, err := http.Head(url)
	if err != nil {
		return "", fmt.Errorf("error doing HEAD on %q: %v", url, err)
	}
	defer response.Body.Close()

	etag := response.Header.Get("ETag")
	etag = strings.TrimSpace(etag)
	etag = strings.Trim(etag, "'\"")

	if etag != "" {
		if len(etag) == 32 {
			// Likely md5
			return etag, nil
		}
	}

	return "", fmt.Errorf("unable to determine hash from HTTP HEAD: %q", url)
}

func (a *AssetStore) AddURL(url string, hash string) error {
	var err error

	if hash == "" {
		hash, err = hashFromHttpHeader(url)
		if err != nil {
			return err
		}
	}

	localFile := path.Join(a.assetDir, hash+"_"+utils.SanitizeString(url))
	_, err = utils.DownloadURL(url, localFile, hash)
	if err != nil {
		return err
	}

	key := path.Base(url)
	err = a.addAsset(key, localFile, url)
	if err != nil {
		return err
	}

	if strings.HasSuffix(key, ".tar.gz") {
		err = a.addArchive(localFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AssetStore) AddFile(f string) error {
	hash, err := utils.HashFile(f, sha256.New())
	if err != nil {
		return err
	}

	key := path.Base(f)

	localFile := path.Join(a.assetDir, hash+"_"+utils.SanitizeString(key))
	_, err = utils.CopyFileIfDifferent(localFile, f, 0644, 0755)
	if err != nil {
		return err
	}

	err = a.addAsset(key, localFile, f)
	if err != nil {
		return err
	}

	if strings.HasSuffix(key, ".tar.gz") {
		err = a.addArchive(localFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AssetStore) addArchive(archiveFile string) error {
	extracted := path.Join(a.assetDir, "extracted/"+path.Base(archiveFile))

	// TODO: Use a temp file so this is atomic
	if _, err := os.Stat(extracted); os.IsNotExist(err) {
		err := os.MkdirAll(extracted, 0755)
		if err != nil {
			return fmt.Errorf("error creating directories %q: %v", path.Dir(extracted), err)
		}

		args := []string{"tar", "zxf", archiveFile, "-C", extracted}
		glog.Infof("running extract command %s", args)
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error expanding asset file %q %v: %s", archiveFile, err, string(output))
		}
	}

	localBase := extracted
	assetBase := ""

	walker := func(localPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error descending into path %q: %v", localPath, err)
		}

		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(localBase, localPath)
		if err != nil {
			return fmt.Errorf("error finding relative path for %q: %v", localPath, err)
		}

		assetPath := path.Join(assetBase, relativePath)

		err = a.addAsset(info.Name(), localPath, assetPath)
		if err != nil {
			return err
		}

		return nil
	}

	err := filepath.Walk(localBase, walker)
	if err != nil {
		return fmt.Errorf("error adding expanded asset fils in tree %q: %v", extracted, err)
	}
	return nil

}

func (a *AssetStore) addAsset(key, localPath string, assetPath string) error {
	asset := &Asset{
		Key:       key,
		LocalPath: localPath,
		AssetPath: assetPath,
	}
	glog.V(2).Infof("added asset %q for %q", key, assetPath)
	a.assets = append(a.assets, asset)
	return nil
}

func (a *Asset) WriteTo(destPath string, fileMode, dirMode os.FileMode) (bool, error) {
	changed, err := utils.CopyFileIfDifferent(destPath, a.LocalPath, fileMode, dirMode)
	if err != nil {
		return changed, fmt.Errorf("error copying asset to %q: %v", destPath, err)
	}

	return changed, nil
}
