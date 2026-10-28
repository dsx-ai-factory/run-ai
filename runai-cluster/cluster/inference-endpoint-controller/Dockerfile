# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.7@sha256:907b68736aa798b2d38255b7aa070b2a70acb90803864a40f05d0ec47556ddd0 AS base

# Wrap /tmp inside /temp so /tmp lands as a *child* during COPY. Docker BuildKit resets the
# mode of any destination directory it creates to 0755, but preserves modes on children
RUN mkdir -p /temp/tmp && \
    chown 65532:0 /temp/tmp && \
    chmod 770 /temp/tmp

FROM scratch

ARG TARGETARCH=amd64
ENV TARGETARCH=$TARGETARCH

# Identity & Permissions
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

# Networking & SSL
COPY --from=base /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=base /etc/pki/ /etc/pki/

# DNS resolution config
COPY --from=base /etc/nsswitch.conf /etc/nsswitch.conf

# Writable Filesystem: Copy /temp's contents into / so /tmp lands at /tmp as a child (mode preserved)
COPY --from=base /temp /

WORKDIR /

COPY --chown=65532:0 bin/inference-endpoint-controller-$TARGETARCH inference-endpoint-controller
COPY NOTICE .

USER 65532:65532

ENTRYPOINT ["./inference-endpoint-controller"]
