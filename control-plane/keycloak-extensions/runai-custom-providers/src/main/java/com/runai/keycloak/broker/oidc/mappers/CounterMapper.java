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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CounterMapper extends UserAttributeMapper {
    public static final String PROVIDER_ID = "runai-counter-mapper";
    private static final Logger logger = LoggerFactory.getLogger(CounterMapper.class);
    private static final List<ProviderConfigProperty> configProperties = new ArrayList<>();
    private static final String DEFAULT_CLAIM_NAME = "groups_count";
    private static final String LIST_USER_ATTRIBUTE = "list.user.attribute";
    static {
        configProperties.add(MapperUtils.createConfigProperty(
            LIST_USER_ATTRIBUTE, "List user attribute", "The user attribute containing the list to be counted.",
            ProviderConfigProperty.STRING_TYPE, true, null
        ));
        OIDCAttributeMapperHelper.addTokenClaimNameConfig(configProperties);
        OIDCAttributeMapperHelper.addIncludeInTokensConfig(configProperties, CounterMapper.class);
    }

    @Override
    public String getDisplayCategory() {
        return MapperUtils.RUNAI_ATTRIBUTE_IMPORTER_CATEGORY;
    }

    @Override
    public String getDisplayType() {
        return "Counter Mapper";
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
        List<String> userAttributeList = MapperUtils.getUserAttribute(mappingModel, userSession, LIST_USER_ATTRIBUTE);
        int userAttributeListSize = userAttributeList.size();
        if (!userAttributeList.isEmpty()) {
            String claimName = getClaimName(mappingModel);
            token.getOtherClaims().put(claimName, userAttributeListSize);
        } else {
            logger.debug("[CounterMapper] User attribute list '{}' is empty or not found for user '{}'",
                    mappingModel.getConfig().get(LIST_USER_ATTRIBUTE),
                    userSession.getUser().getUsername());
        }
    }

    @Override
    public String getHelpText() {
        return "Adds a claim to the token containing the counter.";
    }

    private static String getClaimName(ProtocolMapperModel mappingModel) {
        String claimName = mappingModel.getConfig().get(org.keycloak.protocol.oidc.mappers.OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME);
        if (claimName == null || claimName.isEmpty()) {
            claimName = DEFAULT_CLAIM_NAME;
        }
        return claimName;
    }
}