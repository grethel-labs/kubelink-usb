include project.env

IMG_CONTROLLER ?= $(REGISTRY)/$(CONTROLLER_IMG):$(VERSION)
IMG_AGENT      ?= $(REGISTRY)/$(AGENT_IMG):$(VERSION)

CHART_DIR      = charts/kubelink-usb

.PHONY: build docker-build install run-controller run-agent test test-cover coverage-check fmt lint lint-golangci docs docs-api docs-deps docs-diagrams helm-lint helm-template sync-chart-version

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

lint-golangci:
	golangci-lint run ./...

build:
	go build -o bin/controller ./cmd/controller
	go build -o bin/agent ./cmd/agent
	go build -o bin/kubectl-usb ./cmd/kubectl-usb

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(IMG_CONTROLLER) -f Dockerfile .
	docker build --build-arg VERSION=$(VERSION) -t $(IMG_AGENT) -f Dockerfile.agent .

install:
	kubectl apply -f config/crd/bases/

run-controller:
	go run ./cmd/controller

run-agent:
	go run ./cmd/agent

test:
	go test ./...

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic

coverage-check:
	./hack/coverage-check.sh

docs: docs-api docs-diagrams
	./hack/generate-code-reference.sh

docs-api:
	gomarkdoc --repository.url "https://github.com/grethel-labs/kubelink-usb" --repository.default-branch main --repository.path / --output '{{.Dir}}/DOC.md' ./...

docs-diagrams:
	./hack/generate-diagrams.sh

docs-deps:
	goda graph "github.com/grethel-labs/kubelink-usb/...:mod" | dot -Tsvg -o docs/dependency-graph.svg

helm-lint:
	helm lint $(CHART_DIR)

helm-template:
	helm template kubelink-usb $(CHART_DIR)

sync-chart-version:
	@sed -i.bak 's/^version: .*/version: $(VERSION)/' $(CHART_DIR)/Chart.yaml
	@sed -i.bak 's/^appVersion: .*/appVersion: "$(VERSION)"/' $(CHART_DIR)/Chart.yaml
	@rm -f $(CHART_DIR)/Chart.yaml.bak
	@echo "Chart.yaml synced to version $(VERSION)"
