package com.runai.keycloak.events.audit;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.keycloak.events.Event;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventType;
import org.keycloak.events.admin.AdminEvent;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.models.UserModel;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;

public class PasswordAuditEventListenerProvider implements EventListenerProvider {

    private static final Logger logger = LoggerFactory.getLogger(PasswordAuditEventListenerProvider.class);
    private static final ObjectMapper MAPPER = new ObjectMapper();

    static final String ACTION_FORGOT = "Password reset (Forgot password)";
    static final String ACTION_INVITE = "Password set (Invite link)";
    static final String SUBJECT_TYPE_USER = "user";
    static final String RESULT_SUCCEEDED = "Succeeded";

    static final String CREDENTIAL_TYPE_DETAIL = "credential_type";
    static final String CREDENTIAL_TYPE_PASSWORD = "password";
    static final String USERNAME_DETAIL = "username";
    static final String REDIRECT_URI_DETAIL = "redirect_uri";
    static final String FLOW_QUERY_PARAM = "runai_flow";
    static final String INVITE_FLOW_VALUE = "invite";
    static final String UNKNOWN_USER = "unknown";

    static final String AUDIT_LOG_PATH = "/internal/api/v1/audit/log";
    static final String TENANT_ID_ATTRIBUTE = "runai.tenant.id";

    static final String FIELD_TENANT_ID = "tenant_id";
    static final String FIELD_SUBJECT = "subject";
    static final String FIELD_SUBJECT_TYPE = "subject_type";
    static final String FIELD_ACTION = "action";
    static final String FIELD_RESULT = "result";

    private final KeycloakSession session;
    private final HttpClient httpClient;
    private final String auditServiceUrl;
    private final int timeoutMillis;

    public PasswordAuditEventListenerProvider(KeycloakSession session, HttpClient httpClient,
            String auditServiceUrl, int timeoutMillis) {
        this.session = session;
        this.httpClient = httpClient;
        this.auditServiceUrl = auditServiceUrl;
        this.timeoutMillis = timeoutMillis;
    }

    @Override
    public void onEvent(Event event) {
        if (!EventType.UPDATE_CREDENTIAL.equals(event.getType())) {
            return;
        }
        Map<String, String> details = event.getDetails() != null ? event.getDetails() : Map.of();
        if (!CREDENTIAL_TYPE_PASSWORD.equals(details.get(CREDENTIAL_TYPE_DETAIL))) {
            return;
        }

        try {
            String tenantId = resolveTenantId();
            if (tenantId == null) {
                logger.warn("[password-audit] skipping audit: could not resolve tenant for realm '{}'",
                        realmName());
                return;
            }
            String subject = resolveSubject(event, details);
            sendAuditLog(tenantId, subject, determineAction(details));
        } catch (Exception e) {
            logger.error("[password-audit] failed to create audit log: {}", e.getMessage(), e);
        }
    }

    static String determineAction(Map<String, String> details) {
        return isInviteRedirect(details.getOrDefault(REDIRECT_URI_DETAIL, "")) ? ACTION_INVITE : ACTION_FORGOT;
    }

    private static boolean isInviteRedirect(String redirectUri) {
        if (redirectUri == null || redirectUri.isBlank()) {
            return false;
        }
        String query;
        try {
            query = URI.create(redirectUri).getQuery();
        } catch (IllegalArgumentException e) {
            return false;
        }
        if (query == null) {
            return false;
        }
        for (String pair : query.split("&")) {
            int eq = pair.indexOf('=');
            if (eq > 0
                    && FLOW_QUERY_PARAM.equals(pair.substring(0, eq))
                    && INVITE_FLOW_VALUE.equals(pair.substring(eq + 1))) {
                return true;
            }
        }
        return false;
    }

    String resolveSubject(Event event, Map<String, String> details) {
        String username = details.get(USERNAME_DETAIL);
        if (username != null && !username.isBlank()) {
            return username;
        }
        RealmModel realm = session.getContext().getRealm();
        if (realm != null && event.getUserId() != null) {
            UserModel user = session.users().getUserById(realm, event.getUserId());
            if (user != null) {
                if (user.getEmail() != null && !user.getEmail().isBlank()) {
                    return user.getEmail();
                }
                if (user.getUsername() != null && !user.getUsername().isBlank()) {
                    return user.getUsername();
                }
            }
        }
        return event.getUserId() != null ? event.getUserId() : UNKNOWN_USER;
    }

    private String resolveTenantId() {
        RealmModel realm = session.getContext().getRealm();
        if (realm == null) {
            return null;
        }
        String tenantId = realm.getAttribute(TENANT_ID_ATTRIBUTE);
        return (tenantId != null && !tenantId.isBlank()) ? tenantId : null;
    }

    private void sendAuditLog(String tenantId, String subject, String action) throws Exception {
        Map<String, String> payload = Map.of(
                FIELD_TENANT_ID, tenantId,
                FIELD_SUBJECT, subject,
                FIELD_SUBJECT_TYPE, SUBJECT_TYPE_USER,
                FIELD_ACTION, action,
                FIELD_RESULT, RESULT_SUCCEEDED);
        HttpRequest req = HttpRequest.newBuilder(URI.create(auditServiceUrl + AUDIT_LOG_PATH))
                .timeout(Duration.ofMillis(timeoutMillis))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(MAPPER.writeValueAsString(payload)))
                .build();
        httpClient.sendAsync(req, HttpResponse.BodyHandlers.ofString())
                .whenComplete((resp, err) -> {
                    if (err != null) {
                        logger.error("[password-audit] audit POST failed: {}", err.getMessage());
                    } else if (resp.statusCode() < 200 || resp.statusCode() >= 300) {
                        logger.error("[password-audit] audit service returned status {}", resp.statusCode());
                    }
                });
    }

    private String realmName() {
        RealmModel realm = session.getContext().getRealm();
        return realm != null ? realm.getName() : null;
    }

    @Override
    public void onEvent(AdminEvent adminEvent, boolean includeRepresentation) {
    }

    @Override
    public void close() {
    }
}
