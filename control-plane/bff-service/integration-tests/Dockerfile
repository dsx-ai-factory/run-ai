FROM registry.access.redhat.com/ubi9/ubi-minimal
USER 65532:65532

LABEL description="Run integration tests for bff-service"


WORKDIR /tests
COPY bin/integ-tests-runner ./integ-tests
COPY integration-tests/run_tests.sh ./run_tests.sh

ENTRYPOINT ["/tests/run_tests.sh"]
