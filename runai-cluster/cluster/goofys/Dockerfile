# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Multi-stage build producing a scratch-based runtime image.
#
# The runtime stage carries only the goofys binary, CA certificates and an
# empty mount point, so the image ships no OS packages and therefore no OS
# package CVEs. This is viable because of three properties of how goofys is
# actually run here:
#
#   * The binary is built with CGO_ENABLED=0 (see build/makefile/golang.mk),
#     so it is statically linked and needs no libc at runtime.
#   * Every goofys container is created with Privileged: true (see
#     control-plane/workloads-and-policies/composer/init_containers/s3.go).
#     With CAP_SYS_ADMIN, jacobsa/fuse mounts by opening /dev/fuse and calling
#     mount(2) directly (mount_linux.go directmount) and never shells out to
#     fusermount, which is what previously required the fuse package.
#   * Callers always set an explicit command of ["goofys", "-f", ...] rather
#     than relying on the image entrypoint, so no shell is needed.
#
# Two things are deliberately no longer in the image:
#
#   * fusermount. unmount (jacobsa/fuse unmount_linux.go) still execs it, so
#     that call now fails. Teardown is handled by the container exiting and
#     the kernel releasing the mount, not by that call succeeding.
#   * catfs. goofys only needs it for its --cache mode (internal/flags.go).
#     Nothing in the platform passes --cache; goofysBaseCommand in
#     init_containers/s3.go does not include it. Running this image with
#     --cache will fail.
#
# rootfs/ (run.sh and the syslog config) is no longer copied in either, since
# it needs a shell. Note that run.sh invoked syslog-ng while the old image
# installed rsyslog, so that path was already broken.

# ---- fetch stage: build-time only, nothing here reaches the final image ----
FROM registry.access.redhat.com/ubi9/ubi-minimal AS fetch

RUN microdnf update -y && microdnf install -y ca-certificates && microdnf clean all

# An empty directory to serve as the mount point. scratch has no shell, so the
# directory cannot be created with mkdir in the runtime stage.
RUN mkdir -p /mnt/s3

# ---- runtime stage ----
FROM scratch

ARG TARGETARCH

COPY --from=fetch /etc/pki/tls/certs/ca-bundle.crt /etc/pki/tls/certs/ca-bundle.crt
COPY --from=fetch /etc/pki/tls/certs/ca-bundle.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=fetch /mnt/s3 /mnt/s3

COPY --chmod=0755 bin/goofys-$TARGETARCH /usr/local/bin/goofys

COPY NOTICE /NOTICE

ENV MOUNT_DIR=/mnt/s3
ENV REGION=us-east-1
ENV BUCKET=teleport-bucket
ENV STAT_CACHE_TTL=1m0s
ENV TYPE_CACHE_TTL=1m0s
ENV DIR_MODE=0700
ENV FILE_MODE=0600
ENV UID=0
ENV GID=0

VOLUME /mnt/s3

ENTRYPOINT ["/usr/local/bin/goofys"]
