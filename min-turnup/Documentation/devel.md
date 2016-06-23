# Developing on min-turnup

### How to deploy a cluster

Here is a flow for deploying a min-turnup cluster on GCE.
```
$ # checkout code and enter min-turnup directory
$ cd min-turnup/

$ # build .config.json
$ make config
$ # or
$ make menuconfig
$ make .config.json
$ # mine looks like
$ cat .config.json
{
  "phase1": {
    "num_nodes": 4,
    "instance_prefix": "kuberentes",
    "cloud_provider": "gce",
    "gce": {
      "os_image": "ubuntu-1604-xenial-v20160420c",
      "instance_type": "n1-standard-2",
      "project": "",
      "region": "us-central1",
      "zone": "us-central1-b",
      "network": ""
    }
  },
  "phase1b": {
    "extra_api_sans": "",
    "extra_api_dns_names": "kubernetes-master"
  },
  "phase2": {
    "docker_registry": "gcr.io/google-containers",
    "kubernetes_version": "v1.2.4"
  },
  "phase3": {
    "run_addons": true,
    "kube_proxy": true,
    "dashboard": true,
    "kube_dns": true,
    "elasticsearch": true
  }
}

$ # generate and deploy terraform
$ cd phase1/gce/
$ ./gen
$ terraform apply out/
```

The current dependencies are:

* terraform on your path, available [here](https://www.terraform.io/downloads.html).
* jsonnet which needs to be built from source, available [here](https://github.com/google/jsonnet/releases/tag/v0.8.8).
* linux x86_64, other platforms are not yet tested.
* jq
