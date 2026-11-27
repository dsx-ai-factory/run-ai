/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { identityManagerApi } from '../../api';
import {
    IdpResources,
    MandatoryClaim,
    Mappers,
    SettingsGetResponse,
    SettingsKeyEnum
} from '../../api/apis/identityManagerApi/generated/api';
import { Idp } from '../../api/apis/identityManagerApi/generated/model/idp';
import { IdpCreationRequest } from '../../api/apis/identityManagerApi/generated/model/idpCreationRequest';
import { IdpCreationRequestTypeEnum } from '../../api/apis/identityManagerApi/generated/model/idpCreationRequestTypeEnum';
import { IdpPostResponse } from '../../api/apis/identityManagerApi/generated/model/idpPostResponse';
import { SamlCreationDataMetadataXmlTypeEnum } from '../../api/apis/identityManagerApi/generated/model/samlCreationDataMetadataXmlTypeEnum';
import { UpdateSecuritySettingByKeyRequest } from '../../api/apis/identityManagerApi/generated/model/updateSecuritySettingByKeyRequest';
import { IncomingMessage } from 'http';
import { logger } from '../logger/logger';

export class IdpUtils {
    async getIdp(
        accessToken: string,
        idpName: string
    ): Promise<{
        response: IncomingMessage;
        body: Idp;
    }> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        const result = await identityManagerApi.idpsApi.getIdp(idpName);
        if (!result.body) {
            throw new Error(`Failed to get IDP: ${idpName}`);
        }
        return result as { response: IncomingMessage; body: Idp };
    }

    async getIdpMultiTenant(
        accessToken: string,
        idpName: string,
        tenantId: string
    ): Promise<{
        response: IncomingMessage;
        body: Idp;
    }> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        const result = await identityManagerApi.idpsApi.getIdp(idpName, tenantId);
        if (!result.body) {
            throw new Error(`Failed to get IDP: ${idpName}`);
        }
        return result as { response: IncomingMessage; body: Idp };
    }

    async createIdpOidc(
        accessToken: string,
        idpName: string,
        clientId: string,
        clientSecret: string,
        descoveryURL: string
    ): Promise<IdpPostResponse> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        const idpCreationRequest: IdpCreationRequest = {
            type: IdpCreationRequestTypeEnum.Oidc,
            name: idpName,
            oidcData: {
                clientId: clientId,
                clientSecret: clientSecret,
                discoverDocumentUrl: descoveryURL
            }
        };
        const idp = await identityManagerApi.idpsApi.createIdp(idpCreationRequest);
        return idp.body;
    }

    async deleteIdp(
        accessToken: string,
        idpName: string,
        tenantId?: string
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        const deletedidp = await identityManagerApi.idpsApi.deleteIdp(idpName, tenantId);
        return deletedidp;
    }

    async updateIdpOidc(
        accessToken: string,
        idpName: string,
        clientId: string,
        clientSecret: string,
        discoverDocumentUrl: string,
        defaultScope?: string[],
        mandatoryClaim?: MandatoryClaim,
        tenantId?: string,
        mappers?: Mappers
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        const idp: IdpCreationRequest = {
            type: IdpCreationRequestTypeEnum.Oidc,
            name: idpName,
            mappers: mappers || {
                uid: 'UID',
                gid: 'GID',
                groups: 'GROUPS',
                supplementaryGroups: 'SUPPLEMENTARYGROUPS',
                email: 'email'
            },

            oidcData: {
                clientId: clientId,
                clientSecret: clientSecret,
                discoverDocumentUrl: discoverDocumentUrl,
                scopes: defaultScope,
                mandatoryClaim: mandatoryClaim
            }
        };

        const updateidp = await identityManagerApi.idpsApi.updateIdp(idpName, idp, tenantId);
        return updateidp;
    }

    async getIdpMappers(accessToken: string, idpName: string): Promise<Mappers> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        return (await identityManagerApi.idpsApi.getIdpMappers(idpName)).body;
    }

    async getIdps(accessToken: string): Promise<Idp[]> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        return (await identityManagerApi.idpsApi.getIdps()).body;
    }

    async updateIdpMappers(
        accessToken: string,
        idpName: string,
        email: string,
        gid: string,
        groups: string,
        supplyentarygroups: string,
        uid: string
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        const mappers = {
            email: email,
            gid: gid,
            groups: groups,
            supplementaryGroups: supplyentarygroups,
            uid: uid
        };
        return await identityManagerApi.idpsApi.updateIdpMappers(idpName, mappers);
    }

    async getIdpResources(accessToken: string, idpName: string): Promise<IdpResources> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        return (await identityManagerApi.idpsApi.getIdpResources(idpName)).body;
    }

    async createIdpSaml(
        accessToken: string,
        idpName: string,
        metadataXmlUrl: string,
        entityId: string
    ): Promise<IdpPostResponse> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        const idpCreationRequest: IdpCreationRequest = {
            type: IdpCreationRequestTypeEnum.Saml,
            name: idpName,
            samlData: {
                metadataXmlUrl: metadataXmlUrl,
                metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.Url,
                entityId: entityId
            }
        };
        const idp = await identityManagerApi.idpsApi.createIdp(idpCreationRequest);
        return idp.body;
    }

    async autoRedirectSSO(
        accessToken: string,
        autoRedirectSSO: boolean,
        idpAlias: string
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.settingsApi.accessToken = accessToken;
        const updateIdmSettingByKeyRequest: UpdateSecuritySettingByKeyRequest = {
            enabled: autoRedirectSSO,
            idpAlias: idpAlias
        };
        const response = await identityManagerApi.settingsApi.updateSecuritySettingByKey(
            SettingsKeyEnum.AutoRedirectSso,
            updateIdmSettingByKeyRequest
        );
        return response;
    }

    async getSecuritySettingByKey(
        accessToken: string,
        key: SettingsKeyEnum
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.settingsApi.accessToken = accessToken;
        const response = await identityManagerApi.settingsApi.getSecuritySettingByKey(key);
        return response;
    }

    async getSecuritySettings(accessToken: string): Promise<{
        response: IncomingMessage;
        body?: SettingsGetResponse;
    }> {
        identityManagerApi.settingsApi.accessToken = accessToken;
        const response = await identityManagerApi.settingsApi.getSecuritySettings();
        logger.info(`Security settings: ${JSON.stringify(response.body)}`);
        return response;
    }

    async updateSecuritySettingByKey(
        accessToken: string,
        key: SettingsKeyEnum,
        value: UpdateSecuritySettingByKeyRequest
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.settingsApi.accessToken = accessToken;
        return await identityManagerApi.settingsApi.updateSecuritySettingByKey(key, value);
    }
}
