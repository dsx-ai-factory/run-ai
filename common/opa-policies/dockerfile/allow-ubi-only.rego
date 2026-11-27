# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package main

# Allowed base images
allowed_images := {
    "registry.access.redhat.com/ubi9/ubi-minimal",
    "registry.access.redhat.com/ubi9/ubi-minimal:9.7@sha256:907b68736aa798b2d38255b7aa070b2a70acb90803864a40f05d0ec47556ddd0",
    "registry.access.redhat.com/ubi9/nodejs-18-minimal",
    "registry.access.redhat.com/ubi9/nodejs-24-minimal",
    "registry.access.redhat.com/ubi9/nodejs-24-minimal:9.8-1785358511@sha256:a0245449f38996a5c1f5b0a7fa1e7245050eb044441cb8dab61652364ea06b49",
    "registry.access.redhat.com/ubi9/nginx-120",
    "registry.access.redhat.com/ubi9/openjdk-17",
    "registry.access.redhat.com/ubi9/ubi-micro",
    "docker.io/library/alpine:3.21",
    "scratch"
}

# List of Dockerfile paths that are exceptions
exception_dockerfiles := {
    "runai-cluster/cluster/git-sync/Dockerfile",
    "control-plane/notifications-service/Dockerfile",
    "control-plane/keycloak-extensions/Dockerfile"
}

# Function to check if a value exists in a set
contains(set, elem) {
    set[elem]
}

# Check if the current Dockerfile is in the exception list
exception_exists {
    contains(exception_dockerfiles, input.File)
}

# Deny if the image is not allowed and the file is not an exception
deny[msg] {
    not exception_exists
    some i
    input[i].Cmd == "from"
    val := input[i].Value[0]
    not contains(allowed_images, val)
    msg := sprintf("Image must use UBI registry and allowed tag: %s", [val])
}
