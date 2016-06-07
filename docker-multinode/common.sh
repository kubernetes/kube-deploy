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

# Utility functions for Kubernetes in docker setup

kube::multinode::main(){
  LATEST_STABLE_K8S_VERSION=$(kube::helpers::curl "https://storage.googleapis.com/kubernetes-release/release/stable.txt")
  K8S_VERSION=${K8S_VERSION:-${LATEST_STABLE_K8S_VERSION}}

  ETCD_VERSION=${ETCD_VERSION:-"2.2.5"}

  DNS_DOMAIN=${DNS_DOMAIN:-"cluster.local"}
  DNS_SERVER_IP=${DNS_SERVER_IP:-"10.0.0.10"}

  RESTART_POLICY=${RESTART_POLICY:-"on-failure"}

  CURRENT_PLATFORM=$(kube::helpers::host_platform)
  ARCH=${ARCH:-${CURRENT_PLATFORM##*/}}

  NET_INTERFACE=${NET_INTERFACE:-$(ip -o -4 route show to default | awk '{print $5}')}

  # Constants
  TIMEOUT_FOR_SERVICES=20
  BOOTSTRAP_DOCKER_SOCK="unix:///var/run/docker-bootstrap.sock"
  KUBELET_MOUNTS="\
    -v /sys:/sys:ro \
    -v /var/run:/var/run:rw \
    -v /var/lib/docker:/var/lib/docker:rw \
    -v /var/lib/kubelet:/var/lib/kubelet:shared \
    -v /var/log/containers:/var/log/containers:rw"

  # Paths
  FLANNEL_SUBNET_TMPDIR=$(mktemp -d)

  # Trap errors
  kube::log::install_errexit
}

# Make shared kubelet directory
kube::multinode::make_shared_kubelet_dir() {
    mkdir -p /var/lib/kubelet
    mount --bind /var/lib/kubelet /var/lib/kubelet
    mount --make-shared /var/lib/kubelet
}

# Ensure everything is OK, docker is running and we're root
kube::multinode::check_params() {

  # Make sure docker daemon is running
  if [[ $(docker ps 2>&1 1>/dev/null; echo $?) != 0 ]]; then
    kube::log::error "Docker is not running on this machine!"
    exit 1
  fi

  # Require root
  if [[ "$(id -u)" != "0" ]]; then
    kube::log::error >&2 "Please run as root"
    exit 1
  fi

  # Output the value of the variables
  kube::log::status "K8S_VERSION is set to: ${K8S_VERSION}"
  kube::log::status "ETCD_VERSION is set to: ${ETCD_VERSION}"
  kube::log::status "DNS_DOMAIN is set to: ${DNS_DOMAIN}"
  kube::log::status "DNS_SERVER_IP is set to: ${DNS_SERVER_IP}"
  kube::log::status "RESTART_POLICY is set to: ${RESTART_POLICY}"
  kube::log::status "MASTER_IP is set to: ${MASTER_IP}"
  kube::log::status "ARCH is set to: ${ARCH}"
  kube::log::status "NET_INTERFACE is set to: ${NET_INTERFACE}"
  kube::log::status "--------------------------------------------"
}

# Detect the OS distro, we support ubuntu, debian, mint, centos, fedora and systemd dist
kube::multinode::detect_lsb() {

  if kube::helpers::command_exists lsb_release; then
    lsb_dist="$(lsb_release -si)"
  elif [[ -r /etc/lsb-release ]]; then
    lsb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
  elif [[ -r /etc/debian_version ]]; then
    lsb_dist='debian'
  elif [[ -r /etc/fedora-release ]]; then
    lsb_dist='fedora'
  elif [[ -r /etc/os-release ]]; then
    lsb_dist="$(. /etc/os-release && echo "$ID")"
  elif kube::helpers::command_exists systemctl; then
    lsb_dist='systemd'
  fi

  lsb_dist="$(echo ${lsb_dist} | tr '[:upper:]' '[:lower:]')"

  case "${lsb_dist}" in
      amzn|centos|debian|ubuntu|systemd)
          ;;
      *)
          kube::log::error "Error: We currently only support ubuntu|debian|amzn|centos|systemd."
          exit 1
          ;;
  esac

  kube::log::status "Detected OS: ${lsb_dist}"
}

