
SHELL=/bin/bash

# TODO: things break with this uncommented
#.SHELLFLAGS="-O extglob -o errexit -o pipefail -o nounset -c"

.PHONY: config echo-config


# sorry windows and non amd64
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	OS = linux
endif
ifeq ($(UNAME_S),Darwin)
	OS = darwin
endif

CONF_TOOL_VERSION = 4.6
KCONFIG_FILES = $(shell find . -name 'Kconfig')


docker-build:
	docker build -t kube-min-turnup .

docker-run: docker-build
	docker run -it --net=host -v `pwd`:/opt/kube-min-turnup kube-min-turnup /bin/bash

default:
	$(MAKE) config

.tmp/conf:
	mkdir -p .tmp; curl -sSL --fail -o "$@" \
		"https://storage.googleapis.com/public-mikedanese-k8s/kconfig/$(CONF_TOOL_VERSION)/$(OS)/conf"; \
	chmod +x "$@"

.tmp/mconf:
	mkdir -p .tmp; curl -sSL --fail -o "$@" \
		"https://storage.googleapis.com/public-mikedanese-k8s/kconfig/$(CONF_TOOL_VERSION)/$(OS)/mconf"; \
	chmod +x "$@"

config: .tmp/conf
	CONFIG_="" .tmp/conf Kconfig

menuconfig: .tmp/mconf
	CONFIG_="" .tmp/mconf Kconfig

.config: .tmp/conf $(KCONFIG_FILES)
	$(MAKE) config

.config.json: .config
	util/config_to_json $< > $@

echo-config: .config.json
	cat $<

set-kubeconfig:
	kubectl config  set-cluster $(CLUSTER_NAME) --server https://$(MASTER_IP):443 --certificate-authority=phase1b/crypto/ca.pem --embed-certs=true
	kubectl config set-credentials $(CLUSTER_NAME) --server https://$(MASTER_IP):443 --client-certificate=./phase1b/crypto/kubelet.pem --client-key=./phase1b/crypto/kubelet-key.pem --embed-certs=true
	kubectl config set-context $(CLUSTER_NAME) --cluster $(CLUSTER_NAME) --user $(CLUSTER_NAME)

clean:
	rm -rf .tmp

fmt:
	for f in $$(find . -name '*.jsonnet'); do jsonnet fmt -i -n 2 $${f}; done;
