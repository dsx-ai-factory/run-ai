package com.runai.keycloak.broker.oidc.mappers;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.keycloak.models.*;
import org.keycloak.protocol.oidc.mappers.OIDCAttributeMapperHelper;
import org.keycloak.representations.IDToken;
import org.mockito.MockedStatic;

import java.util.*;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;
import static org.mockito.ArgumentMatchers.*;

import com.runai.keycloak.broker.oidc.mappers.MapperUtils;

class AccessKeyMapperTest {

    private AccessKeyMapper mapper;
    private ProtocolMapperModel mappingModel;
    private UserSessionModel userSession;
    private KeycloakSession keycloakSession;
    private ClientSessionContext clientSessionCtx;
    private AuthenticatedClientSessionModel clientSession;
    private ClientModel client;
    private UserModel serviceAccountUser;
    private UserModel actualUser;
    private RealmModel realm;
    private UserProvider userProvider;
    private IDToken token;
    private Map<String, String> config;

    @BeforeEach
    void setUp() {
        mapper = new AccessKeyMapper();
        mappingModel = mock(ProtocolMapperModel.class);
        userSession = mock(UserSessionModel.class);
        keycloakSession = mock(KeycloakSession.class);
        clientSessionCtx = mock(ClientSessionContext.class);
        clientSession = mock(AuthenticatedClientSessionModel.class);
        client = mock(ClientModel.class);
        serviceAccountUser = mock(UserModel.class);
        actualUser = mock(UserModel.class);
        realm = mock(RealmModel.class);
        userProvider = mock(UserProvider.class);
        token = new IDToken();

        when(clientSessionCtx.getClientSession()).thenReturn(clientSession);
        when(clientSession.getClient()).thenReturn(client);
        when(userSession.getUser()).thenReturn(serviceAccountUser);
        when(userSession.getRealm()).thenReturn(realm);
        when(keycloakSession.users()).thenReturn(userProvider);

        when(serviceAccountUser.getServiceAccountClientLink()).thenReturn("some-client-id");
        when(serviceAccountUser.getUsername()).thenReturn("service-account-test");

        config = new HashMap<>();
        config.put(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME, "user_app_groups");
        config.put(MapperUtils.GROUPS_USER_ATTRIBUTE, "GROUPS");
        config.put(MapperUtils.MULTIVALUED_ATTRIBUTE, "true");
        config.put(MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE, "false");
        when(mappingModel.getConfig()).thenReturn(config);
    }

    @Test
    void testGetDisplayCategory() {
        assertThat(mapper.getDisplayCategory()).isEqualTo(MapperUtils.RUNAI_ATTRIBUTE_IMPORTER_CATEGORY);
    }

    @Test
    void testGetDisplayType() {
        assertThat(mapper.getDisplayType()).isEqualTo("Access Key Attribute Mapper");
    }

    @Test
    void testGetId() {
        assertThat(mapper.getId()).isEqualTo("runai-access-key-attribute-mapper");
    }

    @Test
    void testCreateAccessKeyMapper() {
        ProtocolMapperModel created = AccessKeyMapper.createAccessKeyMapper("test-mapper", "test_claim", "test-tenant", "UID", false, false);

        assertThat(created.getName()).isEqualTo("test-mapper");
        assertThat(created.getProtocolMapper()).isEqualTo("runai-access-key-attribute-mapper");
        assertThat(created.getProtocol()).isEqualTo("openid-connect");
        assertThat(created.getConfig().get(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME)).isEqualTo("test_claim");
        assertThat(created.getConfig().get(MapperUtils.TENANT_ID_ATTRIBUTE)).isEqualTo("test-tenant");
        assertThat(created.getConfig().get(MapperUtils.GROUPS_USER_ATTRIBUTE)).isEqualTo("UID");
        assertThat(created.getConfig().get(MapperUtils.MULTIVALUED_ATTRIBUTE)).isEqualTo("false");
        assertThat(created.getConfig().get(MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE)).isEqualTo("false");
        assertThat(created.getConfig().get(OIDCAttributeMapperHelper.INCLUDE_IN_ACCESS_TOKEN)).isEqualTo("true");
        assertThat(created.getConfig().get(OIDCAttributeMapperHelper.INCLUDE_IN_ID_TOKEN)).isEqualTo("true");
        assertThat(created.getConfig().get(OIDCAttributeMapperHelper.INCLUDE_IN_USERINFO)).isEqualTo("true");
    }

    @Test
    void testConfigProperties() {
        List<org.keycloak.provider.ProviderConfigProperty> properties = mapper.getConfigProperties();

        assertThat(properties).isNotNull().isNotEmpty();
        assertThat(properties).anyMatch(p -> MapperUtils.TENANT_ID_ATTRIBUTE.equals(p.getName()));
        assertThat(properties).anyMatch(p -> MapperUtils.GROUPS_USER_ATTRIBUTE.equals(p.getName()));
        assertThat(properties).anyMatch(p -> MapperUtils.MULTIVALUED_ATTRIBUTE.equals(p.getName()));
        assertThat(properties).anyMatch(p -> MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE.equals(p.getName()));
    }

