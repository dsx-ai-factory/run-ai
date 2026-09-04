package com.runai.keycloak.broker.oidc.mappers;

import org.keycloak.broker.oidc.KeycloakOIDCIdentityProviderFactory;
import org.keycloak.broker.oidc.OIDCIdentityProviderFactory;
import org.keycloak.broker.oidc.mappers.AbstractClaimMapper;
import org.keycloak.broker.provider.BrokeredIdentityContext;
import org.keycloak.models.IdentityProviderMapperModel;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.provider.ProviderConfigProperty;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import jakarta.ws.rs.WebApplicationException;
import jakarta.ws.rs.core.Response;
import java.util.*;

public class ClaimValidator extends AbstractClaimMapper {

    private static final Logger logger = LoggerFactory.getLogger(ClaimValidator.class);
    private static final List<ProviderConfigProperty> configProperties = new ArrayList<>();
    public static final String EXPECTED_CLAIM_VALUE = "claim.value";
    public static final String CLAIM_VALUE_SEPARATOR = "\\s";

    static {
        configProperties.add(MapperUtils.createConfigProperty(
            CLAIM,
            "Claim",
            "Name of claim to search for in token. You can reference nested claims using a '.', i.e. 'address.locality'. To use dot (.) literally, escape it with backslash (\\.))",
            ProviderConfigProperty.STRING_TYPE
        ));
        
        configProperties.add(MapperUtils.createConfigProperty(
            EXPECTED_CLAIM_VALUE,
            "Expected claim value",
            "Mandatory claim value options user must have one of the list in order to complete login",
            ProviderConfigProperty.STRING_TYPE
        ));
    }

    public static final String PROVIDER_ID = "runai-claim-validator";

    @Override
    public String[] getCompatibleProviders() {
        return new String[]{KeycloakOIDCIdentityProviderFactory.PROVIDER_ID, OIDCIdentityProviderFactory.PROVIDER_ID};
    }

    @Override
    public String getDisplayCategory() {
        return "Claim Validator";
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }

    @Override
    public String getHelpText() {
        return "";
    }

    @Override
    public List<ProviderConfigProperty> getConfigProperties() {
        return configProperties;
    }

    @Override
    public String getDisplayType() {
        return "Runai Claim Validator";
    }

    @Override
    public void preprocessFederatedIdentity(KeycloakSession session, RealmModel realm, IdentityProviderMapperModel mapperModel, BrokeredIdentityContext context) {
        Object value = getClaimValue(mapperModel, context);
        String expected_value = mapperModel.getConfig().get(EXPECTED_CLAIM_VALUE);
        validateClaim(value, expected_value);
    }

    public static void validateClaim(Object value, String expected_value) {
        if (!(value instanceof String) || (expected_value == null || expected_value.isBlank()) || !containsClaim(expected_value, (String) value)) {
            logger.warn("Invalid claim value: {} expected {}", value, expected_value);
            Response response = Response.status(Response.Status.UNAUTHORIZED)
                    .entity("Missing claim")
                    .build();
            throw new WebApplicationException(response);
        }
    }

    public static boolean containsClaim(String expected_value, String value) {
        if (expected_value == null || value == null) {
            return false;
        }

        String[] expected_list = expected_value.split(CLAIM_VALUE_SEPARATOR);
        for (String expected : expected_list) {
            if (expected.equals(value)) {
                return true;
            }
        }
        return false;
    }
}