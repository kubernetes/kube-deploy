package machinesetup

import (
	"io"
	clustercommon "k8s.io/kube-deploy/cluster-api/pkg/apis/cluster/common"
	clusterv1 "k8s.io/kube-deploy/cluster-api/pkg/apis/cluster/v1alpha1"
	"reflect"
	"strings"
	"testing"
)

func TestParseMachineSetupYaml(t *testing.T) {
	testTables := []struct {
		reader      io.Reader
		expectedErr bool
	}{
		{
			reader: strings.NewReader(`items:
- machineParams:
  - os: ubuntu-1710
    roles:
    - Master
    versions:
      kubelet: 1.9.3
      controlPlane: 1.9.3
      containerRuntime:
        name: docker
        version: 1.12.0
  - os: ubuntu-1710
    roles:
    - Master
    versions:
      kubelet: 1.9.4
      controlPlane: 1.9.4
      containerRuntime:
        name: docker
        version: 1.12.0
  image: projects/ubuntu-os-cloud/global/images/family/ubuntu-1710
  metadata:
    startupScript: |
      #!/bin/bash
- machineParams:
  - os: ubuntu-1710
    roles:
    - Node
    versions:
      kubelet: 1.9.3
      containerRuntime:
        name: docker
        version: 1.12.0
  - os: ubuntu-1710
    roles:
    - Node
    versions:
      kubelet: 1.9.4
      containerRuntime:
        name: docker
        version: 1.12.0
  image: projects/ubuntu-os-cloud/global/images/family/ubuntu-1710
  metadata:
    startupScript: |
      #!/bin/bash
      echo this is the node config.`),
			expectedErr: false,
		},
		{
			reader:      strings.NewReader("Not valid yaml"),
			expectedErr: true,
		},
	}

	for _, table := range testTables {
		validConfigs, err := parseMachineSetupYaml(table.reader)
		if table.expectedErr {
			if err == nil {
				t.Errorf("An error was not received as expected.")
			}
			if validConfigs != nil {
				t.Errorf("ValidConfigs should be nil, got %v", validConfigs)
			}
		}
		if !table.expectedErr {
			if err != nil {
				t.Errorf("Got unexpected error: %s", err)
			}
			if validConfigs == nil {
				t.Errorf("ValidConfigs should have been parsed, but was nil")
			}
		}
	}
}

func TestGetYaml(t *testing.T) {
	testTables := []struct {
		validConfigs    ValidConfigs
		expectedStrings []string
		expectedErr     bool
	}{
		{
			validConfigs: ValidConfigs{
				configList: &configList{
					Items: []config{
						{
							Params: []ConfigParams{
								{
									OS:    "ubuntu-1710",
									Roles: []clustercommon.MachineRole{clustercommon.MasterRole},
									Versions: clusterv1.MachineVersionInfo{
										Kubelet:      "1.9.4",
										ControlPlane: "1.9.4",
										ContainerRuntime: clusterv1.ContainerRuntimeInfo{
											Name:    "docker",
											Version: "1.12.0",
										},
									},
								},
							},
							Image: "projects/ubuntu-os-cloud/global/images/family/ubuntu-1710",
							Metadata: Metadata{
								StartupScript: "Master startup script",
							},
						},
						{
							Params: []ConfigParams{
								{
									OS:    "ubuntu-1710",
									Roles: []clustercommon.MachineRole{clustercommon.NodeRole},
									Versions: clusterv1.MachineVersionInfo{
										Kubelet: "1.9.4",
										ContainerRuntime: clusterv1.ContainerRuntimeInfo{
											Name:    "docker",
											Version: "1.12.0",
										},
									},
								},
							},
							Image: "projects/ubuntu-os-cloud/global/images/family/ubuntu-1710",
							Metadata: Metadata{
								StartupScript: "Node startup script",
							},
						},
					},
				},
			},
			expectedStrings: []string{"startupScript: Master startup script", "startupScript: Node startup script"},
			expectedErr:     false,
		},
	}

	for _, table := range testTables {
		yaml, err := table.validConfigs.GetYaml()
		if table.expectedErr && err == nil {
			t.Errorf("An error was not received as expected.")
		}
		if !table.expectedErr && err != nil {
			t.Errorf("Got unexpected error: %s", err)
		}
		for _, expectedString := range table.expectedStrings {
			if !strings.Contains(yaml, expectedString) {
				t.Errorf("Yaml did not contain expected string, got:\n%s\nwant:\n%s", yaml, expectedString)
			}
		}
	}
}

func TestMatchMachineSetupConfig(t *testing.T) {
	masterMachineSetupConfig := config{
		Params: []ConfigParams{
			{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.MasterRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet:      "1.9.3",
					ControlPlane: "1.9.3",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.12.0",
					},
				},
			},
			{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.MasterRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet:      "1.9.4",
					ControlPlane: "1.9.4",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.12.0",
					},
				},
			},
		},
		Image: "projects/ubuntu-os-cloud/global/images/family/ubuntu-1710",
		Metadata: Metadata{
			StartupScript: "Master startup script",
		},
	}
	nodeMachineSetupConfig := config{
		Params: []ConfigParams{
			{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.NodeRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet: "1.9.3",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.12.0",
					},
				},
			},
			{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.NodeRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet: "1.9.4",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.12.0",
					},
				},
			},
		},
		Image: "projects/ubuntu-os-cloud/global/images/family/ubuntu-1710",
		Metadata: Metadata{
			StartupScript: "Node startup script",
		},
	}

	validConfigs := ValidConfigs{
		configList: &configList{
			Items: []config{masterMachineSetupConfig, nodeMachineSetupConfig},
		},
	}

	testTables := []struct {
		params        ConfigParams
		expectedMatch *config
		expectedErr   bool
	}{
		{
			params: ConfigParams{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.MasterRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet:      "1.9.4",
					ControlPlane: "1.9.4",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.12.0",
					},
				},
			},
			expectedMatch: &masterMachineSetupConfig,
			expectedErr:   false,
		},
		{
			params: ConfigParams{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.NodeRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet: "1.9.4",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.12.0",
					},
				},
			},
			expectedMatch: &nodeMachineSetupConfig,
			expectedErr:   false,
		},
		{
			params: ConfigParams{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.NodeRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet:      "1.9.4",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.13.0",
					},
				},
			},
			expectedMatch: nil,
			expectedErr:   true,
		},
		{
			params: ConfigParams{
				OS:    "ubuntu-1710",
				Roles: []clustercommon.MachineRole{clustercommon.MasterRole, clustercommon.NodeRole},
				Versions: clusterv1.MachineVersionInfo{
					Kubelet: "1.9.3",
					ContainerRuntime: clusterv1.ContainerRuntimeInfo{
						Name:    "docker",
						Version: "1.12.0",
					},
				},
			},
			expectedMatch: nil,
			expectedErr:   true,
		},
	}

	for _, table := range testTables {
		matched, err := validConfigs.matchMachineSetupConfig(&table.params)
		if !reflect.DeepEqual(matched, table.expectedMatch) {
			t.Errorf("Matched machine setup config was incorrect, got: %+v,\n want %+v.", matched, table.expectedMatch)
		}
		if table.expectedErr && err == nil {
			t.Errorf("An error was not received as expected.")
		}
		if !table.expectedErr && err != nil {
			t.Errorf("Got unexpected error: %s", err)
		}
	}
}