# Start etcd on the master node
kube::multinode::start_etcd() {

  kube::log::status "Launching etcd..."
  
  docker run -d \
    --restart=${RESTART_POLICY} \
    --net=host \
    gcr.io/google_containers/etcd-${ARCH}:${ETCD_VERSION} \
    /usr/local/bin/etcd \
      --listen-client-urls=http://127.0.0.1:4001,http://${MASTER_IP}:4001 \
      --advertise-client-urls=http://${MASTER_IP}:4001 \
      --data-dir=/var/etcd/data

  # Wait for etcd to come up
  sleep 5
}

# Start kubelet first and then the master components as pods
kube::multinode::start_k8s_master() {
  
  kube::log::status "Launching Kubernetes master components..."

  kube::multinode::make_shared_kubelet_dir

  # TODO: Get rid of --hostname-override
  docker run -d \
    --net=host \
    --pid=host \
    --privileged \
    --restart=${RESTART_POLICY} \
    ${KUBELET_MOUNTS} \
    gcr.io/google_containers/hyperkube-${ARCH}:${K8S_VERSION} \
    /hyperkube kubelet \
      --allow-privileged \
      --api-servers=http://localhost:8080 \
      --config=/etc/kubernetes/manifests-multi \
      --cluster-dns=${DNS_SERVER_IP} \
      --cluster-domain=${DNS_DOMAIN} \
      --hostname-override=$(ip -o -4 addr list ${NET_INTERFACE} | awk '{print $4}' | cut -d/ -f1) \
      --v=2 --experimental-flannel-overlay=true
}

# Start kubelet in a container, for a worker node
kube::multinode::start_k8s_worker() {
  
  kube::log::status "Launching Kubernetes worker components..."

  kube::multinode::make_shared_kubelet_dir

  # TODO: Use secure port for communication
  # TODO: Get rid of --hostname-override
  docker run -d \
    --net=host \
    --pid=host \
    --privileged \
    --restart=${RESTART_POLICY} \
    ${KUBELET_MOUNTS} \
    gcr.io/google_containers/hyperkube-${ARCH}:${K8S_VERSION} \
    /hyperkube kubelet \
      --allow-privileged=true \
      --api-servers=http://${MASTER_IP}:8080 \
      --cluster-dns=${DNS_SERVER_IP} \
      --cluster-domain=${DNS_DOMAIN} \
      --hostname-override=$(ip -o -4 addr list ${NET_INTERFACE} | awk '{print $4}' | cut -d/ -f1) \
      --v=2
}

# Start kube-proxy in a container, for a worker node
kube::multinode::start_k8s_worker_proxy() {

  kube::log::status "Launching kube-proxy..."

  # TODO: Run kube-proxy in a DaemonSet
  docker run -d \
    --net=host \
    --privileged \
    --restart=${RESTART_POLICY} \
    gcr.io/google_containers/hyperkube-${ARCH}:${K8S_VERSION} \
    /hyperkube proxy \
        --master=http://${MASTER_IP}:8080 \
        --v=2
}

# Turndown the local cluster
kube::multinode::turndown(){

  if [[ $(kube::helpers::is_running /hyperkube) == "true" ]]; then
    
    kube::log::status "Killing hyperkube containers..."

    # Kill all hyperkube docker images
    docker rm -f $(docker ps | grep gcr.io/google_containers/hyperkube | awk '{print $1}')
  fi

  if [[ $(kube::helpers::is_running /pause) == "true" ]]; then
    
    kube::log::status "Killing pause containers..."

    # Kill all pause docker images
    docker rm -f $(docker ps | grep gcr.io/google_containers/pause | awk '{print $1}')
  fi

  if [[ -d /var/lib/kubelet ]]; then
    read -p "Do you want to clean /var/lib/kubelet? [Y/n] " clean_kubelet_dir

    case $clean_kubelet_dir in
      [n|N]*)
        ;; # Do nothing
      *)
        # umount if there's mounts bound in /var/lib/kubelet
        if [[ ! -z $(mount | grep /var/lib/kubelet | awk '{print $3}') ]]; then
          umount $(mount | grep /var/lib/kubelet | awk '{print $3}')
        fi

        # Delete the directory
        rm -rf /var/lib/kubelet
    esac
  fi
}


## Helpers

# Check if a command is valid
kube::helpers::command_exists() {
    command -v "$@" > /dev/null 2>&1
}

# Usage: kube::helpers::file_replace_line {path_to_file} {value_to_search_for} {replace_that_line_with_this_content}
# Finds a line in a file and replaces the line with the third argument
kube::helpers::file_replace_line(){
  if [[ -z $(grep "$2" $1) ]]; then
    echo "$3" >> $1
  else
    sed -i "/$2/c\\$3" $1
  fi
}

