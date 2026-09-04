FROM registry.access.redhat.com/ubi9/ubi-minimal
ARG TARGETARCH
ENV TARGETARCH=$TARGETARCH

RUN microdnf update -y && microdnf install -y gcc ca-certificates openssl git fuse rsyslog

COPY bin/goofys-$TARGETARCH /usr/local/bin/goofys
RUN chmod +x /usr/local/bin/goofys
RUN curl -sSL -o /usr/local/bin/catfs https://github.com/kahing/catfs/releases/download/v0.9.0/catfs && chmod +x /usr/local/bin/catfs

ARG ENDPOINT
ENV MOUNT_DIR /mnt/s3
ENV REGION us-east-1
ENV BUCKET teleport-bucket
ENV STAT_CACHE_TTL 1m0s
ENV TYPE_CACHE_TTL 1m0s
ENV DIR_MODE 0700
ENV FILE_MODE 0600
ENV UID 0
ENV GID 0

RUN mkdir /mnt/s3

VOLUME /mnt/s3

COPY rootfs/ /

RUN chmod +x /usr/bin/run.sh

COPY NOTICE .

ENTRYPOINT ["sh"]
CMD ["/usr/bin/run.sh"]
