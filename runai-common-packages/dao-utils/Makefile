include ../../build/makefile/index.mk

SERVICE_NAME=dao-utils
TAG_FORMAT=runai-common-packages/$(SERVICE_NAME)

DOCKER_WORK_DIR=/local/runai-common-packages/${SERVICE_NAME}
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

test: git-config-command 
	go test ./...
.PHONY: test

lint: lint-go ## Run linter.
.PHONY: lint

validate:
	# no op
.PHONY: validate

build:
	# no op
.PHONY: build

purify:
	go fmt ./...
	go vet ./...
	go mod tidy
.PHONY: purify
