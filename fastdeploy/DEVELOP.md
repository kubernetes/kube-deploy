Work in progress documentation on how to run this against a cluster.

I take kube-up, and make this change (on the AWS side) so that the script doesn't run:

```
  BOOTSTRAP_SCRIPT="${KUBE_TEMP}/bootstrap-script"

  (
+    echo -e '#!/bin/sh\nexit 0\n'
    # Include the default functions from the GCE configure-vm script
    sed '/^#+AWS_OVERRIDES_HERE/,$d' "${KUBE_ROOT}/cluster/gce/configure-vm.sh"
    # Include the AWS override functions
```

Then I run kube-up, and take the MASTER_IP & NODE_IP...


```
MASTER_IP=<ip>

echo "" | ssh admin@${MASTER_IP} sudo apt-get update
echo "" | ssh admin@${MASTER_IP} sudo apt-get --yes install rsync time
echo "" | ssh admin@${MASTER_IP} sudo mkdir -p /opt/fastdeploy
echo "" | ssh admin@${MASTER_IP} sudo chown admin /opt/fastdeploy

# Copy the assets to the server (TODO it should download them directly)
tar zxf kubernetes.tar.gz kubernetes/server/kubernetes-server-linux-amd64.tar.gz
rsync  kubernetes/server/kubernetes-server-linux-amd64.tar.gz admin@${MASTER_IP}:/tmp/kubernetes-server-linux-amd64.tar.gz


go install k8s.io/kube-deploy/fastdeploy && \
  rsync -avz ~/k8s/bin/fastdeploy admin@${MASTER_IP}:/opt/fastdeploy/fastdeploy && \
  rsync -avz --delete ~/k8s/src/k8s.io/kube-deploy/fastdeploy/template/ admin@${MASTER_IP}:/opt/fastdeploy/template && \
  ssh admin@${MASTER_IP} sudo time /opt/fastdeploy/fastdeploy  --template /opt/fastdeploy/template  --tags=_kubernetes_master,_systemd,_aws,_jessie,_debian_family --v=2


```
NODE_IP=<ip>

echo "" | ssh admin@${NODE_IP} sudo apt-get update
echo "" | ssh admin@${NODE_IP} sudo apt-get --yes install rsync time
echo "" | ssh admin@${NODE_IP} sudo mkdir -p /opt/fastdeploy
echo "" | ssh admin@${NODE_IP} sudo chown admin /opt/fastdeploy

rsync  kubernetes/server/kubernetes-server-linux-amd64.tar.gz admin@${NODE_IP}:/tmp/kubernetes-server-linux-amd64.tar.gz

go install k8s.io/kube-deploy/fastdeploy && \
  rsync -avz ~/k8s/bin/fastdeploy admin@${NODE_IP}:/opt/fastdeploy/fastdeploy && \
  rsync -avz --delete ~/k8s/src/k8s.io/kube-deploy/fastdeploy/template/ admin@${NODE_IP}:/opt/fastdeploy/template && \
  ssh admin@${NODE_IP} sudo time /opt/fastdeploy/fastdeploy  --template /opt/fastdeploy/template  --tags=_kubernetes_pool,_systemd,_aws,_jessie,_debian_family --v=2
```
