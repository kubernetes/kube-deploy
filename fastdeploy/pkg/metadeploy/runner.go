package metadeploy

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/tasks"
	"os"
	"path/filepath"
	"strings"
)

type Runner struct {
	Basedir string
	Tags    map[string]struct{}

	Tasks tasks.TaskList
}

func (r *Runner) Run() error {
	err := filepath.Walk(r.Basedir, r.treewalker)
	if err != nil {
		return fmt.Errorf("error walking tree %q: %v", r.Basedir, err)
	}

	tasks.SortTasks(r.Tasks)

	return nil
}

func (r *Runner) treewalker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error descending into path %q: %v", path, err)
	}

	glog.V(4).Infof("visit %q", path)
	name := info.Name()
	if len(name) == 0 {
		return fmt.Errorf("unexpected empty name: %q", path)
	}

	if IsTag(name) {
		_, found := r.Tags[name]
		if !found {
			glog.V(2).Infof("Skipping directory as tag not present: %q", path)
			return filepath.SkipDir
		} else {
			glog.V(2).Infof("Descending into directory, as tag is present: %q", path)
		}
	}

	switch name {
	case "files", "options", "packages", "services", "sysctls":
		allowTags := name != "files"
		err = r.collectTasks(path, name, allowTags)
		if err != nil {
			return fmt.Errorf("error building %s at %q: %v", name, path, err)
		}
		return filepath.SkipDir
	}

	// Otherwise we assume this is just an organizational directory

	return nil
}

func IsTag(name string) bool {
	return len(name) != 0 && name[0] == '_'
}

func (r *Runner) collectTasks(base string, key string, allowTags bool) error {
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error descending into path %q: %v", path, err)
		}

		glog.V(4).Infof("%s visit %q", key, path)
		name := info.Name()
		if len(name) == 0 {
			return fmt.Errorf("unexpected empty name: %q", path)
		}

		if IsTag(name) {
			if allowTags {
				_, found := r.Tags[name]
				if !found {
					glog.V(2).Infof("Skipping directory as tag not present: %q", path)
					return filepath.SkipDir
				} else {
					glog.V(2).Infof("Descending into directory, as tag is present: %q", path)
				}
			} else {
				glog.Warningf("Found directory %q that looks like a tag, but not allowed in %s content; treating as a directory", name, key)
			}
		}

		if info.IsDir() {
			// Continue the descent
			// TODO: Should we mkdirs for empty dirs?
			// TODO: Should we warn if there are empty dirs?
			return nil
		}

		if strings.HasSuffix(name, ".meta") {
			// We'll read it when we see the actual file
			// But check the actual file is there
			primaryPath := strings.TrimSuffix(path, ".meta")
			if _, err := os.Stat(primaryPath); os.IsNotExist(err) {
				return fmt.Errorf("found .meta file without corresponding file: %q", path)
			}

			return nil
		}

		contentBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %q: %v", path, err)
		}
		contents := string(contentBytes)

		var meta string
		{
			metaBytes, err := ioutil.ReadFile(path + ".meta")
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("error reading file %q: %v", path, err)
				}
				metaBytes = nil
			}
			if metaBytes != nil {
				meta = string(metaBytes)
			}
		}

		relativePath, err := filepath.Rel(base, path)
		if err != nil {
			return fmt.Errorf("error finding relative path for %q: %v", path, err)
		}

		var task tasks.Task

		switch key {
		case "options":
			task, err = tasks.NewOptionsTask(name, contents)
		case "packages":
			task, err = tasks.NewPackage(name, contents)
		case "services":
			task, err = tasks.NewService(name, contents, meta)
		case "sysctls":
			task, err = tasks.NewSysctl(name, contents)
		case "files":
			if strings.HasSuffix(name, ".template") {
				destPath := "/" + strings.TrimSuffix(relativePath, ".template")
				task, err = tasks.NewTemplate(strings.TrimSuffix(name, ".template"), destPath, contents, meta)
			} else if strings.HasSuffix(name, ".asset") {
				destPath := "/" + strings.TrimSuffix(relativePath, ".asset")
				task, err = tasks.NewAsset(strings.TrimSuffix(name, ".asset"), destPath, contents)
			} else {
				task, err = tasks.NewFileTask(name, path, "/"+relativePath, meta)
			}
		default:
			panic(fmt.Sprintf("unhandled task type: %q", key))
		}

		if err != nil {
			return fmt.Errorf("error building %s for %q: %v", key, path, err)
		}

		glog.V(2).Infof("path %q -> task %v", path, task)

		if task != nil {
			r.Tasks = append(r.Tasks, task)
		}
		return nil
	}
	err := filepath.Walk(base, walker)
	if err != nil {
		return fmt.Errorf("error walking %s tree %q: %v", key, base, err)
	}
	return nil
}
