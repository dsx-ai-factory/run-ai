# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ARG NGINX_IMAGE=nginx:1.31.7
FROM ${NGINX_IMAGE} AS nginx-source

FROM registry.access.redhat.com/ubi9/ubi-minimal AS fips-builder
RUN microdnf install -y openssl-libs && microdnf clean all && \
    test -f /usr/lib64/ossl-modules/fips.so

FROM registry.access.redhat.com/ubi9/ubi-minimal AS distribution-assets

RUN microdnf install -y jq && microdnf clean all && \
    command -v curl >/dev/null && \
    command -v jq >/dev/null && \
    command -v sha256sum >/dev/null

# The manifest is committed for the initial implementation. Its input can later
# move to a BuildKit secret mount without changing the downloader contract.
COPY distribution-assets.json /tmp/distribution-assets.json
COPY scripts/download-distribution-assets.sh /tmp/download-distribution-assets.sh

RUN /tmp/download-distribution-assets.sh \
    /tmp/distribution-assets.json \
    /distribution-assets

FROM nginx-source AS base

USER root

# Create the user and group so the scratch image recognizes them
RUN echo "nonroot:x:1001:1001:nonroot:/home/nonroot:/sbin/nologin" >> /etc/passwd && \
    echo "nonroot:x:1001:" >> /etc/group

# Pre-create all nginx temp dirs and the pid dir as subdirectories so their
# g+rwX permissions survive COPY to scratch (Docker resets the COPY root dir
# to 755, but preserves permissions of children). Nginx finds them already
# present and writes into them without needing to mkdir.
RUN mkdir -p /var/cache/nginx/client_temp \
             /var/cache/nginx/proxy_temp \
             /var/cache/nginx/fastcgi_temp \
             /var/cache/nginx/uwsgi_temp \
             /var/cache/nginx/scgi_temp \
             /var/lib/nginx/tmp/client_body \
             /var/log/nginx \
             /nginx-run/pid && \
    touch /var/lib/nginx/.keep /var/log/nginx/.keep /nginx-run/pid/nginx.pid && \
    chown -R 1001:0 /var/cache/nginx /var/lib/nginx /var/log/nginx /nginx-run && \
    chmod -R g+rwX /var/cache/nginx /var/lib/nginx /var/log/nginx /nginx-run

# Stage shared libraries and dynamic linker under /dynlib for a clean COPY to scratch.
# Debian uses arch-specific lib paths (/lib/ARCH-linux-gnu/); handles aarch64 and amd64.
# cp -L dereferences symlinks so each destination file is a real file.
COPY --from=fips-builder /usr/lib64/ossl-modules/fips.so /tmp/fips.so
RUN ARCH=$(uname -m) && \
    LIB="/lib/${ARCH}-linux-gnu" && \
    DEST="/dynlib${LIB}" && \
    mkdir -p "${DEST}" && \
    cp -L "${LIB}/libssl.so.3"     "${DEST}/libssl.so.3" && \
    cp -L "${LIB}/libcrypto.so.3"  "${DEST}/libcrypto.so.3" && \
    cp -L "${LIB}/libz.so.1"       "${DEST}/libz.so.1" && \
    cp -L "${LIB}/libcrypt.so.1"   "${DEST}/libcrypt.so.1" && \
    cp -L "${LIB}/libpcre2-8.so.0" "${DEST}/libpcre2-8.so.0" && \
    cp -L "${LIB}/libc.so.6"       "${DEST}/libc.so.6" && \
    cp -L "${LIB}/libzstd.so.1"    "${DEST}/libzstd.so.1" && \
    if [ "$ARCH" = "aarch64" ]; then \
        # Dynamic linker: aarch64 → /lib/ld-linux-aarch64.so.1
        mkdir -p /dynlib/lib && \
        cp /lib/ld-linux-aarch64.so.1 /dynlib/lib/ld-linux-aarch64.so.1 && \
        # FIPS provider: NIST-validated fips.so from UBI, placed where libcrypto looks for it.
        mkdir -p /dynlib/usr/lib/aarch64-linux-gnu/ossl-modules && \
        cp /tmp/fips.so /dynlib/usr/lib/aarch64-linux-gnu/ossl-modules/fips.so; \
    else \
        # Dynamic linker: amd64 → /lib64/ld-linux-x86-64.so.2
        mkdir -p /dynlib/lib64 && \
        cp /lib64/ld-linux-x86-64.so.2 /dynlib/lib64/ld-linux-x86-64.so.2 && \
        # FIPS provider: NIST-validated fips.so from UBI, placed where libcrypto looks for it.
        mkdir -p /dynlib/usr/lib/x86_64-linux-gnu/ossl-modules && \
        cp /tmp/fips.so /dynlib/usr/lib/x86_64-linux-gnu/ossl-modules/fips.so; \
    fi

FROM base AS prep

COPY nginx.conf /etc/nginx/nginx.conf
COPY dist/ /usr/share/nginx/html
COPY --from=distribution-assets /distribution-assets/ce/ /usr/share/nginx/html/frontend/distribution-assets/ce/

# Prepend error_log/pid directives so nginx writes them to writable scratch paths.
RUN printf 'error_log /dev/stderr notice;\npid /nginx-run/pid/nginx.pid;\n' | \
        cat - /etc/nginx/nginx.conf > /tmp/nginx.conf && \
    mv /tmp/nginx.conf /etc/nginx/nginx.conf


FROM scratch AS runner

LABEL name="Run.AI Backend"
LABEL description="Run.AI Backend is the cloud portion of Run:AI which contains administrative interfaces"
LABEL summary="Run.AI Backend is the cloud portion of Run:AI which contains administrative interfaces"
LABEL vendor="Runai Labs LTD"
LABEL maintainer=support@run.ai
LABEL app=runai-frontend

# Identity & Permissions
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

# Networking & SSL
COPY --from=nginx-source /etc/ssl/certs/ /etc/ssl/certs/

# nginx binary
COPY --from=nginx-source /usr/sbin/nginx /usr/sbin/nginx

# Shared libraries and dynamic linker (arch-specific Debian paths, staged in base)
COPY --from=base /dynlib/ /

# nginx config and html (template substitutions already applied in prep stage)
COPY --from=prep /etc/nginx /etc/nginx
COPY --from=prep /usr/share/nginx /usr/share/nginx

# Writable nginx runtime directories — --chown ensures 1001 is owner, giving rwx on 755 dirs
COPY --chown=1001:0 --from=base /var/cache/nginx /var/cache/nginx
COPY --chown=1001:0 --from=base /var/log/nginx /var/log/nginx
COPY --chown=1001:0 --from=base /nginx-run /nginx-run
COPY --chown=1001:0 --from=base /var/lib/nginx /var/lib/nginx

COPY NOTICE .

# set env OPENSSL_CONF=/etc/ssl/openssl-fips.cnf to enable fips".
COPY openssl-fips.cnf /etc/ssl/openssl-fips.cnf

EXPOSE 8080

USER 1001:0

CMD ["nginx", "-g", "daemon off;"]
