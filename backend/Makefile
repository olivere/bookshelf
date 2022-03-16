GO?=go
IMAGENAME=bookshelf/backend:latest
ECR=$(AWS_ACCOUNT_ID).dkr.ecr.eu-central-1.amazonaws.com

default: build

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build

.PHONY: restart
restart:
	docker compose up --build -d --no-deps backend

.PHONY: container
container: build
	docker build -t $(IMAGENAME) .

.PHONY: push
push: container
	docker tag $(IMAGENAME) $(ECR)/$(IMAGENAME)
	docker push $(ECR)/$(IMAGENAME)
