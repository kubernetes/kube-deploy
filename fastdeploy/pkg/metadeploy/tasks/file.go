package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/utils"
	"os"
)

type File struct {
	Name     string
	DestPath string
	SrcPath  string

	Mode        string `json:"mode"`
	IfNotExists bool   `json:"ifNotExists"`
}

func NewFileTask(name string, srcPath string, destPath string, meta string) (*File, error) {
	f := &File{
		Name:     name,
		SrcPath:  srcPath,
		DestPath: destPath,
	}

	if meta != "" {
		err := json.Unmarshal([]byte(meta), f)
		if err != nil {
			return nil, fmt.Errorf("error parsing meta for file %q: %v", name, err)
		}
	}

	return f, nil
}

func (f *File) SortKey() (int, string) {
	return TaskOrderTemplate, f.DestPath
}

func (f *File) String() string {
	return fmt.Sprintf("File: %q -> %q", f.SrcPath, f.DestPath)
}

func (f *File) Run(c *execution.Context) error {
	dirMode := os.FileMode(0755)
	fileMode, err := utils.ParseFileMode(f.Mode, 0644)
	if err != nil {
		return fmt.Errorf("invalid file mode for %q: %q", f.DestPath, f.Mode)
	}

	// When IfNotExists is set, don't do anything if the file exists
	if f.IfNotExists {
		_, err := os.Stat(f.DestPath)
		if err == nil {
			glog.V(2).Infof("file exists and IfNotExists set; skipping %q", f.DestPath)
			return nil
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking file %q: %v", f.DestPath, err)
		}
	}

	_, err = utils.CopyFileIfDifferent(f.DestPath, f.SrcPath, fileMode, dirMode)
	if err != nil {
		return fmt.Errorf("error copying file %q: %v", f.DestPath, err)
	}

	return nil
}
