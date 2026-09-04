FROM registry.access.redhat.com/ubi9/ubi-minimal AS base

# Wrap /tmp inside /temp so /tmp lands as a *child* during COPY. Docker BuildKit resets the
# mode of any destination directory it creates to 0755, but preserves modes on children
RUN mkdir -p /temp/tmp && \
    chown 65532:0 /temp/tmp && \
    chmod 770 /temp/tmp

FROM scratch

ARG TARGETARCH
ENV TARGETARCH=$TARGETARCH

# Identity & Permissions
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

# Networking & SSL
COPY --from=base /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=base /etc/pki/ /etc/pki/

# Writable Filesystem: Copy /temp's contents into / so /tmp lands at /tmp as a child (mode preserved)
COPY --from=base /temp /

WORKDIR /

COPY bin/policy-service-$TARGETARCH policy-service
COPY NOTICE .

USER 65532:65532

COPY pkg/database/migrations/up/* ./migrations/up/
COPY pkg/database/migrations/down/* ./migrations/down/
COPY swagger/_build/* ./swagger/

ENTRYPOINT ["./policy-service"]
