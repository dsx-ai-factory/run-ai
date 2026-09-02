package com.runai.keycloak.broker.oidc.mappers;


import com.fasterxml.jackson.databind.node.JsonNodeFactory;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.junit.jupiter.api.Test;
import org.keycloak.broker.oidc.OIDCIdentityProvider;
import org.keycloak.broker.provider.BrokeredIdentityContext;
import org.keycloak.models.IdentityProviderMapperModel;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.models.UserModel;
import org.keycloak.provider.ProviderConfigProperty;
import org.mockito.ArgumentCaptor;

import java.util.*;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class CustomClaimToUserAttributeTest {
    @Test
    void testToList_Array() throws Exception {

        CustomClaimToUserAttribute t = new CustomClaimToUserAttribute();
        List<String> list = Arrays.asList("value1", "value2", "value3");
        List<String> result = t.toList(list);
        assertEquals(Arrays.asList("value1", "value2", "value3"), result);
        assertEquals(3, result.size());
    }

    @Test
    void testToList_Map() throws Exception {

        Map<String, Object> input = new HashMap<>();
        input.put("key1", "value1");
        input.put("key2", "value2");

        CustomClaimToUserAttribute t = new CustomClaimToUserAttribute();
        List<String> result = t.toList(input);
        assertEquals(Arrays.asList("value1", "value2"), result);
        assertEquals(2, result.size());
    }

    @Test
    void testToList_WithObjectNode() {
        ObjectNode objectNode = JsonNodeFactory.instance.objectNode();
        objectNode.put("key1", "value1");
        objectNode.put("key2", "value2");

        CustomClaimToUserAttribute t = new CustomClaimToUserAttribute();
        List<String> result = t.toList(objectNode);
        assertEquals(Arrays.asList("value1", "value2"), result);
        assertEquals(2, result.size());
    }

    @Test
    void testToList_WithInteger() {
        CustomClaimToUserAttribute t = new CustomClaimToUserAttribute();
        List<String> result = t.toList(42);
        assertEquals(Collections.singletonList("42"), result);
    }

    @Test
    void testToList_WithNull() {

        CustomClaimToUserAttribute t = new CustomClaimToUserAttribute();
        List<String> result = t.toList(null);
        assertEquals(Collections.emptyList(), result);
    }

    @Test
    void testToList_Empty() {

        CustomClaimToUserAttribute t = new CustomClaimToUserAttribute();
        List<String> result = t.toList("");
        assertEquals(Collections.singletonList(""), result);
    }

    @Test
    void testToList_EmptyMap() {
        ObjectNode objectNode = JsonNodeFactory.instance.objectNode();
        CustomClaimToUserAttribute t = new CustomClaimToUserAttribute();
        List<String> result = t.toList(objectNode);
        assertEquals(Collections.emptyList(), result);
    }

    @Test
    void testConfigProperties() {
        CustomClaimToUserAttribute mapper = new CustomClaimToUserAttribute();
        List<ProviderConfigProperty> properties = mapper.getConfigProperties();
        
        assertNotNull(properties);
        assertEquals(2, properties.size());
        
        // Check that our expected config properties are present
        boolean hasClaimProperty = properties.stream().anyMatch(p -> 
            CustomClaimToUserAttribute.CLAIM.equals(p.getName()) &&
            "Claim".equals(p.getLabel()) &&
            ProviderConfigProperty.STRING_TYPE.equals(p.getType()));
        
        boolean hasUserAttributeProperty = properties.stream().anyMatch(p -> 
            CustomClaimToUserAttribute.USER_ATTRIBUTE.equals(p.getName()) &&
            "User Attribute Name".equals(p.getLabel()) &&
            ProviderConfigProperty.USER_PROFILE_ATTRIBUTE_LIST_TYPE.equals(p.getType()));
        
        assertTrue(hasClaimProperty, "Should have CLAIM config property");
        assertTrue(hasUserAttributeProperty, "Should have USER_ATTRIBUTE config property");
    }

    @Test
    void testUpdateBrokeredUser_updatesWhenDuplicateCountDiffers() {
        IdentityProviderMapperModel mapperModel = mapperModelForRoles();
        UserModel user = mock(UserModel.class);
        when(user.getAttributeStream("roles")).thenAnswer(invocation -> List.of("team-a").stream());

        CustomClaimToUserAttribute mapper = new CustomClaimToUserAttribute();
        mapper.updateBrokeredUser(
                mock(KeycloakSession.class),
                mock(RealmModel.class),
                user,
                mapperModel,
                contextWithGroupsClaim(List.of("team-a", "team-a"))
        );

        @SuppressWarnings("unchecked")
        ArgumentCaptor<List<String>> valuesCaptor = ArgumentCaptor.forClass(List.class);
        verify(user).setAttribute(eq("roles"), valuesCaptor.capture());
        assertEquals(List.of("team-a", "team-a"), valuesCaptor.getValue());
        verify(user, never()).removeAttribute(anyString());
    }

    @Test
    void testUpdateBrokeredUser_skipsWhenSameElementsDifferentOrder() {
        IdentityProviderMapperModel mapperModel = mapperModelForRoles();
        UserModel user = mock(UserModel.class);
        when(user.getAttributeStream("roles")).thenAnswer(invocation -> List.of("team-a", "team-b").stream());

        CustomClaimToUserAttribute mapper = new CustomClaimToUserAttribute();
        mapper.updateBrokeredUser(
                mock(KeycloakSession.class),
                mock(RealmModel.class),
                user,
                mapperModel,
                contextWithGroupsClaim(List.of("team-b", "team-a"))
        );

        verify(user, never()).setAttribute(anyString(), anyList());
        verify(user, never()).removeAttribute(anyString());
    }

    @Test
    void testUpdateBrokeredUser_updatesWhenElementsDiffer() {
        IdentityProviderMapperModel mapperModel = mapperModelForRoles();
        UserModel user = mock(UserModel.class);
        when(user.getAttributeStream("roles")).thenAnswer(invocation -> List.of("team-a").stream());

        CustomClaimToUserAttribute mapper = new CustomClaimToUserAttribute();
        mapper.updateBrokeredUser(
                mock(KeycloakSession.class),
                mock(RealmModel.class),
                user,
                mapperModel,
                contextWithGroupsClaim(List.of("team-b"))
        );

        @SuppressWarnings("unchecked")
        ArgumentCaptor<List<String>> valuesCaptor = ArgumentCaptor.forClass(List.class);
        verify(user).setAttribute(eq("roles"), valuesCaptor.capture());
        assertEquals(List.of("team-b"), valuesCaptor.getValue());
        verify(user, never()).removeAttribute(anyString());
    }

    private static IdentityProviderMapperModel mapperModelForRoles() {
        IdentityProviderMapperModel mapperModel = mock(IdentityProviderMapperModel.class);
        when(mapperModel.getConfig()).thenReturn(Map.of(
                CustomClaimToUserAttribute.USER_ATTRIBUTE, "roles",
                CustomClaimToUserAttribute.CLAIM, "groups"
        ));
        return mapperModel;
    }

    private static BrokeredIdentityContext contextWithGroupsClaim(List<String> groups) {
        ObjectNode userInfo = JsonNodeFactory.instance.objectNode();
        var groupsNode = userInfo.putArray("groups");
        groups.forEach(groupsNode::add);

        BrokeredIdentityContext context = mock(BrokeredIdentityContext.class);
        when(context.getContextData()).thenReturn(Map.of(
                OIDCIdentityProvider.USER_INFO, userInfo
        ));
        return context;
    }
}