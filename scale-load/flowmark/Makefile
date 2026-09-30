# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../build/makefile/index.mk

SERVICE_NAME=flowmark
TAG_FORMAT=scale-load/$(SERVICE_NAME)
DOCKER_WORK_DIR=/local/scale-load/${SERVICE_NAME}
GOLANG_LINTER_CONFIG_PATH=../.golangci.yaml
BIN_DIR=$(shell pwd)/bin

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR) ./cmd/...
.PHONY: build

test: test-go
.PHONY: test

lint: lint-go
.PHONY: lint

validate:
	# no op
.PHONY: validate

SCENARIO ?=
WORKLOADS ?=

run:
	@set -a; [ -f .env ] && . ./.env; set +a; \
	go run ./cmd/flowmark --scenarios ./scenarios --output ./reports \
		$(if $(SCENARIO),--only $(SCENARIO)) $(if $(WORKLOADS),--workloads $(WORKLOADS))
.PHONY: run
