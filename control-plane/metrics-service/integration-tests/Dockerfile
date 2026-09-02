FROM registry.access.redhat.com/ubi9/ubi-minimal
USER 65532:65532

LABEL description="Run integration tests for metrics-service"


WORKDIR /tests
COPY bin/integ-tests ./integ-tests

ENTRYPOINT ["/tests/integ-tests"]
