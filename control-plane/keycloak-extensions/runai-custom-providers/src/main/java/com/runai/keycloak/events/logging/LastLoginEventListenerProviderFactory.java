package com.runai.keycloak.events.logging;

import org.keycloak.Config;
import org.keycloak.events.EventListenerProviderFactory;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.KeycloakSessionFactory;

public class LastLoginEventListenerProviderFactory implements EventListenerProviderFactory {

    @Override
    public LastLoginEventListenerProvider create(KeycloakSession keycloakSession) {
        return new LastLoginEventListenerProvider(keycloakSession);
    }

    @Override
    public void init(Config.Scope scope) {
        //
    }

    @Override
    public void postInit(KeycloakSessionFactory keycloakSessionFactory) {
        //
    }

    @Override
    public void close() {
        //
    }

    @Override
    public String getId() {
        return "last-login-event";
    }

}
