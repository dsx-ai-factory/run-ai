FROM registry.access.redhat.com/ubi9/ubi-minimal
USER 65532:65532

LABEL description="Run integration tests for tenants-manager"


WORKDIR /tests
COPY bin/integ-tests ./integ-tests
COPY integration-tests/run_tests.sh ./run_tests.sh

ENTRYPOINT ["/tests/run_tests.sh"]
