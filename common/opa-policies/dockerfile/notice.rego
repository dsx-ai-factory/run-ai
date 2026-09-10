# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package main

# Deny if the 'COPY NOTICE' or 'ADD NOTICE' command is missing
deny[msg] {
    not notice_exists
    msg := "Dockerfile is missing the 'COPY NOTICE' or 'ADD NOTICE' command."
}

# Check for the 'COPY NOTICE' or 'ADD NOTICE' command
notice_exists {
    some i
    lower(input[i].Cmd) == "copy"  # Check COPY command
    some j
    lower(input[i].Value[j]) == "notice"
}

notice_exists {
    some i
    lower(input[i].Cmd) == "add"  # Check ADD command
    some j
    lower(input[i].Value[j]) == "notice"
}

# Debugging message
debug_msg := "Checking Dockerfile"
