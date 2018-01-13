package cloud

import (
	machinesv1alpha1 "k8s.io/kube-deploy/machine-api-generic-worker/pkg/machines/v1alpha1"
)

type ConfigProvider interface {
	GetCloudConfig(spec machinesv1alpha1.MachineSpec) (config string, name string, err error)
}
