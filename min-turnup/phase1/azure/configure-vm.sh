#! /bin/bash

set -x
set -o errexit
set -o pipefail
set -o nounset

mkdir -p /etc/systemd/system/docker.service.d/
cat <<EOF > /etc/systemd/system/docker.service.d/clear_mount_propagtion_flags.conf
[Service]
MountFlags=shared
EOF
cat <<EOF > /etc/systemd/system/docker.service.d/overlay.conf
[Service]
ExecStart=
ExecStart=/usr/bin/docker daemon -H fd:// --storage-driver=overlay
EOF

curl -sSL https://get.docker.com/ | sh

apt-get update
#apt-get dist-upgrade -y
apt-get install -y jq

systemctl start docker || true

ROLE="node"
if [[ $(hostname) = *master* ]]; then
	ROLE="master"
fi

azure_file="/etc/kubernetes/azure.json"
config_file="/etc/kubernetes/k8s_config.json"

mkdir /etc/kubernetes
# these get filled in from terraform
echo -n "${azure_json}" | base64 -d > "$azure_file"
echo -n "${k8s_config}" | base64 -d > "$config_file"
echo -n "${kubelet_tar}" | base64 -d > "/etc/kubernetes/kubelet.tar"
echo -n "${root_tar}" | base64 -d > "/etc/kubernetes/root.tar"
echo -n "${apiserver_tar}" | base64 -d > "/etc/kubernetes/apiserver.tar"

MASTER_IP="$(cat "$config_file" | jq -r '.phase1.azure.master_private_ip')"

jq ". + {\"role\": \"$ROLE\", \"master_ip\": \"$MASTER_IP\"}" "$config_file" > /etc/kubernetes/k8s_config.new; cp /etc/kubernetes/k8s_config.new "$config_file"

mkdir -p /srv/kubernetes
for bundle in root kubelet apiserver; do
  cat "/etc/kubernetes/$bundle.tar" | sudo tar xv -C /srv/kubernetes
done;

installer_container_spec="$(cat "$config_file" | jq -r '.phase2.installer_container_spec')"

cat << EOF > /etc/kubernetes/install.sh
systemctl stop docker
systemctl start docker
docker pull "$installer_container_spec"
docker run \
  --net=host \
  -v /:/host_root \
  -v /etc/kubernetes/k8s_config.json:/opt/playbooks/config.json:ro \
  "$installer_container_spec" \
  /opt/do_role.sh "$ROLE"
EOF

chmod +x /etc/kubernetes/install.sh
/etc/kubernetes/install.sh

#sudo reboot
