FROM registry.access.redhat.com/ubi9/ubi-minimal AS base

# Wrap /tmp inside /temp so /tmp lands as a *child* during COPY. Docker BuildKit resets the
# mode of any destination directory it creates to 0755, but preserves modes on children.
RUN mkdir -p /temp/tmp && \
    chown 65532:0 /temp/tmp && \
    chmod 770 /temp/tmp

FROM scratch

ARG TARGETARCH
ENV TARGETARCH=$TARGETARCH

COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group
COPY --from=base /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=base /etc/pki/ /etc/pki/
COPY --from=base /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=base /temp /

WORKDIR /

COPY bin/gateway-hub-$TARGETARCH gateway-hub
COPY NOTICE /NOTICE

USER 65532:65532

EXPOSE 8181

ENTRYPOINT ["./gateway-hub"]
