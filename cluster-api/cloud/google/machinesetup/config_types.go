/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machinesetup

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	clustercommon "k8s.io/kube-deploy/cluster-api/pkg/apis/cluster/common"
	clusterv1 "k8s.io/kube-deploy/cluster-api/pkg/apis/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/util"
)

type ConfigWatch struct {
	path string
}

type ValidConfigs struct {
	configList *configList
}

type configList struct {
	Items []config `json:"items"`
}

type config struct {
	// A list of the valid combinations of ConfigParams that will
	// map to the given Image and Metadata.
	Params []ConfigParams `json:"machineParams"`

	// The fully specified image path.
	Image    string   `json:"image"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	StartupScript string `json:"startupScript"`
}

type ConfigParams struct {
	OS       string
	Roles    []clustercommon.MachineRole
	Versions clusterv1.MachineVersionInfo
}

func NewConfigWatch(path string) (*ConfigWatch, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return &ConfigWatch{path: path}, nil
}

func (cw *ConfigWatch) ValidConfigs() (*ValidConfigs, error) {
	file, err := os.Open(cw.path)
	if err != nil {
		return nil, err
	}
	return parseMachineSetupYaml(file)
}

func parseMachineSetupYaml(reader io.Reader) (*ValidConfigs, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	configList := &configList{}
	err = yaml.Unmarshal(bytes, configList)
	if err != nil {
		return nil, err
	}

	return &ValidConfigs{configList}, nil
}

func (vc *ValidConfigs) GetYaml() (string, error) {
	bytes, err := yaml.Marshal(vc.configList)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (vc *ValidConfigs) GetImage(params *ConfigParams) (string, error) {
	machineSetupConfig, err := vc.matchMachineSetupConfig(params)
	if err != nil {
		return "", err
	}
	return machineSetupConfig.Image, nil
}

func (vc *ValidConfigs) GetMetadata(params *ConfigParams) (Metadata, error) {
	machineSetupConfig, err := vc.matchMachineSetupConfig(params)
	if err != nil {
		return Metadata{}, err
	}
	return machineSetupConfig.Metadata, nil
}

func (vc *ValidConfigs) matchMachineSetupConfig(params *ConfigParams) (*config, error) {
	for _, conf := range vc.configList.Items {
		for _, validParams := range conf.Params {
			if params.OS != validParams.OS {
				continue
			}
			if len(params.Roles) != len(validParams.Roles) {
				continue
			}
			foundRoles := true
			for _, role := range params.Roles {
				if !util.RoleContains(role, validParams.Roles) {
					foundRoles = false
					break
				}
			}
			if !foundRoles {
				continue
			}
			if params.Versions != validParams.Versions {
				continue
			}
			return &conf, nil
		}
	}
	return nil, fmt.Errorf("could not find a matching machine setup config for params %+v", params)
}
