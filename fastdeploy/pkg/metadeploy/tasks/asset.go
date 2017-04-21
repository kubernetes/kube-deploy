package tasks

import (
	"encoding/json"
	"fmt"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/utils"
	"os"
	"path"
)

type AssetTask struct {
	Name      string
	DestPath  string

	AssetPath string `json:"assetPath"`
	Mode      string `json:"mode"`
}

func (a *AssetTask) SortKey() (int, string) {
	return TaskOrderTemplate, a.DestPath
}

func (a *AssetTask) String() string {
	return fmt.Sprintf("Asset: %s", a.DestPath)
}

func NewAsset(name string, destPath string, contents string) (*AssetTask, error) {
	a := &AssetTask{
		Name:     name,
		DestPath: destPath,
	}

	if contents != "" {
		err := json.Unmarshal([]byte(contents), a)
		if err != nil {
			return nil, fmt.Errorf("error parsing json for asset %q: %v", name, err)
		}
	}

	return a, nil
}

func (a *AssetTask) Run(c *execution.Context) error {
	key := path.Base(a.DestPath)
	asset, err := c.Assets.Find(key, a.AssetPath)
	if err != nil {
		return fmt.Errorf("error trying to locate asset %q: %v", key, err)
	}

	if asset == nil {
		return fmt.Errorf("unable to locate asset %q", key)
	}

	dirMode := os.FileMode(0755)
	fileMode, err := utils.ParseFileMode(a.Mode, 0644)
	if err != nil {
		return fmt.Errorf("invalid file mode for %q: %q", a.DestPath, a.Mode)
	}

	_, err = asset.WriteTo(a.DestPath, fileMode, dirMode)
	if err != nil {
		return err
	}

	return nil
}
