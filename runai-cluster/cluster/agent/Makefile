include ../../../build/makefile/runai-cluster.mk
include ../../../build/makefile/index.mk
DOCKER_WORK_DIR=${DOCKER_WORK_DIR_CLUSTER}

# Change for every new service
SERVICE_NAME=agent
##@ Basic Commands

test: test-go ## Run unit-tests.
.PHONY: test

test-dev:
	go test -v ./... -coverprofile cover.out
.PHONY: test-dev

lint: lint-go ## Run linter.
.PHONY: lint

validate:
  #no-op
.PHONY: validate

format: format-go ## Format the code.
.PHONY: format

build: build-go ## Build the service.
.PHONY: build

local-deploy: export OS=linux
local-deploy: export ARCH=amd64
local-deploy: build-go docker-build-generic docker-push-generic
    ## Build and push the docker image.
.PHONY: local-deploy

generate-projects:
	rm -f -r pkg/generated/projects_openapi/*
	docker run --rm -v "${PWD}:/local" \
	    -e JAVA_OPTS="-Dlog.level=warn" \
	    openapitools/openapi-generator-cli:v6.0.1 generate \
	    -i /local/swagger/projectsDepartmentsApi.yaml \
	    --package-name projects_openapi \
	    --additional-properties enumClassPrefix=true \
	    --additional-properties onlyInterfaces=true \
	    --additional-properties outputAsLibrary=true \
	    -g go \
	    -o /local/pkg/generated/projects_openapi
	    rm pkg/generated/projects_openapi/go.*

generate-backend:
	rm -f -r pkg/generated/backend_openapi/*
	docker run --rm -v "${PWD}:/local" \
	    -e JAVA_OPTS="-Dlog.level=warn" \
	    openapitools/openapi-generator-cli:v6.0.1 generate \
	    -i /local/swagger/clusterApi.yaml \
	    --package-name backend_openapi \
	    --additional-properties enumClassPrefix=true \
	    --additional-properties onlyInterfaces=true \
	    --additional-properties outputAsLibrary=true \
	    -g go \
	    -o /local/pkg/generated/backend_openapi
	    rm pkg/generated/backend_openapi/go.*

generate-authorization:
	rm -f -r pkg/generated/auth_openapi/*
	docker run --rm -v "${PWD}:/local" \
	    -e JAVA_OPTS="-Dlog.level=warn" \
	    openapitools/openapi-generator-cli:v6.0.1 generate \
	    -i /local/swagger/swagger.yaml \
	    --package-name auth_openapi \
	    --additional-properties enumClassPrefix=true \
	    --additional-properties onlyInterfaces=true \
	    --additional-properties outputAsLibrary=true \
	    -g go \
	    -o /local/pkg/generated/auth_openapi
	    rm pkg/generated/auth_openapi/go.*

generate-datavolumes:
	rm -f -r pkg/generated/datavolumes_openapi/*
	docker run --rm -v "${PWD}:/local" \
	    -e JAVA_OPTS="-Dlog.level=warn" \
	    openapitools/openapi-generator-cli:v6.0.1 generate \
	    -i /local/swagger/datavolumesApi.yaml \
	    --package-name datavolumes_openapi \
	    --additional-properties enumClassPrefix=true \
	    --additional-properties onlyInterfaces=true \
	    --additional-properties outputAsLibrary=true \
	    -g go \
	    -o /local/pkg/generated/datavolumes_openapi
	    rm pkg/generated/datavolumes_openapi/go.*

generate: generate-projects generate-backend generate-authorization generate-datavolumes