    @Test
    void testHelpText() {
        assertThat(mapper.getHelpText())
            .isNotNull()
            .contains("access key authentication")
            .contains("service accounts");
    }

    @Test
    void testUserFoundByUsername_NoFallbackNeeded() {
        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "testuser";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(actualUser);

            List<String> userGroups = Arrays.asList("group1", "group2");
            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(userGroups);

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("user_app_groups")).isEqualTo(userGroups);
            verify(userProvider).getUserByUsername(realm, userIdentifier);
            verify(userProvider, never()).getUserByEmail(realm, userIdentifier);
        }
    }

    @Test
    void testUserNotFoundByUsername_FallbackToEmailSucceeds() {
        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "testuser";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(null);
            when(userProvider.getUserByEmail(realm, userIdentifier)).thenReturn(actualUser);

            List<String> userGroups = Arrays.asList("group1", "group2");
            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(userGroups);

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("user_app_groups")).isEqualTo(userGroups);
            verify(userProvider).getUserByUsername(realm, userIdentifier);
            verify(userProvider).getUserByEmail(realm, userIdentifier);
        }
    }

    @Test
    void testUserNotFoundByUsernameOrEmail_ReturnsEmptyList() {
        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "nonexistentuser";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(null);
            when(userProvider.getUserByEmail(realm, userIdentifier)).thenReturn(null);

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat((List<String>) token.getOtherClaims().get("user_app_groups")).isNotNull().isEmpty();
            verify(userProvider).getUserByUsername(realm, userIdentifier);
            verify(userProvider).getUserByEmail(realm, userIdentifier);
        }
    }

    @Test
    void testEmailAsIdentifier_FallbackBehavior() {
        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String emailIdentifier = "test@example.com";
            when(client.getAttribute("runai.created.by")).thenReturn(emailIdentifier);
            when(userProvider.getUserByUsername(realm, emailIdentifier)).thenReturn(null);
            when(userProvider.getUserByEmail(realm, emailIdentifier)).thenReturn(actualUser);

            List<String> userGroups = Arrays.asList("group1", "group2");
            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(userGroups);

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("user_app_groups")).isEqualTo(userGroups);
            verify(userProvider).getUserByUsername(realm, emailIdentifier);
            verify(userProvider).getUserByEmail(realm, emailIdentifier);
        }
    }

    @Test
    void testUsernameAsIdentifier_DirectMatch() {
        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String usernameIdentifier = "johndoe";
            when(client.getAttribute("runai.created.by")).thenReturn(usernameIdentifier);
            when(userProvider.getUserByUsername(realm, usernameIdentifier)).thenReturn(actualUser);

            List<String> userGroups = Arrays.asList("admin", "users");
            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(userGroups);

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("user_app_groups")).isEqualTo(userGroups);
            verify(userProvider).getUserByUsername(realm, usernameIdentifier);
            verify(userProvider, never()).getUserByEmail(realm, usernameIdentifier);
        }
    }

    @Test
    void testNoRunaiCreatedByAttribute_ReturnsEmptyList() {
        when(client.getAttribute("runai.created.by")).thenReturn(null);

        mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

        assertThat((List<String>) token.getOtherClaims().get("user_app_groups")).isNotNull().isEmpty();
        verify(userProvider, never()).getUserByUsername(any(), any());
        verify(userProvider, never()).getUserByEmail(any(), any());
    }

    @Test
    void testEmptyRunaiCreatedByAttribute_ReturnsEmptyList() {
        when(client.getAttribute("runai.created.by")).thenReturn("");

        mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

        assertThat((List<String>) token.getOtherClaims().get("user_app_groups")).isNotNull().isEmpty();
        verify(userProvider, never()).getUserByUsername(any(), any());
        verify(userProvider, never()).getUserByEmail(any(), any());
    }

    @Test
    void testNonServiceAccount_SkipsProcessing() {
        when(serviceAccountUser.getServiceAccountClientLink()).thenReturn(null);
        when(serviceAccountUser.getUsername()).thenReturn("regularuser");

        mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

        assertThat(token.getOtherClaims()).doesNotContainKey("user_app_groups");
        verify(userProvider, never()).getUserByUsername(any(), any());
        verify(userProvider, never()).getUserByEmail(any(), any());
    }

    @Test
    void testUserFoundButNoAttribute_ReturnsEmptyList() {
        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "userwithnogroups";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(actualUser);

            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(Collections.emptyList());

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat((List<String>) token.getOtherClaims().get("user_app_groups")).isNotNull().isEmpty();
            verify(userProvider).getUserByUsername(realm, userIdentifier);
            verify(userProvider, never()).getUserByEmail(realm, userIdentifier);
        }
    }

    @Test
    void testScalarMode_EmitsFirstValueAsString() {
        config.put(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME, "uid");
        config.put(MapperUtils.GROUPS_USER_ATTRIBUTE, "UID");
        config.put(MapperUtils.MULTIVALUED_ATTRIBUTE, "false");

        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "testuser";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(actualUser);

            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(Collections.singletonList("1000"));

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("uid")).isEqualTo("1000");
        }
    }

    @Test
    void testScalarMode_NoValueEmitsEmptyString() {
        config.put(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME, "uid");
        config.put(MapperUtils.GROUPS_USER_ATTRIBUTE, "UID");
        config.put(MapperUtils.MULTIVALUED_ATTRIBUTE, "false");

        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "testuser";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(actualUser);

            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(Collections.emptyList());

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("uid")).isEqualTo("");
        }
    }

    @Test
    void testMultivaluedMode_EmitsList() {
        config.put(MapperUtils.MULTIVALUED_ATTRIBUTE, "true");

        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "testuser";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(actualUser);

            List<String> supGroups = Arrays.asList("100", "200");
            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(supGroups);

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("user_app_groups")).isEqualTo(supGroups);
        }
    }

    @Test
    void testNonGroupAttributeWithIntersectionEnabled_DoesNotIntersectWithExternalGroups() {
        config.put(OIDCAttributeMapperHelper.TOKEN_CLAIM_NAME, "supplementarygroups");
        config.put(MapperUtils.GROUPS_USER_ATTRIBUTE, "SUPPLEMENTARYGROUPS");
        config.put(MapperUtils.MULTIVALUED_ATTRIBUTE, "true");
        config.put(MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE, "true");

        try (MockedStatic<MapperUtils> mockedUtils = mockStatic(MapperUtils.class)) {
            String userIdentifier = "testuser";
            when(client.getAttribute("runai.created.by")).thenReturn(userIdentifier);
            when(userProvider.getUserByUsername(realm, userIdentifier)).thenReturn(actualUser);

            List<String> supplementaryGroups = new ArrayList<>();
            for (int i = 0; i < MapperUtils.GROUPS_INTERSECTION_THRESHOLD + 1; i++) {
                supplementaryGroups.add(String.valueOf(1000 + i));
            }
            mockedUtils.when(() -> MapperUtils.getUserAttribute(actualUser, mappingModel, MapperUtils.GROUPS_USER_ATTRIBUTE))
                      .thenReturn(supplementaryGroups);
            mockedUtils.when(() -> MapperUtils.intersect(any(), any())).thenCallRealMethod();

            mapper.setClaim(token, mappingModel, userSession, keycloakSession, clientSessionCtx);

            assertThat(token.getOtherClaims().get("supplementarygroups")).isEqualTo(supplementaryGroups);
            mockedUtils.verify(() -> MapperUtils.fetchGroupsFromAuthorizationService(any(), any()), never());
        }
    }

    @Test
    void testMapperUtilsIntersect() {
        assertThat(MapperUtils.intersect(Arrays.asList("a", "b", "c", "d"), Arrays.asList("b", "d", "e", "f")))
            .containsExactly("b", "d");
    }

    @Test
    void testMapperUtilsIntersectEmptyLists() {
        assertThat(MapperUtils.intersect(Collections.emptyList(), Arrays.asList("a", "b"))).isEmpty();
    }

    @Test
    void testMapperUtilsIntersectNoCommonElements() {
        assertThat(MapperUtils.intersect(Arrays.asList("a", "b"), Arrays.asList("c", "d"))).isEmpty();
    }

    @Test
    void testMapperUtilsConstants() {
        assertThat(MapperUtils.GROUPS_INTERSECTION_THRESHOLD).isEqualTo(10);
        assertThat(MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE).isNotNull();
        assertThat(MapperUtils.TENANT_ID_ATTRIBUTE).isNotNull();
        assertThat(MapperUtils.GROUPS_USER_ATTRIBUTE).isNotNull();
        assertThat(MapperUtils.MULTIVALUED_ATTRIBUTE).isNotNull();
        assertThat(MapperUtils.RUNAI_ATTRIBUTE_IMPORTER_CATEGORY).isNotNull();
        assertThat(MapperUtils.DEFAULT_LIST_ATTRIBUTE).isNotNull();
    }

    @Test
    void testMapperUtilsCreateConfigProperty() {
        org.keycloak.provider.ProviderConfigProperty property = MapperUtils.createConfigProperty(
            "test.name", "Test Display", "Test Help",
            org.keycloak.provider.ProviderConfigProperty.STRING_TYPE, true, "default-value"
        );

        assertThat(property.getName()).isEqualTo("test.name");
        assertThat(property.getLabel()).isEqualTo("Test Display");
        assertThat(property.getHelpText()).isEqualTo("Test Help");
        assertThat(property.getType()).isEqualTo(org.keycloak.provider.ProviderConfigProperty.STRING_TYPE);
        assertThat(property.getDefaultValue()).isEqualTo("default-value");
    }
}
