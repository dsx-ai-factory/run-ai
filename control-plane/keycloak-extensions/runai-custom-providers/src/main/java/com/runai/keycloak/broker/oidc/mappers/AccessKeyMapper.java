package com.runai.keycloak.broker.oidc.mappers;

import org.keycloak.models.*;
import org.keycloak.protocol.oidc.mappers.AbstractOIDCProtocolMapper;
import org.keycloak.protocol.oidc.mappers.OIDCAccessTokenMapper;
import org.keycloak.protocol.oidc.mappers.OIDCIDTokenMapper;
import org.keycloak.protocol.oidc.mappers.UserInfoTokenMapper;
import org.keycloak.protocol.oidc.mappers.OIDCAttributeMapperHelper;
import org.keycloak.provider.ProviderConfigProperty;
import org.keycloak.representations.IDToken;

import java.util.ArrayList;
import java.util.List;
import com.runai.keycloak.broker.oidc.mappers.MapperUtils;

public class AccessKeyMapper extends AbstractOIDCProtocolMapper
    implements OIDCAccessTokenMapper, OIDCIDTokenMapper, UserInfoTokenMapper {

    public static final String PROVIDER_ID = "runai-access-key-attribute-mapper";
    private static final String DEFAULT_CLAIM_NAME = "user_app_groups";
    private static final org.slf4j.Logger logger = org.slf4j.LoggerFactory.getLogger(AccessKeyMapper.class);
    private static final List<ProviderConfigProperty> configProperties = new ArrayList<>();

    static {
        configProperties.add(MapperUtils.createConfigProperty(
            MapperUtils.TENANT_ID_ATTRIBUTE, "Tenant ID", "The tenant ID to use for access key attribute mapping.",
            ProviderConfigProperty.STRING_TYPE, false, null
        ));

        configProperties.add(MapperUtils.createConfigProperty(
            MapperUtils.GROUPS_USER_ATTRIBUTE, "User attribute", "The user attribute to copy from the real user (e.g. GROUPS, UID, GID, SUPPLEMENTARYGROUPS).",
            ProviderConfigProperty.STRING_TYPE, true, null
        ));

        configProperties.add(MapperUtils.createConfigProperty(
            MapperUtils.MULTIVALUED_ATTRIBUTE, "Multivalued", "If true, emit the claim as a list. Otherwise emit the first value as a scalar.",
            ProviderConfigProperty.BOOLEAN_TYPE, false, "true"
        ));

        configProperties.add(MapperUtils.createConfigProperty(
            MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE, "Groups Intersection Enabled", "If true, intersect local values with external groups from the authorization service. Otherwise, use local values as-is.",
            ProviderConfigProperty.BOOLEAN_TYPE, false, "true"
        ));

        OIDCAttributeMapperHelper.addTokenClaimNameConfig(configProperties);
        OIDCAttributeMapperHelper.addIncludeInTokensConfig(configProperties, AccessKeyMapper.class);
    }

    @Override
    public String getDisplayCategory() {
        return MapperUtils.RUNAI_ATTRIBUTE_IMPORTER_CATEGORY;
    }

    @Override
    public String getDisplayType() {
        return "Access Key Attribute Mapper";
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }

    @Override
    public List<ProviderConfigProperty> getConfigProperties() {
        return configProperties;
    }

    @Override
    public String getHelpText() {
        return "Maps a configurable attribute (groups, uid, gid, supplementarygroups, ...) for access key authentication by extracting it from the real user. " +
               "For service accounts, it uses the client's 'runai.created.by' attribute to find the actual user and their attribute. " +
               "Supports a configurable user attribute to copy (defaults to 'GROUPS'), emitting the claim as a list or a scalar, " +
               "and optional intersection with external groups from the authorization service.";
    }

    @Override
    protected void setClaim(IDToken token, ProtocolMapperModel mappingModel, UserSessionModel userSession,
                          KeycloakSession keycloakSession, ClientSessionContext clientSessionCtx) {

        try {
            UserModel user = userSession.getUser();
            ClientModel client = clientSessionCtx.getClientSession().getClient();

            if (client == null || user == null) {
                logger.warn("[AccessKeyMapper] CRITICAL: No client / userfound in session");
                return;
            }

            String username = user.getUsername();
            String claimName = getClaimName(mappingModel);
            boolean isServiceAccount = user.getServiceAccountClientLink() != null && username.startsWith("service-account-");

            List<String> resultValues = new ArrayList<>();
            if (!isServiceAccount) {
                logger.warn("[AccessKeyMapper] This mapper is only for service accounts, skipping user: {}", username);
                return;
            }

            RealmModel realm = userSession.getRealm();
            UserModel actualUser = null;

            String realUserIdentifier = client.getAttribute("runai.created.by");
            if (realUserIdentifier != null && !realUserIdentifier.isEmpty()) {
                actualUser = keycloakSession.users().getUserByUsername(realm, realUserIdentifier);
                if (actualUser == null) {
                    actualUser = keycloakSession.users().getUserByEmail(realm, realUserIdentifier);
                }
                if (actualUser != null) {
                    List<String> localValues = MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE);
                    if (!localValues.isEmpty()) {
                        resultValues = processWithIntersection(localValues, mappingModel);
                    }
                }
            } else {
                logger.warn("[AccessKeyMapper] No runai.created.by attribute found - cannot determine real user");
            }

            if (isMultivalued(mappingModel)) {
                token.getOtherClaims().put(claimName, resultValues);
            } else {
                token.getOtherClaims().put(claimName, resultValues.isEmpty() ? "" : resultValues.get(0));
            }
        } catch (Exception e) {
            logger.error("[AccessKeyMapper] Error in setClaim", e);
        }
    }

    private static boolean isMultivalued(ProtocolMapperModel mappingModel) {
        String multivalued = mappingModel.getConfig().get(MapperUtils.MULTIVALUED_ATTRIBUTE);
        return multivalued == null || Boolean.parseBoolean(multivalued);
    }

    private List<String> processWithIntersection(List<String> localValues, ProtocolMapperModel mappingModel) {
        int localValuesSize = localValues.size();
        String configuredAttribute = mappingModel.getConfig().get(MapperUtils.GROUPS_USER_ATTRIBUTE);
        boolean mapsGroups = configuredAttribute == null
            || configuredAttribute.isBlank()
            || MapperUtils.DEFAULT_LIST_ATTRIBUTE.equals(configuredAttribute);
        boolean intersectionEnabled = Boolean.parseBoolean(
            mappingModel.getConfig().get(MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE));
        List<String> result;
        if (mapsGroups && intersectionEnabled && localValuesSize > MapperUtils.GROUPS_INTERSECTION_THRESHOLD) {
            String tenantId = mappingModel.getConfig().get(MapperUtils.TENANT_ID_ATTRIBUTE);
            try {
                List<String> externalGroups = MapperUtils.fetchGroupsFromAuthorizationService(tenantId, logger);
                result = externalGroups.isEmpty() ? new ArrayList<>() : MapperUtils.intersect(localValues, externalGroups);
            } catch (Exception e) {
                logger.error("[AccessKeyMapper] Failed to fetch external groups, returning empty list", e);
                result = new ArrayList<>();
            }
        } else {
            result = new ArrayList<>(localValues);
        }

        return result;
    }

    private static String getClaimName(ProtocolMapperModel mappingModel) {
        String claimName = mappingModel.getConfig().get(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME);
        if (claimName == null || claimName.isEmpty()) {
            claimName = DEFAULT_CLAIM_NAME;
        }
        return claimName;
    }

    public static ProtocolMapperModel createAccessKeyMapper(String name, String claimName, String tenantId, String userAttribute, boolean multivalued, boolean intersectionEnabled) {
        ProtocolMapperModel mapper = new ProtocolMapperModel();
        mapper.setName(name);
        mapper.setProtocolMapper(PROVIDER_ID);
        mapper.setProtocol("openid-connect");
        mapper.setConfig(new java.util.HashMap<>());
        mapper.getConfig().put(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME, claimName);
        mapper.getConfig().put(MapperUtils.TENANT_ID_ATTRIBUTE, tenantId);
        mapper.getConfig().put(MapperUtils.GROUPS_USER_ATTRIBUTE, userAttribute);
        mapper.getConfig().put(MapperUtils.MULTIVALUED_ATTRIBUTE, String.valueOf(multivalued));
        mapper.getConfig().put(MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE, String.valueOf(intersectionEnabled));
        mapper.getConfig().put(OIDCAttributeMapperHelper.INCLUDE_IN_ACCESS_TOKEN, "true");
        mapper.getConfig().put(OIDCAttributeMapperHelper.INCLUDE_IN_ID_TOKEN, "true");
        mapper.getConfig().put(OIDCAttributeMapperHelper.INCLUDE_IN_USERINFO, "true");

        return mapper;
    }
}
