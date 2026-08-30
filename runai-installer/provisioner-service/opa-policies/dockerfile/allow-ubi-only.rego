package main

# Allowed base images
allowed_images := {
    "registry.access.redhat.com/ubi9/ubi-minimal",
    "registry.access.redhat.com/ubi9/nodejs-18-minimal",
    "registry.access.redhat.com/ubi9/nginx-120",
    "registry.access.redhat.com/ubi9/openjdk-17"
}

# List of Dockerfile paths that are exceptions
exception_dockerfiles := {
    "provisioner-service/Dockerfile"
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
deny_allowed_images[msg] {
    not exception_exists
    some i
    input[i].Cmd == "from"
    val := input[i].Value[0]
    not contains(allowed_images, val)
    msg := sprintf("%s - Image must use UBI registry and allowed tag: %s", [input.File, val])
}