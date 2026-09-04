include ../../build/makefile/index.mk

SERVICE_NAME=metricsctr
TAG_FORMAT=runai-common-packages/$(SERVICE_NAME)

DOCKER_WORK_DIR=/local/runai-common-packages/${SERVICE_NAME}
DOCKER_COMMAND=${DOCKER_COMMAND_WORK_DIR}

test: test-go
.PHONY: test

lint: lint-go 
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
