ImageBuilder
============

ImageBuilder is a tool for building an optimized k8s images, currently only supporting AWS.

Please also see the README in templates for documentation as to the motivation for building a custom image.

It is a wrapper around bootstrap-vz (the tool used to build official Debian cloud images).
It adds functionality to spin up an instance for building the image, and publishing the image to all regions.

Imagebuilder create an instance to build the image, builds the image as specified by `TemplatePath`, makes the
image public and copies it to all accessible regions (on AWS), and then shuts down the builder instance.
Each of these stages can be controlled through flags
(for example, you might not want use `--publish=false` for an internal image.)

## Usage

Check out `--help`, but these options control which operations we perform, and may be useful for debugging or publishing a lot of images:

* `-down=true/false` – Set to shut down instance (if found) (default true).
* `-up=true/false` – Set to create instance (if not found) (default true).
* `-publish=true/false` – Set to whether we should make image public or not (default true).
* `-replicate=true/false` – Set to copy the image to all regions (default true).
* `-config=<configpath>` – Which config file to use. See `aws.yml` and `gce.yml` for examples.

## Guide

1. Run `go get k8s.io/kube-deploy/imagebuilder`.
2. Change directory to `${GOPATH}/src/k8s.io/kube-deploy/imagebuilder`.

### AWS

1. Install [Terrafom](https://www.terraform.io/downloads.html).
2. Create a `terraform/aws/terraform.tfvars` file with the following configuration set appropriately:
   ```
   access_key = "foo"
   secret_key = "bar"
   region = "baz"
   ```
3. View insfrastructure creation plan using `make aws_infrastructure_plan`.
4. Implement infrastructure plan using `make aws_infrastructure`.
5. Install imagebuilder with code with `make install`.
6. Run imagebuilder with `imagebuilder --config aws.yaml --v=8`. It will print the IDs of the image in each region, but it will also tag the image with a Name as specified in the template as this is the easier way to retrieve the image.
7. View cleanup plan using  `make aws_cleanup_plan`.
8. Remove created infrastructure with `make aws_cleanup`.

### GCE

1. Install imagebuilder with code with `make install`.
2. Edit `gce.yaml`, at least to specify the `Project` and `GCSDestination` to use.
3. Create the `GCS` bucket in `GCSDestination` (if it does not exist) `gsutil mb gs://<bucketname>/`.
4. Run imagebuilder with `imagebuilder --config gce.yaml --v=8 --publish=false`. Note that because GCE does not currently support publishing images, you must pass `--publish=false`. Also, images on
GCE are global, so `replicate` does not actually need to do anything.
