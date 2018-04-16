package openstack

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"sync"

	"github.com/gophercloud/gophercloud"
	goopenstack "github.com/gophercloud/gophercloud/openstack"
	osavailabilityzones "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/availabilityzones"
	oskeypairs "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	osflavors "github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	osimages "github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	osregions "github.com/gophercloud/gophercloud/openstack/identity/v3/regions"
	osfloatingips "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	ossecuritygroups "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	osecruritygrouprules "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	osnetworks "github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	osports "github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	ossubnets "github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/pagination"
	"golang.org/x/crypto/ssh"
)

var (
	errNotFound = errors.New("not found")
)

// Protects creation of public key
var publicKeyCreationLock = &sync.Mutex{}

func getRegion(client *gophercloud.ProviderClient, id string) (*osregions.Region, error) {
	idClient, err := goopenstack.NewIdentityV3(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic})
	if err != nil {
		return nil, err
	}

	return osregions.Get(idClient, id).Extract()
}

func getAvailabilityZone(client *gophercloud.ProviderClient, region, name string) (*osavailabilityzones.AvailabilityZone, error) {
	computeClient, err := goopenstack.NewComputeV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	allPages, err := osavailabilityzones.List(computeClient).AllPages()
	if err != nil {
		return nil, err
	}
	zones, err := osavailabilityzones.ExtractAvailabilityZones(allPages)
	if err != nil {
		return nil, err
	}

	for _, z := range zones {
		if z.ZoneName == name {
			return &z, nil
		}
	}

	return nil, errNotFound
}

func getImageByName(client *gophercloud.ProviderClient, region, name string) (*osimages.Image, error) {
	computeClient, err := goopenstack.NewComputeV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	var allImages []osimages.Image
	pager := osimages.ListDetail(computeClient, osimages.ListOpts{})
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		images, err := osimages.ExtractImages(page)
		if err != nil {
			return false, err
		}
		allImages = append(allImages, images...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	for _, i := range allImages {
		if i.Name == name {
			return &i, nil
		}
	}

	return nil, errNotFound
}

func getFlavor(client *gophercloud.ProviderClient, region, name string) (*osflavors.Flavor, error) {
	computeClient, err := goopenstack.NewComputeV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	var allFlavors []osflavors.Flavor

	pager := osflavors.ListDetail(computeClient, osflavors.ListOpts{})
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		flavors, err := osflavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}
		allFlavors = append(allFlavors, flavors...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	for _, f := range allFlavors {
		if f.Name == name {
			return &f, nil
		}
	}

	return nil, errNotFound
}

func getSecurityGroup(client *gophercloud.ProviderClient, region, name string) (*ossecuritygroups.SecGroup, error) {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	var allGroups []ossecuritygroups.SecGroup
	pager := ossecuritygroups.List(netClient, ossecuritygroups.ListOpts{})
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		groups, err := ossecuritygroups.ExtractGroups(page)
		if err != nil {
			return false, err
		}
		allGroups = append(allGroups, groups...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	for _, g := range allGroups {
		if g.Name == name {
			return &g, nil
		}
	}

	return nil, errNotFound
}

func getNetwork(client *gophercloud.ProviderClient, region, name string) (*osnetworks.Network, error) {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	var allNetworks []osnetworks.Network
	pager := osnetworks.List(netClient, osnetworks.ListOpts{})
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		networks, err := osnetworks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}
		allNetworks = append(allNetworks, networks...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	for _, n := range allNetworks {
		if n.Name == name {
			return &n, nil
		}
	}

	return nil, errNotFound
}

func getSubnet(client *gophercloud.ProviderClient, region, name string) (*ossubnets.Subnet, error) {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	var allSubnets []ossubnets.Subnet
	pager := ossubnets.List(netClient, ossubnets.ListOpts{})
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		subnets, err := ossubnets.ExtractSubnets(page)
		if err != nil {
			return false, err
		}
		allSubnets = append(allSubnets, subnets...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	for _, s := range allSubnets {
		if s.Name == name {
			return &s, nil
		}
	}

	return nil, errNotFound
}

func ensureSSHKeysExist(client *gophercloud.ProviderClient, region, name string, rsa rsa.PublicKey) (string, error) {
	publicKeyCreationLock.Lock()
	defer publicKeyCreationLock.Unlock()

	computeClient, err := goopenstack.NewComputeV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return "", err
	}

	pk, err := ssh.NewPublicKey(&rsa)
	if err != nil {
		return "", fmt.Errorf("failed to parse publickey: %v", err)
	}

	kp, err := oskeypairs.Get(computeClient, name).Extract()
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); ok {
			kp, err = oskeypairs.Create(computeClient, oskeypairs.CreateOpts{
				Name:      name,
				PublicKey: string(ssh.MarshalAuthorizedKey(pk)),
			}).Extract()
			if err != nil {
				return "", fmt.Errorf("failed to create publickey %q: %v", name, err)
			}

			return kp.Name, nil
		}
		return "", fmt.Errorf("failed to get publickey %q: %v", name, err)
	}

	return kp.Name, nil
}

