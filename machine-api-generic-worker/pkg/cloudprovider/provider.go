package cloudprovider

import (
	"errors"

	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/cloudprovider/cloud"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/cloudprovider/provider/aws"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/cloudprovider/provider/digitalocean"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/cloudprovider/provider/openstack"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/providerconfig"
)

var (
	// ErrProviderNotFound tells that the requested cloud provider was not found
	ErrProviderNotFound = errors.New("cloudprovider not found")

	providers = map[providerconfig.CloudProvider]cloud.Provider{
		providerconfig.CloudProviderDigitalocean: digitalocean.New(),
		providerconfig.CloudProviderAWS:          aws.New(),
		providerconfig.CloudProviderOpenstack:    openstack.New(),
	}
)

// ForProvider returns a CloudProvider actuator for the requested provider
func ForProvider(p providerconfig.CloudProvider) (cloud.Provider, error) {
	if p, found := providers[p]; found {
		return p, nil
	}
	return nil, ErrProviderNotFound
}
