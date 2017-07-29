aws_infrastructure_plan:
	cd terraform/aws/ && \
	terraform init && \
	terraform plan -out="imagebuilder.plan"

aws_infrastructure:
	cd terraform/aws/ && \
	terraform apply "imagebuilder.plan"

aws_cleanup_plan:
	cd terraform/aws/ && \
	terraform plan -destroy

aws_cleanup:
	cd terraform/aws/ && \
	terraform destroy

install:
	go install k8s.io/kube-deploy/imagebuilder

gofmt:
	gofmt -w -s main.go
	gofmt -w -s pkg