# Check if a process is running
kube::helpers::is_running(){
  if [[ ! -z $(ps aux | grep ${1} | grep -v grep) ]]; then
    echo "true"
  else
    echo "false"
  fi
}

# Wraps curl or wget in a helper function.
# Output is redirected to stdout
kube::helpers::curl(){
  if [[ $(which curl 2>&1) ]]; then
    curl -sSL $1
  elif [[ $(which wget 2>&1) ]]; then
    wget -qO- $1
  else
    kube::log::error "Couldn't find curl or wget. Bailing out."
    exit 4
  fi
}

# This figures out the host platform without relying on golang. We need this as
# we don't want a golang install to be a prerequisite to building yet we need
# this info to figure out where the final binaries are placed.
kube::helpers::host_platform() {
  local host_os
  local host_arch
  case "$(uname -s)" in
    Linux)
      host_os=linux;;
    *)
      kube::log::error "Unsupported host OS. Must be linux."
      exit 1;;
  esac

  case "$(uname -m)" in
    x86_64*)
      host_arch=amd64;;
    i?86_64*)
      host_arch=amd64;;
    amd64*)
      host_arch=amd64;;
    aarch64*)
      host_arch=arm64;;
    arm64*)
      host_arch=arm64;;
    arm*)
      host_arch=arm;;
    ppc64le*)
      host_arch=ppc64le;;
    *)  
      kube::log::error "Unsupported host arch. Must be x86_64, arm, arm64 or ppc64le."
      exit 1;;
  esac
  echo "${host_os}/${host_arch}"
}

# Print a status line. Formatted to show up in a stream of output.
kube::log::status() {
  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "+++ $timestamp $1"
  shift
  for message; do
    echo "    $message"
  done
}

# Handler for when we exit automatically on an error.
# Borrowed from https://gist.github.com/ahendrix/7030300
kube::log::errexit() {
  local err="${PIPESTATUS[@]}"

  # If the shell we are in doesn't have errexit set (common in subshells) then
  # don't dump stacks.
  set +o | grep -qe "-o errexit" || return

  set +o xtrace
  local code="${1:-1}"
  kube::log::error_exit "'${BASH_COMMAND}' exited with status $err" "${1:-1}" 1
}

kube::log::install_errexit() {
  # trap ERR to provide an error handler whenever a command exits nonzero  this
  # is a more verbose version of set -o errexit
  trap 'kube::log::errexit' ERR

  # setting errtrace allows our ERR trap handler to be propagated to functions,
  # expansions and subshells
  set -o errtrace
}

# Print out the stack trace
#
# Args:
#   $1 The number of stack frames to skip when printing.
kube::log::stack() {
  local stack_skip=${1:-0}
  stack_skip=$((stack_skip + 1))
  if [[ ${#FUNCNAME[@]} -gt $stack_skip ]]; then
    echo "Call stack:" >&2
    local i
    for ((i=1 ; i <= ${#FUNCNAME[@]} - $stack_skip ; i++))
    do
      local frame_no=$((i - 1 + stack_skip))
      local source_file=${BASH_SOURCE[$frame_no]}
      local source_lineno=${BASH_LINENO[$((frame_no - 1))]}
      local funcname=${FUNCNAME[$frame_no]}
      echo "  $i: ${source_file}:${source_lineno} ${funcname}(...)" >&2
    done
  fi
}

# Log an error and exit.
# Args:
#   $1 Message to log with the error
#   $2 The error code to return
#   $3 The number of stack frames to skip when printing.
kube::log::error_exit() {
  local message="${1:-}"
  local code="${2:-1}"
  local stack_skip="${3:-0}"
  stack_skip=$((stack_skip + 1))

  local source_file=${BASH_SOURCE[$stack_skip]}
  local source_line=${BASH_LINENO[$((stack_skip - 1))]}
  echo "!!! Error in ${source_file}:${source_line}" >&2
  [[ -z ${1-} ]] || {
    echo "  ${1}" >&2
  }

  kube::log::stack $stack_skip

  echo "Exiting with status ${code}" >&2
  exit "${code}"
}

# Log an error but keep going.  Don't dump the stack or exit.
kube::log::error() {
  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "!!! $timestamp ${1-}" >&2
  shift
  for message; do
    echo "    $message" >&2
  done
}
