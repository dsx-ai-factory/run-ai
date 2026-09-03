package com.runai.keycloak.broker.oidc.mappers;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.keycloak.models.UserSessionModel;
import org.keycloak.models.UserModel;
import org.keycloak.models.ProtocolMapperModel;
import org.keycloak.representations.IDToken;

import java.util.*;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;
import org.mockito.MockedStatic;
import static org.mockito.ArgumentMatchers.*;

import jakarta.ws.rs.WebApplicationException;
import jakarta.ws.rs.core.Response;

import com.runai.keycloak.broker.oidc.mappers.MapperUtils;
import com.runai.keycloak.broker.oidc.mappers.CounterMapper;

class CustomGroupsMapperTest {
    ProtocolMapperModel mappingModel;
    UserSessionModel userSession;
    UserModel user;
    IDToken token;
    Map<String, List<String>> attributes;

    @BeforeEach
    void setUp() {
        mappingModel = mock(ProtocolMapperModel.class);
        userSession = mock(UserSessionModel.class);
        user = mock(UserModel.class);
        token = new IDToken();
        attributes = new HashMap<>();
        when(userSession.getUser()).thenReturn(user);
        when(user.getAttributes()).thenReturn(attributes);
        when(user.getUsername()).thenReturn("testuser");
    }

    private void setupLocalGroups(List<String> groups) {
        attributes.put("GROUPS", groups);
    }

    private void setupConfig(Map<String, String> config) {
        when(mappingModel.getConfig()).thenReturn(config);
    }

    private void setUsername(String username) {
        when(user.getUsername()).thenReturn(username);
    }

    private void setupDefaultConfig() {
        setupConfig(Map.of(
            MapperUtils.GROUPS_USER_ATTRIBUTE, "GROUPS",
            "claim.name", "groups",
            MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE, "true"
        ));
    }

    private void stubAllStaticMethods(MockedStatic<MapperUtils> mocked) {
        mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenCallRealMethod();
        mocked.when(() -> MapperUtils.intersect(anyList(), anyList())).thenCallRealMethod();
        mocked.when(() -> MapperUtils.getenvOrDefault(anyString(), anyString())).thenCallRealMethod();
        mocked.when(() -> MapperUtils.getAuthorizationServiceUrl()).thenCallRealMethod();
    }

