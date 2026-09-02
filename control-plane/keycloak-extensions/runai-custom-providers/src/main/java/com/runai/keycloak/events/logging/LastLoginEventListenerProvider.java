package com.runai.keycloak.events.logging;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.keycloak.events.Event;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventType;
import org.keycloak.events.admin.AdminEvent;
import org.keycloak.models.*;

import java.util.List;
import java.util.Map;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.Optional;

public class LastLoginEventListenerProvider implements EventListenerProvider {

    private static final Logger logger = LoggerFactory.getLogger(LastLoginEventListenerProvider.class);

    private final KeycloakSession session;
    private final RealmProvider model;

    public LastLoginEventListenerProvider(KeycloakSession session) {
        this.session = session;
        this.model = session.realms();
    }

    @Override
    public void onEvent(Event event) {
        try {
            if (EventType.CLIENT_LOGIN.equals(event.getType())) {
                RealmModel realm = this.session.getContext().getRealm();
                ClientModel client = this.session.clients().getClientByClientId(realm, event.getClientId());
                if (client != null) {
                    client.setAttribute("runai.login.time", (System.currentTimeMillis() / 1000L) + "");
                } else {
                    logger.warn("[LastLoginEventListenerProvider] Client '{}' not found in realm, skipping login time update", event.getClientId());
                }
            }
            if (EventType.LOGIN.equals(event.getType())) {
                RealmModel realm = this.session.getContext().getRealm();
                UserModel user = this.session.users().getUserById(realm, event.getUserId());
                if (user != null) {
                    user.setSingleAttribute("runai.login.time", (System.currentTimeMillis() / 1000L) + "");
                } else {
                    logger.warn("[LastLoginEventListenerProvider] User '{}' not found in realm, skipping login time update", event.getUserId());
                }
            }
        } catch (Exception e) {
            logger.error("[LastLoginEventListenerProvider] Failed to record login time for event type '{}': {}", event.getType(), e.getMessage(), e);
        }
    }

    @Override
    public void onEvent(AdminEvent adminEvent, boolean b) {
    }

    @Override
    public void close() {
    }

}
