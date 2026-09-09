/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package com.runai.keycloak.broker.oidc.mappers;

import org.keycloak.models.ProtocolMapperModel;
import org.keycloak.models.UserSessionModel;
import org.keycloak.provider.ProviderConfigProperty;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import com.fasterxml.jackson.databind.ObjectMapper;

public class MapperUtils {
    public static final String AUTHORIZATION_SERVICE_URL_ENV_VAR = "KC_RUNAI_AUTHORIZATION_SERVICE_URL";
    public static final String DEFAULT_AUTHORIZATION_SERVICE_URL = "http://runai-backend-authorization:8080";
    public static final String AUTHORIZATION_SERVICE_GROUPS_PATH = "/internal/authorization/groups";
    public static final String NAMES_KEY = "names";
    public static final String TENANT_HEADER = "X-Runai-Tenant-Id";
    public static final String DEFAULT_LIST_ATTRIBUTE = "GROUPS";
    public static final int TIMEOUT_MILLIS = 2000;
    public static final int GROUPS_INTERSECTION_THRESHOLD = 10;
    public static final String GROUPS_INTERSECTION_ENABLED_ATTRIBUTE = "groups.intersection.enabled";
    public static final String TENANT_ID_ATTRIBUTE = "tenant.id";
    public static final String GROUPS_USER_ATTRIBUTE = "groups.user.attribute";
    public static final String MULTIVALUED_ATTRIBUTE = "multivalued";
    public static final String RUNAI_ATTRIBUTE_IMPORTER_CATEGORY = "Runai Attribute Importer";
    private static final class HttpClientHolder {
        static final HttpClient INSTANCE = HttpClient.newBuilder()
            .connectTimeout(Duration.ofMillis(TIMEOUT_MILLIS))
            .build();
    }

    public static List<String> getUserAttribute(ProtocolMapperModel mappingModel, UserSessionModel userSession, String attributeKey) {
        var listAttr = mappingModel.getConfig().get(attributeKey);
        listAttr = Objects.requireNonNullElse(listAttr, DEFAULT_LIST_ATTRIBUTE);
        if (listAttr.isBlank()) {
            listAttr = DEFAULT_LIST_ATTRIBUTE;
        }
        var userAttributeList = userSession.getUser().getAttributes().get(listAttr);
        if (userAttributeList == null) {
            return List.of();
        }
        return List.copyOf(userAttributeList);
    }

    public static List<String> getUserAttribute(org.keycloak.models.UserModel user, org.keycloak.models.ProtocolMapperModel mappingModel, String attributeKey) {
        String listAttr = mappingModel.getConfig().get(attributeKey);
        listAttr = Objects.requireNonNullElse(listAttr, DEFAULT_LIST_ATTRIBUTE);
        if (listAttr.isBlank()) {
            listAttr = DEFAULT_LIST_ATTRIBUTE;
        }
        List<String> userAttributeList = user.getAttributes().get(listAttr);
        if (userAttributeList == null) {
            return List.of();
        }
        return List.copyOf(userAttributeList);
    }

    public static List<String> fetchGroupsFromAuthorizationService(String tenantId, org.slf4j.Logger logger) {
        var groups = new ArrayList<String>();
        var authorizationGroupsApiUrl = getAuthorizationServiceUrl() + AUTHORIZATION_SERVICE_GROUPS_PATH;
        logger.debug("[MapperUtils] Fetching groups from URL: {}", authorizationGroupsApiUrl);
        try {
            var requestBuilder = HttpRequest.newBuilder()
                .uri(new URL(authorizationGroupsApiUrl).toURI())
                .timeout(Duration.ofMillis(TIMEOUT_MILLIS))
                .GET();
            if (tenantId != null && !tenantId.isBlank()) {
                logger.debug("[MapperUtils] Setting X-Runai-Tenant-Id header: {}", tenantId);
                requestBuilder.header(TENANT_HEADER, tenantId);
            }
            var request = requestBuilder.build();
            var response = HttpClientHolder.INSTANCE.send(request, HttpResponse.BodyHandlers.ofString());
            int status = response.statusCode();
            if (status == HttpURLConnection.HTTP_OK) {
                var json = response.body();
                var result = new ObjectMapper().readValue(json, Map.class);
                var namesObj = result.get(NAMES_KEY);
                if (namesObj instanceof List<?> list) {
                    groups = new ArrayList<>((List<String>) list);
                }
                logger.debug("[MapperUtils] Parsed groups from 'names': {}", groups);
            } else {
                logger.error("[MapperUtils] Failed to fetch groups from authorization service. Status code: {}", status);
                throw new jakarta.ws.rs.WebApplicationException("Internal server error", jakarta.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR);
            }
        } catch (Exception e) {
            logger.error("[MapperUtils] About to throw WebApplicationException due to exception: {}", e.getMessage(), e);
            throw new jakarta.ws.rs.WebApplicationException("Internal server error", e, jakarta.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR);
        }
        return List.copyOf(groups);
    }

    public static List<String> intersect(List<String> a, List<String> b) {
        if (a == null || b == null) return new ArrayList<>();
        return a.stream().filter(b::contains).toList();
    }

    public static String getAuthorizationServiceUrl() {
        return getenvOrDefault(AUTHORIZATION_SERVICE_URL_ENV_VAR, DEFAULT_AUTHORIZATION_SERVICE_URL);
    }

    public static String getenvOrDefault(String key, String defaultValue) {
        String value = System.getenv(key);
        return (value != null && !value.isEmpty()) ? value : defaultValue;
    }

    public static ProviderConfigProperty createConfigProperty(
        String name, String label, String helpText, String type, boolean required, String defaultValue) {
        ProviderConfigProperty property = new ProviderConfigProperty();
        property.setName(name);
        property.setLabel(label);
        property.setHelpText(helpText);
        property.setType(type);
        property.setRequired(required);
        if (defaultValue != null) {
            property.setDefaultValue(defaultValue);
        }
        return property;
    }

    /**
     * Convenience method to create a ProviderConfigProperty with default required=false and no default value
     */
    public static ProviderConfigProperty createConfigProperty(
        String name, String label, String helpText, String type) {
        return createConfigProperty(name, label, helpText, type, false, null);
    }


} 