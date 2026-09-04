package com.runai.keycloak.extensions.users;

import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.services.resources.admin.AdminEventBuilder;
import org.keycloak.services.resources.admin.ext.AdminRealmResourceProvider;
import org.keycloak.services.resources.admin.fgap.AdminPermissionEvaluator;

/**
 * Admin REST extension that exposes Run:ai's custom user-listing endpoint under the
 * standard admin API tree at {@code /admin/realms/{realm}/runaienteties/users}.
 *
 * <p>Implementing {@link AdminRealmResourceProvider} (rather than a plain
 * {@code RealmResourceProvider}) means Keycloak's admin REST layer performs the standard
 * admin authentication before {@link #getResource} is invoked and hands us the
 * {@link AdminPermissionEvaluator}. The returned sub-resource only has to enforce the
 * specific permission it needs.
 */
public class CustomUserResourceProvider implements AdminRealmResourceProvider {

    @Override
    public Object getResource(KeycloakSession session, RealmModel realm, AdminPermissionEvaluator auth, AdminEventBuilder adminEvent) {
        return new CustomUserResource(session, realm, auth);
    }

    @Override
    public void close() {
    }
}
