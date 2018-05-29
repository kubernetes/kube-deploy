# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: genapi clean

APIBUILDER_GENERATORS=apiregister conversion openapi
APIBUILDER_IMAGE_TAG=dev-cluster-api-generator

all: genapi clean

# Use apiserver-builder to automatically enerate API code
# This will:
# - Build a Docker image with the appropriate apiserver-builder version
# - Add the current host user as a user in the container so that the code
#   may be bind mounted with the correct permissions
# - Execute the apiserver-builder generation process
# - Clean up the build and tagged dev-cluster-api-generator Docker image
genapi:
	docker build --build-arg DEVUSERNAME=`id -un` --build-arg DEVUSERID=`id -u` -f Dockerfile.apigen . -t ${APIBUILDER_IMAGE_TAG}
	for generator in ${APIBUILDER_GENERATORS}; do \
		docker run --rm --user=`id -un` -v ${PWD}:/go/src/k8s.io/kube-deploy/cluster-api ${APIBUILDER_IMAGE_TAG} /opt/apiserver-builder/bin/apiserver-boot build generated --generator $$generator; \
	done;

clean:
	-docker ps -a -q --filter ancestor=${APIBUILDER_IMAGE_TAG} --format="{{.ID}}" | xargs -r docker stop | xargs -r docker rm
	-docker rmi ${APIBUILDER_IMAGE_TAG}