    @Test
    void testIntersectionLogic() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            stubAllStaticMethods(mocked);
            setupLocalGroups(new ArrayList<>() {{ for (int i = 0; i < 15; i++) add("group" + i); add("shared1"); add("shared2"); }});
            setupConfig(Map.of(
                MapperUtils.GROUPS_USER_ATTRIBUTE, "GROUPS",
                MapperUtils.TENANT_ID_ATTRIBUTE, "tenant1",
                "claim.name", "groups",
                MapperUtils.GROUPS_INTERSECTION_ENABLED_ATTRIBUTE, "true"
            ));
            setUsername("testuser");
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenReturn(Arrays.asList("shared1", "shared2", "external1"));
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenReturn(Arrays.asList("shared1", "shared2", "external1"));
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            Object claim = token.getOtherClaims().get("groups");
            assertTrue(claim instanceof List);
            @SuppressWarnings("unchecked")
            List<String> result = (List<String>) claim;
            assertEquals(Arrays.asList("shared1", "shared2"), result);
        }
    }

    @Test
    void testNoIntersectionWhenFewGroups() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            stubAllStaticMethods(mocked);
            setupLocalGroups(Arrays.asList("g1", "g2", "g3"));
            setupConfig(Map.of(
                MapperUtils.GROUPS_USER_ATTRIBUTE, "GROUPS",
                MapperUtils.TENANT_ID_ATTRIBUTE, "tenant1",
                "claim.name", "groups"
            ));
            setUsername("testuser");
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenCallRealMethod();
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            Object claim = token.getOtherClaims().get("groups");
            assertTrue(claim instanceof List);
            @SuppressWarnings("unchecked")
            List<String> result = (List<String>) claim;
            assertEquals(Arrays.asList("g1", "g2", "g3"), result);
        }
    }

    @Test
    void testNoLocalGroups() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            stubAllStaticMethods(mocked);
            setupLocalGroups(Collections.emptyList());
            setupDefaultConfig();
            setUsername("user1");
            mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenReturn(Collections.emptyList());
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenCallRealMethod();
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertFalse(token.getOtherClaims().containsKey("groups"));
        }
    }

    @Test
    void testNoExternalGroups() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            stubAllStaticMethods(mocked);
            setupLocalGroups(new ArrayList<>() {{ for (int i = 0; i < 11; i++) add("g" + i); }});
            setupDefaultConfig();
            setUsername("user2");
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenReturn(Collections.emptyList());
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenReturn(Collections.emptyList());
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertEquals(Collections.emptyList(), token.getOtherClaims().get("groups"));
        }
    }

    @Test
    void testNullLocalGroupsAttribute() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            attributes.clear();
            setupDefaultConfig();
            setUsername("user3");
            mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenReturn(Collections.emptyList());
            mocked.when(() -> MapperUtils.intersect(anyList(), anyList())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenCallRealMethod();
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertFalse(token.getOtherClaims().containsKey("groups"));
        }
    }

    @Test
    void testNullExternalGroups() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            setupLocalGroups(new ArrayList<>() {{ for (int i = 0; i < 11; i++) add("g" + i); }});
            setupDefaultConfig();
            setUsername("user4");
            mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenCallRealMethod();
            mocked.when(() -> MapperUtils.intersect(anyList(), anyList())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenReturn(Collections.emptyList());
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenReturn(Collections.emptyList());
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertEquals(Collections.emptyList(), token.getOtherClaims().get("groups"));
        }
    }

    @Test
    void testTenantIdNull() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            setupLocalGroups(List.of("g1", "g2"));
            setupDefaultConfig();
            setUsername("user5");
            mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenCallRealMethod();
            mocked.when(() -> MapperUtils.intersect(anyList(), anyList())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenReturn(List.of("g1", "g2"));
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenReturn(List.of("g1", "g2"));
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertEquals(List.of("g1", "g2"), token.getOtherClaims().get("groups"));
        }
    }

    @Test
    void testTenantIdEmpty() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            setupLocalGroups(List.of("g1", "g2"));
            setupDefaultConfig();
            setUsername("user6");
            mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenCallRealMethod();
            mocked.when(() -> MapperUtils.intersect(anyList(), anyList())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenReturn(List.of("g1", "g2"));
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenReturn(List.of("g1", "g2"));
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertEquals(List.of("g1", "g2"), token.getOtherClaims().get("groups"));
        }
    }

    @Test
    void testConfigAttributeMissing() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            setupLocalGroups(List.of("g1", "g2"));
            setupDefaultConfig();
            setUsername("user7");
            mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenCallRealMethod();
            mocked.when(() -> MapperUtils.intersect(anyList(), anyList())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenReturn(List.of("g1", "g2"));
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenReturn(List.of("g1", "g2"));
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertEquals(List.of("g1", "g2"), token.getOtherClaims().get("groups"));
        }
    }

    @Test
    void testConfigAttributeEmptyString() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            setupLocalGroups(List.of("g1", "g2"));
            setupConfig(Map.of(MapperUtils.GROUPS_USER_ATTRIBUTE, "", "claim.name", "groups"));
            setUsername("user8");
            mocked.when(() -> MapperUtils.getUserAttribute(any(ProtocolMapperModel.class), any(UserSessionModel.class), eq(MapperUtils.GROUPS_USER_ATTRIBUTE))).thenCallRealMethod();
            mocked.when(() -> MapperUtils.intersect(anyList(), anyList())).thenCallRealMethod();
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenReturn(List.of("g1", "g2"));
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenReturn(List.of("g1", "g2"));
            CustomGroupsMapper mapper = new CustomGroupsMapper();
            mapper.setClaim(token, mappingModel, userSession);
            assertEquals(List.of("g1", "g2"), token.getOtherClaims().get("groups"));
        }
    }

    @Test
    void testGroupsCountClaimWithNonEmptyGroups() {
        CounterMapper mapper = spy(new CounterMapper());
        setupLocalGroups(List.of("g1", "g2", "g3"));
        setupConfig(Map.of(
            "list.user.attribute", "GROUPS",
            "count.claim", "groups_count"
        ));
        mapper.setClaim(token, mappingModel, userSession);
        assertEquals(3, token.getOtherClaims().get("groups_count"));
    }

    @Test
    void testGroupsCountClaimWithEmptyGroups() {
        CounterMapper mapper = spy(new CounterMapper());
        setupLocalGroups(Collections.emptyList());
        setupConfig(Map.of(
            "list.user.attribute", "GROUPS",
            "count.claim", "groups_count"
        ));
        mapper.setClaim(token, mappingModel, userSession);
        assertNull(token.getOtherClaims().get("groups_count"));
    }

    @Test
    void testGetAuthorizationServiceUrl_returnsDefaultWhenEnvVarNotSet() {
        String result = MapperUtils.getAuthorizationServiceUrl();
        assertEquals(MapperUtils.DEFAULT_AUTHORIZATION_SERVICE_URL, result);
    }

    @Test
    void testFetchGroupsFromAuthorizationServiceThrowsWebApplicationException() {
        try (MockedStatic<MapperUtils> mocked = mockStatic(MapperUtils.class)) {
            stubAllStaticMethods(mocked);
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), any())).thenThrow(new WebApplicationException("Simulated error", Response.Status.INTERNAL_SERVER_ERROR));
            mocked.when(() -> MapperUtils.fetchGroupsFromAuthorizationService(anyString(), isNull())).thenThrow(new WebApplicationException("Simulated error", Response.Status.INTERNAL_SERVER_ERROR));
            WebApplicationException exception = assertThrows(WebApplicationException.class, () -> {
                MapperUtils.fetchGroupsFromAuthorizationService("tenant-test", null);
            });
            assertEquals(Response.Status.INTERNAL_SERVER_ERROR.getStatusCode(), exception.getResponse().getStatus());
        }
    }
} 