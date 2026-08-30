package com.runai.keycloak.extensions.users;

import org.keycloak.Config;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.KeycloakSessionFactory;
import org.keycloak.services.resources.admin.ext.AdminRealmResourceProvider;
import org.keycloak.services.resources.admin.ext.AdminRealmResourceProviderFactory;

public class CustomUserResourceProviderFactory implements AdminRealmResourceProviderFactory {

    // Path segment under /admin/realms/{realm}/ that this resource is mounted at.
    public static final String PROVIDER_ID = "runaienteties";

    @Override
    public AdminRealmResourceProvider create(KeycloakSession session) {
        return new CustomUserResourceProvider();
    }

    @Override
    public void init(Config.Scope config) {
        // Initialize if needed
    }

    @Override
    public void postInit(KeycloakSessionFactory factory) {
        // Post initialization if needed
    }

    @Override
    public void close() {
        // Clean up resources if needed
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }
}
