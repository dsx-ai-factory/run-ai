/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// all Variables are here

import { Keycloak } from '../utils/Keycloak';
import { config, settings } from './config';
import { NodePoolsApi } from '../../api/apis/backendClusterApi/generated/api/nodePoolsApi';
import { backendClusterApi } from '../../api';
import { Utils } from '../utils/utils';
import { AuditLog } from '../../api/apis/auditApi';

export const PUBLIC_VLLM_IMAGE = 'vllm/vllm-openai:latest';

export const namespaceCP = 'runai-backend';
export const namespaceCluster = 'runai';
export const helmRepoName = 'runai-backend';
export const helmRepoNameStaging = 'runai-backend-staging';
export const helmChart = 'control-plane';
export const controlPlaneVersion = '2.12.1-staging.0';
export const clusterVersion = '2.9.6-staging.0';
export const tenantName: string = settings.isSelfHosted ? 'runai' : Utils.extractSaasTenantName();
export const realmName = 'automation';
export const samlRealmName = 'autosaml';
export const idpNameOidc = 'oidc';
export const idpNameSaml = 'saml';
export const idpNameOcp = 'openshift-v4';
export const keycloakClientName = 'automationclient';
export const oidcSsoUser = 'automationsso@example.com';
export const emailMapperConfig = {
    'attribute.nameformat': 'urn:oasis:names:tc:SAML:2.0:attrname-format:uri',
    'user.attribute': 'email',
    'friendly.name': 'email',
    'attribute.name': 'urn:oid:1.2.840.113549.1.9.1'
};
export const firstNameMapperConfig = {
    'attribute.nameformat': 'urn:oasis:names:tc:SAML:2.0:attrname-format:basic',
    'user.attribute': 'firstName',
    'friendly.name': 'firstName',
    'attribute.name': 'given_name'
};
export const lastNameMapperConfig = {
    'attribute.nameformat': 'urn:oasis:names:tc:SAML:2.0:attrname-format:basic',
    'user.attribute': 'lastName',
    'friendly.name': 'lastName',
    'attribute.name': 'family_name'
};
export const groupMapperConfig = {
    'attribute.name': 'GROUPS',
    'attribute.nameformat': 'urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified',
    single: false
};
export const ssoRedirectUri = `${config.tenantUrl}/auth/realms/runai/broker/saml/endpoint`;
// Lazy-loaded getter to avoid circular dependency
export const getSsoEntityId = (): string => Keycloak?.getSsoEntityId(config.tenantUrl);
export const samlUser = 'saml@example.com';
// Lazy-loaded getter to avoid circular dependency
export const getNodePoolsApi = (): NodePoolsApi => backendClusterApi?.nodePoolsApi;
export const invalidUuid = '550e8400-e29b-41d4-a716-446655440000';
export const kartaE2EFakeGroup = 'e2e.karta.test';
export const kartaE2EFakeKind = 'E2EFakeResource';
export const kartaE2EFakeVersion = 'v1';
export const hostnameLabel = 'kubernetes.io/hostname';
export const grantTypeMetadataKey = 'grant_type';
export const loginAuditAction = 'Login' as unknown as AuditLog.ActionEnum;
