package userdata

import (
	"errors"

	machinesv1alpha1 "k8s.io/kube-deploy/machine-api-generic-worker/pkg/machines/v1alpha1"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/providerconfig"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/userdata/cloud"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/userdata/coreos"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/userdata/ubuntu"
)

var (
	ErrProviderNotFound = errors.New("no user data provider for the given os found")

	providers = map[providerconfig.OperatingSystem]Provider{
		providerconfig.OperatingSystemCoreos: coreos.Provider{},
		providerconfig.OperatingSystemUbuntu: ubuntu.Provider{},
	}
)

func ForOS(os providerconfig.OperatingSystem) (Provider, error) {
	if p, found := providers[os]; found {
		return p, nil
	}
	return nil, ErrProviderNotFound
}

type Provider interface {
	UserData(spec machinesv1alpha1.MachineSpec, kubeconfig string, ccProvider cloud.ConfigProvider) (string, error)
}
