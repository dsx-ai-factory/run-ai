FROM docker.io/library/alpine:3.21

RUN mkdir -p /extensions/providers /extensions/themes /extensions/agent /providers /themes /agent /scripts && \
    chown -R 65532:65532 /extensions /providers /themes /agent /scripts

COPY bin/runai-custom-providers.jar /extensions/providers/
COPY bin/opentelemetry-javaagent.jar /extensions/agent/
COPY runai-theme/ /extensions/themes/runai-theme/
COPY NOTICE /extensions/
COPY scripts/ /scripts/
RUN chmod +x /scripts/*

USER 65532:65532

CMD ["sh", "-c", "sh /scripts/copy-extensions.sh && sh /scripts/write-keycloak-conf.sh"]
