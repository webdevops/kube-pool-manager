PROJECT_NAME		:= kube-pool-manager
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS				:= -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static"

FIRST_GOPATH			:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN		:= $(FIRST_GOPATH)/bin/golangci-lint

.PHONY: all
all: build

.PHONY: clean
clean:
	git clean -Xfd .

.PHONY: build
build:
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(PROJECT_NAME) .

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: image
image: build
	docker build -t $(PROJECT_NAME):$(GIT_TAG) .

build-push-development:
	docker build -t webdevops/$(PROJECT_NAME):development . && docker push webdevops/$(PROJECT_NAME):development

.PHONY: test
test:
	go test ./...

.PHONY: dependencies
dependencies:
	go mod vendor

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run -E exportloopref,gofmt --timeout=25m

.PHONY: gosec
gosec: $(GOSEC_BIN)
	$(GOSEC_BIN) ./...

$(GOLANGCI_LINT_BIN):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

$(GOSEC_BIN):
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(FIRST_GOPATH)/bin v2.7.0
