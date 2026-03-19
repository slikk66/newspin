-include .env

AWS_ACCOUNT ?=
AWS_REGION ?= us-west-2
AWS_PROFILE ?=
ECR_REPO := $(AWS_ACCOUNT).dkr.ecr.$(AWS_REGION).amazonaws.com/newspin-newspin
TAG ?= latest

.PHONY: build run ecr-login buildimage pushimage deploy

build:
	go build ./cmd/server/

run:
	go run ./cmd/server/

ecr-login:
	aws ecr get-login-password --region $(AWS_REGION) --profile $(AWS_PROFILE) | docker login --username AWS --password-stdin $(AWS_ACCOUNT).dkr.ecr.$(AWS_REGION).amazonaws.com

buildimage:
	docker build -t newspin:$(TAG) .

pushimage: ecr-login
	docker tag newspin:$(TAG) $(ECR_REPO):$(TAG)
	docker push $(ECR_REPO):$(TAG)

deploy: buildimage pushimage
