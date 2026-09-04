FROM registry.access.redhat.com/ubi9/ubi-minimal AS base

# Create the user and group so the scratch image recognizes them
RUN echo "nonroot:x:65532:65532:nonroot:/home/nonroot:/sbin/nologin" > /etc/passwd && \
    echo "nonroot:x:65532:" > /etc/group

# Create a dedicated writable directory and set ownership (group 0 for OpenShift)
RUN mkdir -p /tmp && \
    chown -R 65532:0 /tmp && \
    chmod 1777 /tmp

FROM scratch

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

# Writable Filesystem (owned by group 0 so OpenShift's arbitrary UID can write)
COPY --from=base --chown=65532:0 /tmp /tmp

WORKDIR /

COPY --chown=65532:0 bin/{{cookiecutter.app_name}}-$TARGETARCH {{cookiecutter.app_name}}

USER 65532:65532

ENTRYPOINT ["./{{cookiecutter.app_name}}"]
