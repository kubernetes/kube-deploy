#!/bin/bash

# Copyright 2015 The Kubernetes Authors All rights reserved.
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

kube::multinode::main

case $1 in
  bootstrap)
    kube::multinode::check_params
    
    kube::multinode::detect_lsb
    
    kube::multinode::turndown
    
    # If user wants to pull images from registry, do it
    if [[ ! -z "$2" ]]; then
      kube::multinode::pull_from_registry_and_retag "$2"
    fi
    
    kube::multinode::bootstrap_daemon
    
    kube::multinode::offline_bootstrap
    
    kube::multinode::turndown
  ;;
  list)
    kube::multinode::list_images "$2" "${2:-${REG_PREFIX}}/"
  ;;
  pull)
    kube::multinode::pull_images
  ;;
  push)
    if [[ -z "$2" ]]; then
      kube::log::error "Must specify registry_prefix for 'push'"
      exit -1
    else
      kube::multinode::push "$2"
    fi
  ;;
  *)
    kube::log::error "Usage: $0 bootstrap|list|pull|push"
    exit -1
  ;;
esac
