IMG_CONTROLLER ?= ghcr.io/grethel-labs/kubelink-usb-controller:latest
IMG_AGENT ?= ghcr.io/grethel-labs/kubelink-usb-agent:latest

.PHONY: build docker-build install run test test-cover coverage-check fmt lint docs

fmt:
	go fmt ./...

lint:
	files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "The following files are not gofmt-formatted:"; \
		echo "$$files"; \
		exit 1; \
	fi
	go vet ./...

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

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic

coverage-check:
	./hack/coverage-check.sh

docs:
	./hack/generate-code-reference.sh
