package imagebuilder

import (
	"strings"

	"github.com/golang/glog"
)

type Config struct {
	Cloud         string
	TemplatePath  string
	SetupCommands [][]string

	BootstrapVZRepo   string
	BootstrapVZBranch string

	SSHUsername   string
	SSHPublicKey  string
	SSHPrivateKey string

	InstanceProfile string

	// Tags to add to the image
	Tags map[string]string
}

func (c *Config) InitDefaults() {
	c.BootstrapVZRepo = "https://github.com/andsens/bootstrap-vz.git"
	c.BootstrapVZBranch = "master"

	c.SSHUsername = "admin"
	c.SSHPublicKey = "~/.ssh/id_rsa.pub"
	c.SSHPrivateKey = "~/.ssh/id_rsa"

	c.InstanceProfile = ""

	setupCommands := []string{
		"sudo apt-get update",
		"sudo apt-get install --yes git python debootstrap python-pip kpartx parted",
		"sudo pip install --upgrade requests termcolor jsonschema fysom docopt pyyaml boto boto3 json_minify",
	}
	for _, cmd := range setupCommands {
		c.SetupCommands = append(c.SetupCommands, strings.Split(cmd, " "))
	}
}

type AWSConfig struct {
	Config

	Arch            string
	Region          string
	ImageID         string
	InstanceType    string
	SSHKeyName      string
	SubnetID        string
	SecurityGroupID string
	Tags            map[string]string
}

func (c *AWSConfig) InitDefaults(arch, region string) {
	c.Config.InitDefaults()
	c.InstanceType = "m3.medium"
	if region == "" {
		region = "us-east-1"
	}
	c.Region = region
	if arch == "" {
		arch = "amd64"
	}
	c.Arch = arch
	ami, err := chooseAWSImage(c.Arch, c.Region)
	if err != nil {
		glog.Warningf("No image known for this architecture (%q) and region (%q), please specify them explicitly: %v", c.Arch, c.Region, err)
	} else {
		c.ImageID = ami
	}
	c.InstanceType = chooseAWSInstanceType(c.Arch, c.Region)
}

type GCEConfig struct {
	Config

	// To create an image on GCE, we have to upload it to a bucket first
	GCSDestination string

	Project     string
	Zone        string
	MachineName string

	MachineType string
	Image       string
	Tags        map[string]string
}

func (c *GCEConfig) InitDefaults() {
	c.Config.InitDefaults()
	c.MachineName = "k8s-imagebuilder"
	c.Zone = "us-central1-f"
	c.MachineType = "n1-standard-2"
	c.Image = "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-8-jessie-v20160329"
}
