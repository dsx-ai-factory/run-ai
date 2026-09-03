FROM registry.access.redhat.com/ubi9/ubi-minimal

WORKDIR /tests
COPY bin/integ-tests ./integ-tests

CMD ["/tests/integ-tests", "-test.v"]
