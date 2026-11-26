# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM registry.access.redhat.com/ubi9/ubi-minimal
USER 65532:65532

LABEL description="Run integration tests for cluster-service"


WORKDIR /tests
COPY bin/integ-tests ./integ-tests
COPY integration-tests/run_tests.sh ./run_tests.sh

ENTRYPOINT ["/tests/run_tests.sh"]
