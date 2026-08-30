package main

# Required USER value for security compliance
required_user := "65532:65532"

# List of Dockerfile paths that are exceptions for user validation
user_exception_dockerfiles := {
    "provisioner-service/Dockerfile"
}

# Check if the current Dockerfile is in the exception list for user validation
user_exception_exists {
    user_exception_dockerfiles[input.File]
}

# Check if the USER command exists with the correct value
user_exists {
    some i
    lower(input[i].Cmd) == "user"
    input[i].Value[0] == required_user
}

# Deny if the USER command is missing or has wrong value
deny_user[msg] {
    not user_exception_exists
    not user_exists
    msg := sprintf("%s - Dockerfile must contain 'USER %s' for security compliance.", [input.File, required_user])
}
