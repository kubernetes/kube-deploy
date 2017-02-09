#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
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

# Source common.sh
source $(dirname "${BASH_SOURCE}")/common.sh

# Set MASTER_IP to localhost when deploying a master
MASTER_IP=localhost

kube::multinode::main

# check if init script was run and matches the desired K8S version
if [[ -z $(ls /etc/kubernetes/) ]]; then
  echo "Please run init.sh to create manifest files in /etc/kubernetes."
  exit 1
fi
K8S_VERSION_IN_MANIFEST=$(sed -n 's/.*hyperkube.*\(v.*\)\".*,/\1/p' /etc/kubernetes/manifests-multi/master-multi.json | head -1)
if [[ $K8S_VERSION != $K8S_VERSION_IN_MANIFEST ]]; then
  echo "Please run init.sh to create manifest files that match to the Kubernetes version $K8S_VERSION"
  exit 1
fi

kube::multinode::log_variables

kube::multinode::turndown

if [[ ${USE_CNI} == "true" ]]; then
  kube::cni::ensure_docker_settings

  kube::multinode::start_etcd

  kube::multinode::start_flannel
else
  kube::bootstrap::bootstrap_daemon

  kube::multinode::start_etcd

  kube::multinode::start_flannel

  kube::bootstrap::restart_docker
fi

kube::multinode::start_k8s_master

kube::log::status "Done. It may take about a minute before apiserver is up."
