FROM python:2.7.11-slim

ADD kube-addons.sh /opt/
ADD kube-addon-update.sh /opt/
ADD namespace.yaml /opt/

ENV KUBECTL_VERSION v1.2.1-beta.0
ADD "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
	/usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl

CMD /opt/kube-addons.sh
