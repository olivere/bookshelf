GO?=go
IMAGENAME=bookshelf/frontend:latest
ECR=935517411172.dkr.ecr.eu-central-1.amazonaws.com

default: build

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build

.PHONY: restart
restart:
	docker compose up --build -d --no-deps frontend

.PHONY: container
container: build
	docker build -t $(IMAGENAME) .

.PHONY: push
push: container
	docker tag $(IMAGENAME) $(ECR)/$(IMAGENAME)
	docker push $(ECR)/$(IMAGENAME)
