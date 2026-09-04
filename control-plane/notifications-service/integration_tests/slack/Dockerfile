FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY server.go .

RUN go build -o mock_server server.go

FROM alpine:latest

RUN addgroup -g 65532 nonroot && adduser -D -H -u 65532 -G nonroot nonroot

WORKDIR /app

COPY --from=builder --chown=65532:65532 /app/mock_server .

USER 65532:65532

EXPOSE 4010

CMD ["./mock_server"]
