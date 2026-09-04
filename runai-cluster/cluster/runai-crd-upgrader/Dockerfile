FROM scratch

ARG TARGETARCH

COPY --chown=65532:0 bin/runai-crd-upgrader-$TARGETARCH /usr/bin/runai-crd-upgrader
COPY crds /crds
COPY NOTICE .

USER 65532:65532

ENTRYPOINT ["/usr/bin/runai-crd-upgrader"]
