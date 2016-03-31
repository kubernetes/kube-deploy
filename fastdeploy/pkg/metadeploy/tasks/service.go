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
	"time"
)

type Service struct {
	Name     string
	contents string

	NoStart bool `json:"noStart"`
}

func (s *Service) SortKey() (int, string) {
	return TaskOrderService, s.Name
}

func (s *Service) String() string {
	return fmt.Sprintf("Service: %s", s.Name)
}

func NewService(name string, contents string, meta string) (*Service, error) {
	s := &Service{Name: name}
	s.contents = contents

	if meta != "" {
		err := json.Unmarshal([]byte(meta), s)
		if err != nil {
			return nil, fmt.Errorf("error parsing json for service %q: %v", name, err)
		}
	}

	return s, nil
}

func (s *Service) Run(c *execution.Context) error {
	systemdSystemPath := "/lib/systemd/system" // TODO: Different on redhat
	servicePath := path.Join(systemdSystemPath, s.Name)
	changed, err := utils.WriteFile(servicePath, []byte(s.contents), 0644, 0755)
	if err != nil {
		return fmt.Errorf("error writing systemd service file: %v", err)
	}

	if changed {
		glog.Infof("Reloading systemd configuration")
		cmd := exec.Command("systemctl", "daemon-reload")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error doing systemd daemon-reload: %v\nOutput: %s", err, output)
		}
	}

	if !s.NoStart {
		shouldRestart := changed
		if !shouldRestart {
			var dependencies []string

			for _, line := range strings.Split(s.contents, "\n") {
				line = strings.TrimSpace(line)
				tokens := strings.SplitN(line, "=", 2)
				if len(tokens) != 2 {
					continue
				}
				k := strings.TrimSpace(tokens[0])
				v := strings.TrimSpace(tokens[1])
				if k == "EnvironmentFile" {
					dependencies = append(dependencies, v)
				}
				// TODO: ExecStart=/usr/local/bin/kubelet "$DAEMON_ARGS"
			}

			var newest time.Time
			for _, dependency := range dependencies {
				stat, err := os.Stat(dependency)
				if err != nil {
					glog.Infof("Ignoring error checking service dependency %q: %v", dependency, err)
					continue
				}
				modTime := stat.ModTime()
				if newest.IsZero() || newest.Before(modTime) {
					newest = modTime
				}
			}
			if !newest.IsZero() {
				serviceName := s.Name
				glog.V(2).Infof("querying state of service %q", serviceName)
				cmd := exec.Command("systemctl", "show", "--all", serviceName)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("error doing systemd restart %s: %v\nOutput: %s", serviceName, err, output)
				}
				properties := make(map[string]string)
				for _, line := range strings.Split(string(output), "\n") {
					tokens := strings.SplitN(line, "=", 2)
					if len(tokens) != 2 {
						continue
					}
					properties[tokens[0]] = tokens[1]
				}

				startedAt := properties["ExecMainStartTimestamp"]
				if startedAt == "" {
					shouldRestart = true
				} else {
					startedAtTime, err := time.Parse("Mon 2006-01-02 15:04:05 MST", startedAt)
					if err != nil {
						return fmt.Errorf("unable to parse service ExecMainStartTimestamp: %q", startedAt)
					}
					if startedAtTime.Before(newest) {
						glog.V(2).Infof("will restart service %q because dependency changed after service start", serviceName)
						shouldRestart = true
					} else {
						glog.V(2).Infof("will not restart service %q - started after dependencies", serviceName)
					}
				}
			}
			glog.Infof("TODO: allow more depenendencies")
		}

		if shouldRestart {
			serviceName := s.Name
			glog.Infof("Restarting service %q", serviceName)
			cmd := exec.Command("systemctl", "restart", serviceName)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error doing systemd restart %s: %v\nOutput: %s", serviceName, err, output)
			}
		}
	}

	return nil
}
