package imagebuilder

import (
	"fmt"
)

// chooseAWSImage stores a list of
func chooseAWSImage(arch, region string) (string, error) {
	amis := make(map[string]map[string]string)
	knownArchs := []string{"arm64", "amd64"}
	for _, ka := range knownArchs {
		amis[ka] = make(map[string]string)
	}
	// Debian 9.8 images from https://wiki.debian.org/Cloud/AmazonEC2Image/Stretch
	amis["amd64"]["ap-northeast-1"] = "ami-0c4290d7ce45d7bbe"
	amis["amd64"]["ap-northeast-2"] = "ami-0fa1392d5d545f9e8"
	amis["amd64"]["ap-south-1"] = "ami-0b6490868957ce747"
	amis["amd64"]["ap-southeast-1"] = "ami-04c9740a9ed018dba"
	amis["amd64"]["ap-southeast-2"] = "ami-0b91189c4f9f5cd9e"
	amis["amd64"]["ca-central-1"] = "ami-0857efbad274a1a89"
	amis["amd64"]["eu-central-1"] = "ami-05449f21272b4ee56"
	amis["amd64"]["eu-north-1"] = "ami-043a919b6dc7c51cc"
	amis["amd64"]["eu-west-1"] = "ami-035c67e6a9ef8f024"
	amis["amd64"]["eu-west-2"] = "ami-0ef10a4062f24d89d"
	amis["amd64"]["eu-west-3"] = "ami-0cb185e7696ffe300"
	amis["amd64"]["sa-east-1"] = "ami-0bc0ce4ab8b82305c"
	amis["amd64"]["us-east-1"] = "ami-0f9e7e8867f55fd8e"
	amis["amd64"]["us-east-2"] = "ami-00c5940f2b52c5d98"
	amis["amd64"]["us-west-1"] = "ami-0afda78f1d0272d99"
	amis["amd64"]["us-west-2"] = "ami-01d07e14f082b3ba1"
	amis["arm64"]["ap-northeast-1"] = "ami-0fea662cdcd9b9dc9"
	amis["arm64"]["ap-northeast-2"] = "ami-0399c4789441957b4"
	amis["arm64"]["ap-south-1"] = "ami-0bd4d9e505ef0baed"
	amis["arm64"]["ap-southeast-1"] = "ami-03a5c6bce47208f68"
	amis["arm64"]["ap-southeast-2"] = "ami-04a251232bc12248d"
	amis["arm64"]["ca-central-1"] = "ami-02827c87632b288a9"
	amis["arm64"]["eu-central-1"] = "ami-0aeab0ea5ff2b82f8"
	amis["arm64"]["eu-north-1"] = "ami-03c6ccf3b408e6b55"
	amis["arm64"]["eu-west-1"] = "ami-0ef6a89d286837ca7"
	amis["arm64"]["eu-west-2"] = "ami-0edc104967d61153e"
	amis["arm64"]["eu-west-3"] = "ami-0bf9401479fab5f4b"
	amis["arm64"]["sa-east-1"] = "ami-0c08b611b50b95e55"
	amis["arm64"]["us-east-1"] = "ami-0e890f2e1ecc27745"
	amis["arm64"]["us-east-2"] = "ami-0b9d7068010cf08c9"
	amis["arm64"]["us-west-1"] = "ami-010e523aff65dce8a"
	amis["arm64"]["us-west-2"] = "ami-046aead1494919fc1"
	// A slightly older image, but the newest one we have
	// FIXME: indicate provenance here?
	amis["amd64"]["cn-north-1"] = "ami-da69a1b7"
	if archMap, archOK := amis[arch]; archOK {
		// arch is OK
		if ami, amiOK := archMap[region]; amiOK {
			return ami, nil
		}
		return "", fmt.Errorf("unknown region specified: %s", region)
	}
	return "", fmt.Errorf("unknown architecture specified: %s", arch)
}

// chooseAWSInstanceType handles the various quirks of instance type selection
// An alternate approach might be to use the AWS Pricing API
func chooseAWSInstanceType(arch, region string) string {
	if region == "us-east-2" && arch == "amd64" {
		// no m3.medium here
		return "m4.large"
	}
	if arch == "arm64" {
		// not available everywhere yet but rather than hardcode a list of
		// functional regions, just let the instance launch fail
		return "a1.medium"
	}
	return "m3.medium"
}
