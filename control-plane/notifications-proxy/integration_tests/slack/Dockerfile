# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.26.8-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o mock_server ./...

FROM alpine:3.23

RUN addgroup -g 65532 nonroot && adduser -D -H -u 65532 -G nonroot nonroot

WORKDIR /app

COPY --from=builder --chown=65532:65532 /app/mock_server .

USER 65532:65532

EXPOSE 8081

CMD ["./mock_server"]
