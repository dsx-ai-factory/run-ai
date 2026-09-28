# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM registry.access.redhat.com/ubi9/ubi-minimal AS base

# Wrap the shared-volume mount points inside /temp so they land as *children*
# during COPY. Docker BuildKit resets the mode of any destination directory
# it creates to 0755, but preserves modes on children.
RUN mkdir -p /temp/providers /temp/themes /temp/agent && \
    chown -R 65532:0 /temp && \
    chmod -R 770 /temp

FROM scratch

ARG TARGETARCH

# Identity & permissions for the non-root UID this image runs as.
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

# Writable shared-volume mount points (770 mode preserved from the base stage).
COPY --from=base /temp /

WORKDIR /

COPY bin/runai-custom-providers.jar /extensions/providers/
COPY bin/opentelemetry-javaagent.jar /extensions/agent/
COPY runai-theme/ /extensions/themes/runai-theme/
COPY NOTICE /extensions/
COPY bin/keycloak-extensions-$TARGETARCH keycloak-extensions-init

USER 65532:65532

ENTRYPOINT ["./keycloak-extensions-init"]
