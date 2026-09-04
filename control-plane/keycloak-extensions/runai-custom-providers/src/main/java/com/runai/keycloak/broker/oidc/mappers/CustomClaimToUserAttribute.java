package com.runai.keycloak.broker.oidc.mappers;

import com.fasterxml.jackson.databind.JsonNode;
import org.keycloak.broker.oidc.KeycloakOIDCIdentityProvider;
import org.keycloak.broker.oidc.OIDCIdentityProvider;
import org.keycloak.broker.oidc.mappers.AbstractJsonUserAttributeMapper;
import org.keycloak.broker.oidc.mappers.UserAttributeMapper;
import org.keycloak.broker.provider.BrokeredIdentityContext;
import org.keycloak.jose.jws.JWSInput;
import org.keycloak.jose.jws.JWSInputException;
import org.keycloak.models.IdentityProviderMapperModel;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.models.UserModel;
import org.keycloak.provider.ProviderConfigProperty;
import org.keycloak.representations.JsonWebToken;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.util.List;
import java.util.stream.Collectors;
import com.runai.keycloak.broker.oidc.mappers.MapperUtils;

import java.util.*;

// add supports for a claim in map format (not only list and plain text) - in function "toList"
// give priority to data coming from userinfo rather than data coming from id token - in function "getClaimValue"
public class CustomClaimToUserAttribute extends UserAttributeMapper {

    private static final Logger logger = LoggerFactory.getLogger(CustomClaimToUserAttribute.class);

    private static final List<ProviderConfigProperty> configProperties = new ArrayList<>();

    static {
        configProperties.add(MapperUtils.createConfigProperty(
            CLAIM,
            "Claim",
            "Name of claim to search for in token. You can reference nested claims using a '.', i.e. 'address.locality'. To use dot (.) literally, escape it with backslash (\\.))",
            ProviderConfigProperty.STRING_TYPE
        ));
        
        configProperties.add(MapperUtils.createConfigProperty(
            USER_ATTRIBUTE,
            "User Attribute Name",
            "User attribute name to store claim. support claims of format: string, list and map. data coming from userinfo will take priority on data coming form id token",
            ProviderConfigProperty.USER_PROFILE_ATTRIBUTE_LIST_TYPE
        ));
    }

    public static final String PROVIDER_ID = "runai-oidc-user-attribute-idp-mapper";

    @Override
    public List<ProviderConfigProperty> getConfigProperties() {
        return configProperties;
    }

    @Override
    public String getDisplayCategory() {
        return MapperUtils.RUNAI_ATTRIBUTE_IMPORTER_CATEGORY;
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }

    @Override
    public String getDisplayType() {
        return "Runai Attribute Importer (support plain text, list and map)";
    }

    @Override
    public void preprocessFederatedIdentity(KeycloakSession session, RealmModel realm, IdentityProviderMapperModel mapperModel, BrokeredIdentityContext context) {
        String attribute = mapperModel.getConfig().get(USER_ATTRIBUTE);
        if(attribute == null || attribute.isBlank()){
            return;
        }
        Object value = getClaimValue(mapperModel, context);
        List<String> values = toList(value);

        List<String> valuesToString = values.stream()
                .filter(Objects::nonNull)
                .map(Object::toString)
                .collect(Collectors.toList());

        context.setUserAttribute(attribute, valuesToString);
    }

    @Override
    public void updateBrokeredUser(KeycloakSession session, RealmModel realm, UserModel user, IdentityProviderMapperModel mapperModel, BrokeredIdentityContext context) {
        String attribute = mapperModel.getConfig().get(USER_ATTRIBUTE);
        if(attribute == null || attribute.isBlank()){
            return;
        }
        Object value = getClaimValue(mapperModel, context);
        List<String> values = toList(value);

        List<String> current = user.getAttributeStream(attribute).collect(Collectors.toList());
        if (!collectionEquals(values, current)) {
            user.setAttribute(attribute, values);
        } else if (values.isEmpty()) {
            user.removeAttribute(attribute);
        }
    }

    private static boolean collectionEquals(List<String> col1, List<String> col2) {
        if (col1.size() != col2.size()) {
            return false;
        }
        Map<String, Integer> countMap = new HashMap<>();
        for (String value : col1) {
            countMap.merge(value, 1, Integer::sum);
        }
        for (String value : col2) {
            Integer count = countMap.get(value);
            if (count == null) {
                return false;
            }
            if (count == 1) {
                countMap.remove(value);
            } else {
                countMap.put(value, count - 1);
            }
        }
        return countMap.isEmpty();
    }

    // to list is a private function in UserAttributeMapper and cant be overridden
    public List<String> toList(Object value) {
        List<Object> values;

        if (value instanceof List) {
            values = (List<Object>) value;
        } else if (value instanceof Map) {
            values = new ArrayList<>(((Map<?, ?>) value).values()); // Data coming from user info as a map
        } else if (value instanceof ObjectNode) { // data coming from user info inst parsed to a map but to ObjectNode
            ObjectNode objectNode = (ObjectNode) value;
            List<String> extractedValues = new ArrayList<>();
            objectNode.fields().forEachRemaining(entry -> extractedValues.add(entry.getValue().asText()));
            return extractedValues;
        } else {
            values = Collections.singletonList(value);
        }

        return values.stream()
                .filter(Objects::nonNull)
                .map(Object::toString)
                .collect(Collectors.toList());
    }

    public static Object getClaimValue(IdentityProviderMapperModel mapperModel, BrokeredIdentityContext context) {
        String claim = mapperModel.getConfig().get(CLAIM);
        return getClaimValue(context, claim);
    }

    public static Object getClaimValue(BrokeredIdentityContext context, String claim) {
        {
            // Search the OIDC UserInfo claim set (if any)
            JsonNode profileJsonNode = (JsonNode) context.getContextData().get(OIDCIdentityProvider.USER_INFO);
            Object value = AbstractJsonUserAttributeMapper.getJsonValue(profileJsonNode, claim);
            if (value != null) return value;
        }
        {  // search access token
            JsonWebToken token = (JsonWebToken)context.getContextData().get(KeycloakOIDCIdentityProvider.VALIDATED_ACCESS_TOKEN);
            if (token != null) {
                Object value = getClaimValue(token, claim);
                if (value != null) return value;
            }

        }
        {  // search ID Token
            Object rawIdToken = context.getContextData().get(OIDCIdentityProvider.VALIDATED_ID_TOKEN);
            JsonWebToken idToken = null;

            if (rawIdToken instanceof String) {
                try {
                    idToken = new JWSInput(rawIdToken.toString()).readJsonContent(JsonWebToken.class);
                } catch (JWSInputException e) {
                    logger.error("[CustomClaimToUserAttribute] Failed to parse ID token: {}", e.getMessage(), e);
                    return null;
                }
            } else if (rawIdToken instanceof JsonWebToken) {
                idToken = (JsonWebToken) rawIdToken;
            }

            if (idToken != null) {
                Object value = getClaimValue(idToken, claim);
                if (value != null)
                    return value;
            }
        }

        return null;
    }
}