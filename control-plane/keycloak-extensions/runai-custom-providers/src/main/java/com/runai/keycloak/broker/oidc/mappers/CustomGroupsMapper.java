package com.runai.keycloak.broker.oidc.mappers;

import org.keycloak.models.ProtocolMapperModel;
import org.keycloak.models.UserSessionModel;
import org.keycloak.protocol.oidc.mappers.UserAttributeMapper;
import org.keycloak.protocol.oidc.mappers.OIDCAttributeMapperHelper;
import org.keycloak.provider.ProviderConfigProperty;
import org.keycloak.representations.IDToken;
import java.util.ArrayList;
import java.util.List;
import com.runai.keycloak.broker.oidc.mappers.MapperUtils;

public class CustomGroupsMapper extends UserAttributeMapper {
    public static final String PROVIDER_ID = "runai-oidc-custom-groups-mapper";
    private static final List<ProviderConfigProperty> configProperties = new ArrayList<>();
    private static final String DEFAULT_CLAIM_NAME = "custom_groups_mapper";
    private static final org.slf4j.Logger logger = org.slf4j.LoggerFactory.getLogger(CustomGroupsMapper.class);

    static {
        configProperties.add(MapperUtils.createConfigProperty(
            MapperUtils.TENANT_ID_ATTRIBUTE, "Tenant ID", "The tenant ID to use for custom group mapping.",
            ProviderConfigProperty.STRING_TYPE, true, null
        ));
        configProperties.add(MapperUtils.createConfigProperty(
            MapperUtils.GROUPS_USER_ATTRIBUTE, "Groups user attribute", "The user attribute to use for custom group mapping.",
            ProviderConfigProperty.STRING_TYPE, true, null
        ));
        configProperties.add(MapperUtils.createConfigProperty(
            MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE, "Groups Intersection Enabled", "If true, intersect local groups with external groups. Otherwise, use local groups as-is.",
            ProviderConfigProperty.BOOLEAN_TYPE, false, "true"
        ));
        OIDCAttributeMapperHelper.addTokenClaimNameConfig(configProperties);
        OIDCAttributeMapperHelper.addIncludeInTokensConfig(configProperties, CustomGroupsMapper.class);
    }

    @Override
    public String getDisplayCategory() {
        return MapperUtils.RUNAI_ATTRIBUTE_IMPORTER_CATEGORY;
    }

    @Override
    public String getDisplayType() {
        return "Custom Groups Mapper";
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
    protected void setClaim(IDToken token, ProtocolMapperModel mappingModel, UserSessionModel userSession) {
        List<String> localGroups = MapperUtils.getUserAttribute(mappingModel, userSession, MapperUtils.GROUPS_USER_ATTRIBUTE);
        int localGroupsSize = localGroups.size();
        boolean groupsIntersectionEnabled = Boolean.parseBoolean(mappingModel.getConfig().get(MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE));
        String claimName = getClaimName(mappingModel);
        if (localGroups.isEmpty()) {
            return;
        }
        List<String> resultGroups = new ArrayList<>();
        if (groupsIntersectionEnabled && localGroupsSize > MapperUtils.GROUPS_INTERSECTION_THRESHOLD) {
            String tenantId = mappingModel.getConfig().get(MapperUtils.TENANT_ID_ATTRIBUTE);
            List<String> externalGroups = MapperUtils.fetchGroupsFromAuthorizationService(tenantId, logger);
            if (!externalGroups.isEmpty()) {
                resultGroups = MapperUtils.intersect(localGroups, externalGroups);
            }
        } else {
            resultGroups = localGroups;
        }
        token.getOtherClaims().put(claimName, resultGroups);
    }

    @Override
    public String getHelpText() {
        return "Adds a custom claim to the token containing the user's group names (intersection or local).";
    }

    private static String getClaimName(ProtocolMapperModel mappingModel) {
        String claimName = mappingModel.getConfig().get(org.keycloak.protocol.oidc.mappers.OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME);
        if (claimName == null || claimName.isEmpty()) {
            claimName = DEFAULT_CLAIM_NAME;
        }
        return claimName;
    }
} 