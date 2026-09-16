# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM registry.access.redhat.com/ubi9/ubi-minimal AS base

# Wrap /tmp inside /temp so /tmp lands as a *child* during COPY. Docker BuildKit resets the
# mode of any destination directory it creates to 0755, but preserves modes on children
RUN mkdir -p /temp/tmp && \
    chown 65532:0 /temp/tmp && \
    chmod 770 /temp/tmp

FROM scratch

LABEL name="Run.AI Karta Migration"
LABEL description="Run.AI Karta migration job"
LABEL summary="Migrates legacy ResourceInterface CRs to the new Karta CRD"
LABEL vendor="Runai Labs LTD"
LABEL maintainer=support@run.ai

ARG TARGETARCH
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

WORKDIR /app

COPY --chown=65532:0 bin/karta-migrate-$TARGETARCH /app/karta-migrate
COPY NOTICE .

USER 65532:65532

ENTRYPOINT ["/app/karta-migrate"]
