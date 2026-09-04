package com.runai.keycloak.events.audit;

import org.keycloak.Config;
import org.keycloak.events.EventListenerProviderFactory;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.KeycloakSessionFactory;

import java.net.http.HttpClient;
import java.time.Duration;

public class PasswordAuditEventListenerProviderFactory implements EventListenerProviderFactory {

    public static final String PROVIDER_ID = "password-audit-event";

    static final String AUDIT_SERVICE_URL_ENV = "KC_RUNAI_AUDIT_SERVICE_URL";
    static final String DEFAULT_AUDIT_SERVICE_URL = "http://runai-backend-audit-service:8080";
    static final int TIMEOUT_MILLIS = 3000;

    private HttpClient httpClient;
    private String auditServiceUrl;

    @Override
    public PasswordAuditEventListenerProvider create(KeycloakSession session) {
        return new PasswordAuditEventListenerProvider(session, httpClient, auditServiceUrl, TIMEOUT_MILLIS);
    }

    @Override
    public void init(Config.Scope scope) {
        this.auditServiceUrl = resolveConfig(scope, "auditServiceUrl", AUDIT_SERVICE_URL_ENV, DEFAULT_AUDIT_SERVICE_URL);
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofMillis(TIMEOUT_MILLIS))
                .build();
    }

    private static String resolveConfig(Config.Scope scope, String scopeKey, String envVar, String defaultValue) {
        if (scope != null) {
            String value = scope.get(scopeKey);
            if (value != null && !value.isBlank()) {
                return value;
            }
        }
        String env = System.getenv(envVar);
        return (env != null && !env.isEmpty()) ? env : defaultValue;
    }

    @Override
    public void postInit(KeycloakSessionFactory factory) {
    }

    @Override
    public void close() {
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }
}
