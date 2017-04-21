package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/utils"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Package struct {
	Name string

	Version      string `json:"version"`
	Source       string `json:"source"`
	Hash         string `json:"hash"`
	PreventStart bool   `json:"preventStart"`
}

func (p *Package) SortKey() (int, string) {
	order := TaskOrderPackage
	// Install standard packages first, so that if we want to install a bare deb,
	// that happens after we install apt-get supporting libs
	if p.Source != "" {
		order++
	}
	return order, p.Name
}

func (p *Package) String() string {
	return fmt.Sprintf("Package: %s", p.Name)
}

func NewPackage(name string, contents string) (*Package, error) {
	p := &Package{Name: name}
	if contents != "" {
		err := json.Unmarshal([]byte(contents), p)
		if err != nil {
			return nil, fmt.Errorf("error parsing json for package %q: %v", name, err)
		}
	}
	return p, nil
}

const stateKey = "packages"

type packageState struct {
	installed map[string]bool
}

func (s *packageState) isInstalled(name string) bool {
	return s.installed[name]
}

func (s *packageState) markInstalled(name string) {
	s.installed[name] = true
}

func findInstalledPackages() (*packageState, error) {
	glog.V(2).Infof("Listing installed packages")
	cmd := exec.Command("dpkg-query", "-f", "${db:Status-Abbrev}${binary:Package}\\n", "-W")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing installed packages: %v: %s", err, string(output))
	}
	installed := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		tokens := strings.Split(line, " ")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("error parsing dpkg-query line %q", line)
		}
		state := tokens[0]
		name := tokens[1]

		// If the package has an arch suffix (e.g. libapparmor1:amd64), ignore it
		colonIndex := strings.Index(name, ":")
		if colonIndex != -1 {
			name = name[:colonIndex]
		}

		switch state {
		case "ii":
			installed[name] = true
		case "rc":
			installed[name] = false
		default:
			glog.Warningf("unknown package state %q in line %q", state, line)
		}
	}
	return &packageState{installed: installed}, nil
}

func (p *Package) Run(c *execution.Context) error {
	state, err := c.GetState(stateKey, func() (interface{}, error) {
		return findInstalledPackages()
	})
	if err != nil {
		return err
	}
	packageState := state.(*packageState)
	if packageState.isInstalled(p.Name) {
		glog.V(2).Infof("Package already installed: %q", p.Name)
		return nil
	}

	glog.Infof("Installing package %q", p.Name)

	if p.Source != "" {
		// Install a deb
		local := path.Join("/var/fastdeploy/packages/" + p.Name)
		err := os.MkdirAll(path.Dir(local), 0755)
		if err != nil {
			return fmt.Errorf("error creating directories %q: %v", path.Dir(local), err)
		}

		_, err = utils.DownloadURL(p.Source, local, p.Hash)
		if err != nil {
			return err
		}

		args := []string{"dpkg", "-i", local}
		glog.Infof("running command %s", args)
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error installing package %q: %v: %s", p.Name, err, string(output))
		}
	} else {
		args := []string{"apt-get", "install", "--yes", p.Name}
		glog.Infof("running command %s", args)
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error installing package %q: %v: %s", p.Name, err, string(output))
		}
	}

	packageState.markInstalled(p.Name)

	return nil
}
