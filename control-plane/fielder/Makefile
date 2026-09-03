
include ../../build/makefile/control-plane.mk
include ../../build/makefile/index.mk

SERVICE_NAME=fielder
PYTHON_REQ_PATH=.

##@ Basic Commands

test: 
.PHONY: test

lint: 
.PHONY: lint

validate: 
	@echo "Running validation mode..."
	$(MAKE) rebuild PYTHON_SCRIPT_FILE_NAME=". --validate"
.PHONY: validate

purify: 
.PHONY: purify

build: # developer should run 'make rebuild' locally and then merge the changes. in CI we only validate
.PHONY: build

rebuild: 
	$(DOCKER_PYTHON_COMMAND) 
.PHONY: build
