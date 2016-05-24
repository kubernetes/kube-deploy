FROM docker.io/buildpack-deps:xenial

RUN bash -c "\
	apt-get update && apt-get -y upgrade && \
	apt-get install -y make unzip jq && \
	rm -rf /var/lib/apt/lists/*"

ENV TERRAFORM_VERSION 0.7.0-rc1
RUN bash -c "\
	wget -q -O /tmp/terraform.zip \"https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip\" && \
	unzip /tmp/terraform.zip -d /usr/local/bin/ && \
	rm /tmp/terraform.zip"

################################################################################################
# TEMP: For now get terraform from source so we get ipv4 forwarding for Azure
################################################################################################
#RUN bash -c "\
#	apt-get update && apt-get -y upgrade && \
#	apt-get install -y nodejs npm golang-go && \
#	rm -rf /var/lib/apt/lists/*"
#ENV GOPATH "/root/gopkgs"
#ENV PATH "$GOPATH/bin:$PATH"
#RUN bash -c "mkdir -p $GOPATH/src/github.com/hashicorp/ && \
#	git clone https://github.com/hashicorp/terraform $GOPATH/src/github.com/hashicorp/terraform && \
#	cd $GOPATH/src/github.com/hashicorp/terraform; make dev"

################################################################################################
# Azure - Required dependencies
################################################################################################
RUN bash -c "\
	apt-get update && apt-get -y upgrade && \
	apt-get install -y nodejs npm && \
	rm -rf /var/lib/apt/lists/*"
RUN bash -c "npm install -g azure-cli"

################################################################################################

################################################################################################
# Jsonnet
################################################################################################

RUN bash -c "\
	cd /tmp; \
	git clone https://github.com/google/jsonnet; \
	cd jsonnet; \
	make; \
	cp jsonnet /usr/bin/jsonnet;"

WORKDIR /opt/kube-deploy

ADD . /opt/kube-deploy/

CMD make
