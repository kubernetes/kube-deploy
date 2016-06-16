TAG=v1
IMAGE?=docker.io/colemickens/install-k8s

build:
	docker build -t "$(IMAGE):$(TAG)" .

push: build
	docker push "$(IMAGE):$(TAG)"

.PHONY: build push
