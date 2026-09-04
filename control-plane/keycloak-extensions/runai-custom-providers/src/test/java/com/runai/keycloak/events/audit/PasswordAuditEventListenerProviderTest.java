package com.runai.keycloak.events.audit;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.io.IOException;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Flow;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.keycloak.events.Event;
import org.keycloak.events.EventType;
import org.keycloak.models.KeycloakContext;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.models.UserModel;
import org.keycloak.models.UserProvider;

import com.fasterxml.jackson.databind.ObjectMapper;

class PasswordAuditEventListenerProviderTest {

    private static final ObjectMapper MAPPER = new ObjectMapper();
    private static final String AUDIT_URL = "http://audit:8080";
    private static final String REALM = "runai";
    private static final String TENANT_ID = "1001";

    private KeycloakSession session;
    private RealmModel realm;
    private HttpClient httpClient;
    private PasswordAuditEventListenerProvider provider;

    private final AtomicReference<HttpRequest> auditRequest = new AtomicReference<>();

    @BeforeEach
    void setUp() {
        session = mock(KeycloakSession.class);
        KeycloakContext ctx = mock(KeycloakContext.class);
        realm = mock(RealmModel.class);
        when(session.getContext()).thenReturn(ctx);
        when(ctx.getRealm()).thenReturn(realm);
        when(realm.getName()).thenReturn(REALM);
        when(realm.getAttribute("runai.tenant.id")).thenReturn(TENANT_ID);

        httpClient = mock(HttpClient.class);
        provider = new PasswordAuditEventListenerProvider(session, httpClient, AUDIT_URL, 3000);
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    private void stubHttp() throws Exception {
        when(httpClient.sendAsync(any(HttpRequest.class), any())).thenAnswer(inv -> {
            HttpRequest req = inv.getArgument(0);
            String uri = req.uri().toString();
            if (uri.contains("/audit/log")) {
                auditRequest.set(req);
                return CompletableFuture.completedFuture(response(200, ""));
            }
            throw new IllegalStateException("unexpected request URI: " + uri);
        });
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    private static HttpResponse<String> response(int status, String body) {
        HttpResponse<String> resp = mock(HttpResponse.class);
        when(resp.statusCode()).thenReturn(status);
        when(resp.body()).thenReturn(body);
        return resp;
    }

    private static Event passwordEvent(String redirectUri, String username) {
        Map<String, String> details = new HashMap<>();
        details.put("credential_type", "password");
        if (redirectUri != null) {
            details.put("redirect_uri", redirectUri);
        }
        if (username != null) {
            details.put("username", username);
        }
        Event event = mock(Event.class);
        when(event.getType()).thenReturn(EventType.UPDATE_CREDENTIAL);
        when(event.getDetails()).thenReturn(details);
        when(event.getUserId()).thenReturn("uid-1");
        return event;
    }

    private Map<String, String> capturedAuditPayload() throws Exception {
        HttpRequest req = auditRequest.get();
        assertThat(req).as("an audit POST should have been sent").isNotNull();
        return MAPPER.readValue(readBody(req), Map.class);
    }

    // --- filtering -------------------------------------------------------------------------------

    @Test
    void ignoresNonCredentialEvents() throws Exception {
        Event event = mock(Event.class);
        when(event.getType()).thenReturn(EventType.LOGIN);

        provider.onEvent(event);

        verify(httpClient, never()).sendAsync(any(), any());
    }

    @Test
    void ignoresNonPasswordCredential() throws Exception {
        Map<String, String> details = new HashMap<>();
        details.put("credential_type", "otp");
        Event event = mock(Event.class);
        when(event.getType()).thenReturn(EventType.UPDATE_CREDENTIAL);
        when(event.getDetails()).thenReturn(details);

        provider.onEvent(event);

        verify(httpClient, never()).sendAsync(any(), any());
    }

    // --- determineAction (pure, no I/O) ----------------------------------------------------------

    @Test
    void determineActionInviteWhenFlowMarkerPresent() {
        assertThat(PasswordAuditEventListenerProvider.determineAction(
                Map.of("redirect_uri", "https://host/login/callback?runai_flow=invite")))
                .isEqualTo("Password set (Invite link)");
    }

    @Test
    void determineActionInviteWhenFlowMarkerAmongOtherParams() {
        assertThat(PasswordAuditEventListenerProvider.determineAction(
                Map.of("redirect_uri", "https://host/login/callback?state=abc&runai_flow=invite&code=x")))
                .isEqualTo("Password set (Invite link)");
    }

    @Test
    void determineActionForgotForPlainCallback() {
        assertThat(PasswordAuditEventListenerProvider.determineAction(
                Map.of("redirect_uri", "https://host/login/callback")))
                .isEqualTo("Password reset (Forgot password)");
    }

    @Test
    void determineActionForgotWhenQueryHasNoFlowMarker() {
        assertThat(PasswordAuditEventListenerProvider.determineAction(
                Map.of("redirect_uri", "https://host/login/callback?state=abc&code=x")))
                .isEqualTo("Password reset (Forgot password)");
    }

    @Test
    void determineActionForgotWhenRedirectUriAbsent() {
        assertThat(PasswordAuditEventListenerProvider.determineAction(Map.of()))
                .isEqualTo("Password reset (Forgot password)");
    }

    @Test
    void determineActionForgotForLegacyUidPathWithoutMarker() {
        assertThat(PasswordAuditEventListenerProvider.determineAction(
                Map.of("redirect_uri", "https://host/login/callback/OTk0OTRhNmE")))
                .isEqualTo("Password reset (Forgot password)");
    }

    @Test
    void determineActionForgotWhenFlowValueIsNotInvite() {
        assertThat(PasswordAuditEventListenerProvider.determineAction(
                Map.of("redirect_uri", "https://host/login/callback?runai_flow=welcome")))
                .isEqualTo("Password reset (Forgot password)");
    }

    // --- resolveSubject --------------------------------------------------------------------------

    @Test
    void resolveSubjectPrefersUsernameDetailWithoutTouchingSession() {
        Event event = mock(Event.class);
        String subject = provider.resolveSubject(event, Map.of("username", "detail@example.com"));

        assertThat(subject).isEqualTo("detail@example.com");
        verify(session, never()).users();
    }

    // --- invite vs forgot discriminator (end-to-end through onEvent) -----------------------------

    @Test
    void inviteRedirectIsLabeledAsInvite() throws Exception {
        stubHttp();
        provider.onEvent(passwordEvent(
                "https://host/login/callback?runai_flow=invite", "invitee@example.com"));

        Map<String, String> payload = capturedAuditPayload();
        assertThat(payload.get(PasswordAuditEventListenerProvider.FIELD_ACTION))
                .isEqualTo("Password set (Invite link)");
        assertThat(payload.get(PasswordAuditEventListenerProvider.FIELD_SUBJECT))
                .isEqualTo("invitee@example.com");
        assertThat(payload.get(PasswordAuditEventListenerProvider.FIELD_SUBJECT_TYPE))
                .isEqualTo("user");
        assertThat(payload.get(PasswordAuditEventListenerProvider.FIELD_RESULT))
                .isEqualTo("Succeeded");
        assertThat(payload.get(PasswordAuditEventListenerProvider.FIELD_TENANT_ID))
                .isEqualTo("1001");
    }

    @Test
    void plainCallbackIsLabeledAsForgot() throws Exception {
        stubHttp();
        provider.onEvent(passwordEvent(
                "https://host/login/callback", "user@example.com"));

        assertThat(capturedAuditPayload().get(PasswordAuditEventListenerProvider.FIELD_ACTION)).isEqualTo("Password reset (Forgot password)");
    }

    @Test
    void missingRedirectUriIsLabeledAsForgot() throws Exception {
        stubHttp();
        provider.onEvent(passwordEvent(null, "user@example.com"));

        assertThat(capturedAuditPayload().get(PasswordAuditEventListenerProvider.FIELD_ACTION)).isEqualTo("Password reset (Forgot password)");
    }

    // --- subject resolution -----------------------------------------------------------------------

    @Test
    void fallsBackToUserEmailWhenNoUsernameDetail() throws Exception {
        stubHttp();
        UserModel user = mock(UserModel.class);
        when(user.getEmail()).thenReturn("fallback@example.com");
        UserProvider users = mock(UserProvider.class);
        when(session.users()).thenReturn(users);
        when(users.getUserById(any(RealmModel.class), any())).thenReturn(user);

        provider.onEvent(passwordEvent("https://host/login/callback", null));

        assertThat(capturedAuditPayload().get(PasswordAuditEventListenerProvider.FIELD_SUBJECT)).isEqualTo("fallback@example.com");
    }

    // --- tenant resolution / best-effort ----------------------------------------------------------

    @Test
    void skipsAuditWhenTenantAttributeMissing() throws Exception {
        when(realm.getAttribute("runai.tenant.id")).thenReturn(null);

        provider.onEvent(passwordEvent("https://host/login/callback", "user@example.com"));

        verify(httpClient, never()).sendAsync(any(), any());
    }

    @Test
    void neverThrowsWhenHttpFails() throws Exception {
        when(httpClient.sendAsync(any(HttpRequest.class), any()))
                .thenAnswer(inv -> CompletableFuture.failedFuture(new IOException("audit-service unreachable")));

        provider.onEvent(passwordEvent("https://host/login/callback", "user@example.com"));
    }

    // --- helpers ----------------------------------------------------------------------------------

    private static String readBody(HttpRequest request) throws Exception {
        HttpRequest.BodyPublisher publisher = request.bodyPublisher().orElseThrow();
        StringBuilder sb = new StringBuilder();
        AtomicReference<Throwable> error = new AtomicReference<>();
        CountDownLatch latch = new CountDownLatch(1);
        publisher.subscribe(new Flow.Subscriber<ByteBuffer>() {
            @Override
            public void onSubscribe(Flow.Subscription subscription) {
                subscription.request(Long.MAX_VALUE);
            }

            @Override
            public void onNext(ByteBuffer item) {
                sb.append(StandardCharsets.UTF_8.decode(item));
            }

            @Override
            public void onError(Throwable throwable) {
                error.set(throwable);
                latch.countDown();
            }

            @Override
            public void onComplete() {
                latch.countDown();
            }
        });
        if (!latch.await(5, TimeUnit.SECONDS)) {
            throw new IllegalStateException("timed out reading request body");
        }
        if (error.get() != null) {
            throw new RuntimeException(error.get());
        }
        return sb.toString();
    }
}
