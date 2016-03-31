package tasks

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/utils"
	"os/exec"
	"path"
	"strings"
)

type Sysctl struct {
	Name   string
	Config string
}

func (s *Sysctl) SortKey() (int, string) {
	return TaskOrderSysctl, s.Name
}

func (s *Sysctl) String() string {
	return fmt.Sprintf("Sysctl: %s", s.Name)
}

func NewSysctl(name string, contents string) (*Sysctl, error) {
	s := &Sysctl{Name: name}
	s.Config = contents

	_, err := s.parseSettings()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sysctl) parseSettings() (map[string]string, error) {
	settings := make(map[string]string)
	for _, line := range strings.Split(s.Config, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		tokens := strings.Split(line, "=")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("cannot parse sysctl line: %q", line)
		}

		settings[tokens[0]] = tokens[1]
	}
	return settings, nil
}

func (s *Sysctl) Run(c *execution.Context) error {
	settings, err := s.parseSettings()
	if err != nil {
		return err
	}

	dirty := false
	for k, v := range settings {
		glog.V(2).Infof("Reading sysctl setting for %q", k)
		cmd := exec.Command("sysctl", "-n", k)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error reading sysctl for  %q: %v: %s", k, err, string(output))
		}
		existingValue := strings.TrimSpace(string(output))

		if existingValue == v {
			glog.V(2).Infof("sysctl value already in place: %s=%s", k, v)
		} else {
			glog.V(2).Infof("changing sysctl value: %s=%s to =%s", k, existingValue, v)
			dirty = true
		}
	}

	f := path.Join("/etc/sysctl.d", "99-"+s.Name+".conf")
	_, err = utils.WriteFile(f, []byte(s.Config), 0644, 0755)
	if err != nil {
		return fmt.Errorf("error writing sysctl file %q: %v", f, err)
	}

	if dirty {
		glog.V(2).Infof("TODO: Only reload sysctl once?")

		glog.Infof("Applying sysctl settings")
		cmd := exec.Command("sysctl", "--system")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error reloading new sysctl settings: %v\nOutput: %s", err, output)
		}
	}
	return nil
}
