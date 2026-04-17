IMG_CONTROLLER ?= ghcr.io/yourname/k8s-usb-fabric-controller:latest
IMG_AGENT ?= ghcr.io/yourname/k8s-usb-fabric-agent:latest

.PHONY: build docker-build install run test fmt

fmt:
	go fmt ./...

build:
	go build -o bin/controller ./
	go build -o bin/agent ./cmd/agent

docker-build:
	docker build -t $(IMG_CONTROLLER) -f Dockerfile .
	docker build -t $(IMG_AGENT) -f Dockerfile.agent .

install:
	kubectl apply -f config/crd/bases/

run:
	go run ./main.go

test:
	go test ./...
