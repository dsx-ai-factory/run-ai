FROM registry.access.redhat.com/ubi9/ubi-minimal AS base

# Create the user and group so the scratch image recognizes them
RUN echo "nonroot:x:65532:65532:nonroot:/home/nonroot:/sbin/nologin" > /etc/passwd && \
    echo "nonroot:x:65532:" > /etc/group

# Create a dedicated writable directory and set ownership
RUN mkdir -p /tmp && \
    chown -R 65532:65532 /tmp && \
    chmod 777 /tmp

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

# Writable Filesystem
COPY --from=base --chown=65532:65532 /tmp /tmp

WORKDIR /

COPY bin/{{cookiecutter.app_name}}-$TARGETARCH {{cookiecutter.app_name}}
COPY NOTICE .

USER 65532:65532

COPY pkg/database/migrations/up/* ./migrations/up/
COPY pkg/database/migrations/down/* ./migrations/down/
COPY swagger/_build/* ./swagger/

ENTRYPOINT ["./{{cookiecutter.app_name}}"]
