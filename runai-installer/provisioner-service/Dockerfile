FROM registry.access.redhat.com/ubi9/ubi-minimal
ARG TARGETARCH
ENV TARGETARCH=$TARGETARCH
ENV HOME=/home/nonroot
WORKDIR /

RUN mkdir -p /home/nonroot && chown 65532:65532 /home/nonroot

COPY bin/provisioner-service-$TARGETARCH provisioner-service
COPY NOTICE .

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["./provisioner-service"]