func ensureKubernetesSecurityGroupExist(client *gophercloud.ProviderClient, region, name string) error {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return err
	}

	_, err = getSecurityGroup(client, region, name)
	if err != nil {
		if err == errNotFound {
			sg, err := ossecuritygroups.Create(netClient, ossecuritygroups.CreateOpts{Name: name}).Extract()
			if err != nil {
				return fmt.Errorf("failed to create security group %s: %v", name, err)
			}

			rules := []osecruritygrouprules.CreateOpts{
				{
					// Allows ipv4 traffic within this group
					Direction:     osecruritygrouprules.DirIngress,
					EtherType:     osecruritygrouprules.EtherType4,
					SecGroupID:    sg.ID,
					RemoteGroupID: sg.ID,
				},
				{
					// Allows ipv6 traffic within this group
					Direction:     osecruritygrouprules.DirIngress,
					EtherType:     osecruritygrouprules.EtherType6,
					SecGroupID:    sg.ID,
					RemoteGroupID: sg.ID,
				},
			}

			for _, opts := range rules {
				if _, err := osecruritygrouprules.Create(netClient, opts).Extract(); err != nil {
					return fmt.Errorf("failed to create security group rule: %v", err)
				}
			}
		}
	}

	return nil
}

func getFreeFloatingIPs(client *gophercloud.ProviderClient, region string, floatingIPPool *osnetworks.Network) ([]osfloatingips.FloatingIP, error) {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	allPages, err := osfloatingips.List(netClient, osfloatingips.ListOpts{FloatingNetworkID: floatingIPPool.ID}).AllPages()
	if err != nil {
		return nil, err
	}

	allFIPs, err := osfloatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return nil, err
	}

	var freeFIPs []osfloatingips.FloatingIP
	for _, f := range allFIPs {
		if f.Status == "ACTIVE" && f.PortID == "" {
			freeFIPs = append(freeFIPs, f)
		}
	}

	return freeFIPs, nil
}

func assignFloatingIP(client *gophercloud.ProviderClient, region, ipID, instanceID, networkID string) error {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return err
	}

	port, err := getInstancePort(client, region, instanceID, networkID)
	if err != nil {
		return fmt.Errorf("failed to get instance port: %v", err)
	}
	_, err = osfloatingips.Update(netClient, ipID, osfloatingips.UpdateOpts{
		PortID: &port.ID,
	}).Extract()

	return err
}

func createFloatingIP(client *gophercloud.ProviderClient, region string, floatingIPPool *osnetworks.Network) (*osfloatingips.FloatingIP, error) {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	return osfloatingips.Create(netClient, osfloatingips.CreateOpts{FloatingNetworkID: floatingIPPool.ID}).Extract()
}

func getInstancePort(client *gophercloud.ProviderClient, region, instanceID, networkID string) (*osports.Port, error) {
	netClient, err := goopenstack.NewNetworkV2(client, gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic, Region: region})
	if err != nil {
		return nil, err
	}

	allPages, err := osports.List(netClient, osports.ListOpts{}).AllPages()
	if err != nil {
		return nil, err
	}

	allPorts, err := osports.ExtractPorts(allPages)
	if err != nil {
		return nil, err
	}

	for _, p := range allPorts {
		if p.NetworkID == networkID && p.DeviceID == instanceID {
			return &p, nil
		}
	}

	return nil, errNotFound
}
