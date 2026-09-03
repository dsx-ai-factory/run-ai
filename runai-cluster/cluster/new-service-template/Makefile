include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=new-service-template
# Only scan cmd directory to exclude cookiecutter template
GO_PACKAGES=./cmd/...

# Support building in debug mode
ifeq ($(DEBUG), 1)
GO_BUILD_ADDITIONAL_FLAGS=-gcflags="all=-N -l"
DOCKERFILE_PATH=./Dockerfile.debug
DOCKER_GO_COMMAND=echo "nop" &&
endif

##@ Basic Commands

test: test-go ## Run unit-tests.
.PHONY: test

# Override test-go to only test cmd directory (exclude cookiecutter template)
test-go: git-config-command gocache go-list
	@ ${ECHO_COMMAND} ${GREEN_CONSOLE} "${CONSOLE_PREFIX} Running unit-tests for ${SERVICE_NAME}" ${BASE_CONSOLE}
	${DOCKER_GO_COMMAND_LINUX} go test ${GO_PACKAGES} -v -failfast -coverprofile=$(COVERAGE_REPORT_PATH) || ${FAILURE_MESSAGE_HANDLER}
	${SUCCESS_MESSAGE_HANDLER}
.PHONY: test-go

test-dev:
	go test -v ${GO_PACKAGES} -coverprofile cover.out
.PHONY: test-dev

lint:
	#no-op
.PHONY: lint

validate:
	#no-op
.PHONY: validate

format: format-go ## Format the code.
.PHONY: format

# Override format-go to only format cmd directory (exclude cookiecutter template)
format-go: git-config-command gocache
	${DOCKER_GO_COMMAND} go fmt ${GO_PACKAGES}
.PHONY: format-go

build: build-go ## Build the service.
.PHONY: build

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy
