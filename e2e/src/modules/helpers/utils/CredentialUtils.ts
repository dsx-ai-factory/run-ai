/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { credentialUtils, flowRootState, logger, objectGenerator, waitUtils } from '..';
import { assetsApi, k8sObjectTrackerApi } from '../../api';
import {
    AccessKey,
    AccessKeyCreationRequest,
    AssetCreationRequest,
    AssetKind,
    AssetSyncStatusEnum,
    CredentialsListResponseEntry,
    DockerRegistry,
    DockerRegistryCreationRequest,
    GenericSecret,
    GenericSecretCreationRequest,
    KeyValuePair,
    NgcApiKey,
    NgcApiKeyCreationRequest,
    Password,
    PasswordCreationRequest,
    PasswordListResponse,
    Scope
} from '../../api/apis/assetsApi';
import { SecretListInfo } from '../../api/apis/k8sObjectsTrackerApi/generated/model/secretListInfo';
import { SecretType } from '../../api/apis/k8sObjectsTrackerApi/generated/model/secretType';
import { kubeCli } from '../../cli';
import { KubernetesObject, V1ObjectMeta, V1Secret } from '@kubernetes/client-node';
import { AssetMetaOptions } from './DataSourceUtils';
import { IncomingMessage } from 'http';
import { StatusCodes } from 'http-status-codes';

const RUNAI_ASSET_LABEL = 'run.ai/asset';

// Payload builders below use dummy credential values for asset API tests.
export class CredentialUtils {
    async list(): Promise<Array<CredentialsListResponseEntry>> {
        assetsApi.credentialsApi.accessToken = flowRootState.users.editor.accessToken;
        return (await assetsApi.credentialsApi.listCredentialsAssets(flowRootState.cluster.uuid)).body
            .entries;
    }

    async deleteAllCredentials(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            const credentials: CredentialsListResponseEntry[] = (
                await assetsApi.credentialsApi.listCredentialsAssets()
            ).body.entries;

            const credentialKindMap = {
                [AssetKind.AccessKey]: this.deleteAccessKeyAsset,
                [AssetKind.Password]: this.deleteAccessKeyAsset,
                [AssetKind.DockerRegistry]: this.deleteDockerRegistryAsset,
                [AssetKind.GenericSecret]: this.deleteGenericSecretAsset,
                [AssetKind.NgcApiKey]: this.deleteNgcApiKeyAsset
            };
            await Promise.all(
                credentials.map((cr) => {
                    if (credentialKindMap[cr.meta.kind]) {
                        return credentialKindMap[cr.meta.kind](cr.meta.id, accessToken);
                    } else {
                        throw new Error(`Failed to delete unknown credential kind: ${cr.meta.kind}`);
                    }
                })
            );
            if (credentials.length > 0) {
                logger.info('deleted all credentials successfully');
            }
        } catch (error) {
            logger.warn(`Failed to delete all credentials: ${error}`);
        }
    }

    async waitForCredentialsToBeDeleted(name: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const credentialsNames = (await this.list()).map((credential) => credential.meta.name);
                return !credentialsNames.includes(name);
            },
            'waiting for credentials to be deleted',
            true,
            10
        );
    }

    getAssetMeta(options: AssetMetaOptions, description?: string): AssetCreationRequest {
        let departmentId: string;
        let clusterId: string;
        let projectId: number;

        switch (options.scope) {
            case Scope.Department:
                departmentId = options.scopeId;
                break;
            case Scope.Project:
                projectId = +options.scopeId;
                break;
            case Scope.Cluster:
                clusterId = flowRootState.cluster.uuid;
                break;
        }

        return {
            name: options.name,
            description: description,
            scope: options.scope,
            clusterId,
            projectId,
            departmentId
        };
    }

    getDockerRegistryAssetPayload(
        credentialName: string,
        description?: string
    ): DockerRegistryCreationRequest {
        return {
            meta: credentialUtils.getAssetMeta(
                {
                    name: credentialName,
                    scope: Scope.Cluster,
                    scopeId: flowRootState.cluster.uuid
                },
                description
            ),
            spec: {
                user: 'username',
                password: 'password',
                url: 'https://my.url.com'
            }
        };
    }

    getAccessKeyAssetPayload(credentialName: string, description?: string): AccessKeyCreationRequest {
        return {
            meta: credentialUtils.getAssetMeta(
                {
                    name: credentialName,
                    scope: Scope.Cluster,
                    scopeId: flowRootState.cluster.uuid
                },
                description
            ),
            spec: {
                accessKeyId: '123',
                secretAccessKey: '456'
            }
        };
    }

    getPasswordAssetPayload(credentialName: string, description?: string): PasswordCreationRequest {
        return {
            meta: credentialUtils.getAssetMeta(
                {
                    name: credentialName,
                    scope: Scope.Cluster,
                    scopeId: flowRootState.cluster.uuid
                },
                description
            ),
            spec: {
                user: 'user',
                password: 'password'
            }
        };
    }

    getGenericSecretAssetPayload(
        credentialName: string,
        description?: string,
        projectId?: string
    ): GenericSecretCreationRequest {
        return {
            meta: credentialUtils.getAssetMeta(
                {
                    name: credentialName,
                    scope: projectId ? Scope.Project : Scope.Cluster,
                    scopeId: projectId ? projectId : flowRootState.cluster.uuid
                },
                description
            ),
            spec: {
                keyValuePairs: [{ key: 'key1', value: 'value1' }]
            }
        };
    }

    async getSecretInfo(
        name: string,
        scopeType: string,
        scopeId: string,
        secretType: SecretType
    ): Promise<Array<SecretListInfo>> {
        k8sObjectTrackerApi.secretsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        return (
            await k8sObjectTrackerApi.secretsApi.listSecrets(scopeType, scopeId, name, secretType.toString())
        ).body.secrets;
    }

    async getPasswordInfo(): Promise<{
        response: IncomingMessage;
        body: PasswordListResponse;
    }> {
        assetsApi.credentialsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        return await assetsApi.credentialsApi.listPasswords(
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            true
        );
    }

    buildUnlabeledSecret(name: string, namespace: string): V1Secret {
        return {
            apiVersion: 'v1',
            kind: 'Secret',
            metadata: { name, namespace },
            type: 'Opaque',
            stringData: { key: 'value' }
        };
    }

    buildLabeledSecret(
        name: string,
        namespace: string,
        labelKey = RUNAI_ASSET_LABEL,
        labelValue = 'generic'
    ): V1Secret {
        return {
            apiVersion: 'v1',
            kind: 'Secret',
            metadata: {
                name,
                namespace,
                labels: { [labelKey]: labelValue }
            },
            type: 'Opaque',
            stringData: { key: 'value' }
        };
    }

    async createLabeledSecret(secret: V1Secret): Promise<V1Secret> {
        const namespace = secret.metadata?.namespace;
        logger.debug(`Creating Secret ${namespace}/${secret.metadata?.name}`);
        return await kubeCli.CoreV1Api.createNamespacedSecret({ namespace, body: secret });
    }

    async deleteNamespacedSecret(name: string, namespace: string): Promise<void> {
        await kubeCli.CoreV1Api.deleteNamespacedSecret({ name, namespace });
    }

    getSecret(
        resourceType: SecretType,
        hasClusterWideLabel: boolean,
        departmentId?: string,
        namespace = 'runai'
    ): V1Secret {
        let namePrefix = `${resourceType}`;
        const labels = { 'run.ai/resource': resourceType.toString() };
        if (hasClusterWideLabel) labels['run.ai/cluster-wide'] = 'true';
        if (departmentId) {
            labels['run.ai/department'] = departmentId;
            namePrefix += `-dep-${departmentId}`;
        }
        return {
            apiVersion: 'v1',
            kind: 'Secret',
            metadata: {
                name: objectGenerator.generateShortNameByPrefix(namePrefix),
                namespace: namespace,
                labels: labels
            },
            type: 'Opaque',
            data: { user: Buffer.from('password').toString('base64') }
        };
    }

    async createExistingSecret(
        secretType: SecretType,
        hasClusterWideLabel: boolean,
        scopeId?: string,
        namespace = 'runai'
    ): Promise<V1ObjectMeta> {
        const secret = this.getSecret(secretType, hasClusterWideLabel, scopeId, namespace);
        await kubeCli.CoreV1Api.createNamespacedSecret({ namespace, body: secret });
        logger.info(`created ${secretType} secret named: ${secret.metadata.name}`);
        return secret.metadata;
    }

    async createGenericSecret(
        secretName: string,
        namespace: string,
        scope = Scope.Project,
        departmentId?: string
    ): Promise<KubernetesObject> {
        const secret = {
            apiVersion: 'v1',
            kind: 'Secret',
            metadata: {
                name: secretName,
                namespace: namespace,
                labels: {
                    'run.ai/resource': 'generic'
                }
            },
            type: 'Opaque',
            data: {
                username: 'dG9t',
                password: 'dG9t'
            }
        };
        if (scope === Scope.Cluster) secret.metadata.labels['run.ai/cluster-wide'] = 'true';
        if (scope === Scope.Department) secret.metadata.labels['run.ai/department'] = departmentId;
        return await kubeCli.KubernetesObjectApi.create(secret);
    }

    async createGenericSecretByApi(
        secretName: string,
        projectId: number,
        keyValuePairs: Array<KeyValuePair>,
        isPrivate: boolean = false
    ): Promise<GenericSecret> {
        try {
            const genericSecret: GenericSecretCreationRequest = {
                meta: {
                    name: secretName,
                    description: null,
                    scope: isPrivate ? Scope.Tenant : Scope.Project,
                    projectId: isPrivate ? undefined : projectId,
                    clusterId: null,
                    departmentId: null,
                    _private: isPrivate
                },
                spec: { existingSecretName: null, keyValuePairs }
            };

            const genericSecretResponse = await credentialUtils.createGenericSecretAsset(genericSecret);
            await credentialUtils.waitForGenericSecretAssetToBeReady(genericSecretResponse);
            return genericSecretResponse;
        } catch (error) {
            logger.error(`failed to create generic secret ${secretName}. Error: ${JSON.stringify(error)}`);
            throw error;
        }
    }

    async waitForExistingSecret(
        name: string,
        scopeType: string,
        scopeId: string,
        secretType: SecretType
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async (): Promise<boolean> => {
                const credentialsNames = await this.getSecretInfo(name, scopeType, scopeId, secretType);
                return credentialsNames.some((secret) => secret.name === name);
            },
            `waiting for existing credential. name: ${name}, scope: ${scopeType}, scope id: ${scopeId}, type: ${secretType}`,
            true,
            60
        );
    }

    async waitForPasswordStatus(name: string, status: AssetSyncStatusEnum): Promise<void> {
        await waitUtils.waitForCondition(
            async (): Promise<boolean> => {
                const passwdNames = await this.getPasswordInfo();
                const password: Password | undefined = passwdNames.body.entries.find(
                    (passwd) => passwd.meta.name === name
                );
                return password?.status.status === status;
            },
            `waiting for password status, name: ${name}, status: ${status}.`,
            true,
            60
        );
    }

    // ASSET CREATION UTILS //
    async createAccessKeyAsset(
        accessKey: AccessKeyCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<AccessKey> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            const res = await assetsApi.credentialsApi.createAccessKey(accessKey);
            logger.info(`created access-key ${accessKey.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`Failed to create access-key: ${error}`);
            throw error;
        }
    }

    async createGenericSecretWithArbitraryFields(projectId: string): Promise<GenericSecret> {
        const name = objectGenerator.generateShortNameByPrefix('gs');
        const genericSecretCCreationRequest = credentialUtils.getGenericSecretAssetPayload(
            name,
            undefined,
            projectId
        );
        return await credentialUtils.createGenericSecretAsset(genericSecretCCreationRequest);
    }

    async createAccessKeyWithArbitraryFields(): Promise<AccessKey> {
        const accessKeyCreationRequest: AccessKeyCreationRequest = {
            meta: {
                name: objectGenerator.generateShortNameByPrefix('access-key'),
                scope: Scope.Cluster,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                accessKeyId: objectGenerator.generateNameByPrefix('ak-id'),
                secretAccessKey: objectGenerator.generateNameByPrefix('ak-secret')
            }
        };
        return await credentialUtils.createAccessKeyAsset(accessKeyCreationRequest);
    }

    async createUserPasswordWithArbitraryFields(): Promise<Password> {
        const passwordCreationRequest: PasswordCreationRequest = {
            meta: {
                name: objectGenerator.generateShortNameByPrefix('user-pass'),
                scope: Scope.Cluster,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                user: 'user',
                password: 'password'
            }
        };
        return await credentialUtils.createPasswordAsset(passwordCreationRequest);
    }

    async createPasswordAsset(
        password: PasswordCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Password> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            const res = await assetsApi.credentialsApi.createPassword(password);
            logger.info(`created password ${password.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`Failed to create password: ${error}`);
            throw error;
        }
    }

    async createDockerRegistryAsset(
        dockerRegistry: DockerRegistryCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<DockerRegistry> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            const res = await assetsApi.credentialsApi.createDockerRegistry(dockerRegistry);
            logger.info(`created docker-registry ${dockerRegistry.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`Failed to create docker-registry: ${error}`);
            throw error;
        }
    }

    async createGenericSecretAsset(
        genericSecret: GenericSecretCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<GenericSecret> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            const res = await assetsApi.credentialsApi.createGenericSecret(genericSecret);
            logger.info(`created generic-secret ${genericSecret.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`Failed to create generic-secret: ${error}`);
            throw error;
        }
    }

    async createNgcApiKeyAsset(
        ngcApiKey: NgcApiKeyCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<NgcApiKey> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            const res = await assetsApi.credentialsApi.createNgcApiKey(ngcApiKey);
            logger.info(`created ngc-api-key asset ${ngcApiKey.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`Failed to create ngc-api-key asset: ${error}`);
            throw error;
        }
    }

    // ASSET DELETION UTILS //
    async deleteAccessKeyAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        assetsApi.credentialsApi.accessToken = accessToken;
        await waitUtils.waitForCondition(
            async () => {
                try {
                    await assetsApi.credentialsApi.deleteAccessKey(id);
                    logger.info(`Deleted access key with id: ${id} successfully`);
                    return true; // Successfully deleted
                } catch (error: any) {
                    if (error?.body?.code === 412) {
                        // Retry only if the error is 412 (asset in use)
                        logger.warn(`Retrying delete for access-key with id: ${id} due to 412 error...`);
                        return false; // Continue waiting
                    } else if (error?.body?.code === StatusCodes.NOT_FOUND) {
                        return true;
                    } else {
                        logger.error(`Failed to delete for access-key with id: ${id}, error: [${error}]`);
                        throw error;
                    }
                }
            },
            `Delete access key with ID ${id}`,
            true,
            40,
            2
        );
    }

    async deletePasswordAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        assetsApi.credentialsApi.accessToken = accessToken;
        await waitUtils.waitForCondition(
            async () => {
                try {
                    await assetsApi.credentialsApi.deletePassword(id);
                    logger.info(`Deleted user-password with id: ${id} successfully`);
                    return true; // Successfully deleted
                } catch (error: any) {
                    if (error?.body?.code === 412) {
                        // Retry only if the error is 412 (asset in use)
                        logger.warn(`Retrying delete for user-password with id: ${id} due to 412 error...`);
                        return false; // Continue waiting
                    } else {
                        logger.error(`Failed to delete for user-password with id: ${id}, error: [${error}]`);
                        throw error;
                    }
                }
            },
            `Delete user-password with ID ${id}`,
            true,
            40,
            2
        );
    }

    async deleteDockerRegistryAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            await assetsApi.credentialsApi.deleteDockerRegistry(id);
            logger.info(`deleted docker-registry ${id} successfully`);
        } catch (error) {
            logger.error(`Failed to delete docker-registry: ${error}`);
            throw error;
        }
    }

    async deleteGenericSecretAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            await assetsApi.credentialsApi.deleteGenericSecret(id);
            logger.info(`deleted generic-secret ${id} successfully`);
        } catch (error) {
            logger.error(`Failed to delete generic-secret: ${error}`);
            throw error;
        }
    }

    async deleteNgcApiKeyAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.credentialsApi.accessToken = accessToken;
            await assetsApi.credentialsApi.deleteNgcApiKey(id);
            logger.info(`deleted ngc-api-key ${id} successfully`);
        } catch (error) {
            logger.error(`Failed to delete ngc-api-key: ${error}`);
            throw error;
        }
    }

    async validateExpectedStatus(credentialName: string, expectedStatus: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                assetsApi.credentialsApi.accessToken = flowRootState.users.system_administrator.accessToken;
                const credential = (
                    await assetsApi.credentialsApi.listCredentialsAssets(
                        credentialName,
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        true
                    )
                ).body.entries[0];
                return credential?.status?.status.toString() === expectedStatus;
            },
            `waiting for ${credentialName} status to be ${expectedStatus}`,
            true
        );
    }

    async waitForGenericSecretAssetToBeReady(genericSecret: GenericSecret): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const response = await assetsApi.credentialsApi.getGenericSecretById(
                    genericSecret.meta.id,
                    undefined,
                    true
                );
                const genericSecretAsset = response.body;

                logger.debug(
                    `generic secret asset ${genericSecret.meta.name} status: ${genericSecretAsset.status?.status}`
                );
                return genericSecretAsset.status?.status.toString() === 'No issues found';
            },
            `waiting for generic secret asset ${genericSecret.meta.name} to be ready`,
            true
        );
    }

    async waitForNgcApiKeyAssetToBeReady(ngcApiKey: NgcApiKey): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const response = await assetsApi.credentialsApi.getNgcApiKeyById(
                    ngcApiKey.meta.id,
                    undefined,
                    true
                );
                const ngcApiKeyAsset = response.body;
                logger.debug(
                    `ngc api key asset ${ngcApiKey.meta.name} status: ${ngcApiKeyAsset.status?.status}`
                );
                return ngcApiKeyAsset.status?.status.toString() === 'No issues found';
            },
            `waiting for ngc api key asset ${ngcApiKey.meta.name} to be ready`,
            true
        );
    }

    async waitForDockerRegistryAssetToBeReady(dockerRegistry: DockerRegistry): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const response = await assetsApi.credentialsApi.getDockerRegistryById(
                    dockerRegistry.meta.id,
                    undefined,
                    true
                );
                const dockerRegistryAsset = response.body;
                logger.debug(
                    `docker registry asset ${dockerRegistry.meta.name} status: ${dockerRegistryAsset.status?.status}`
                );
                return dockerRegistryAsset.status?.status.toString() === 'No issues found';
            },
            `waiting for docker registry asset ${dockerRegistry.meta.name} to be ready`,
            true
        );
    }

    async waitForPersonalCredentialSecrets(
        credentialName: string,
        accessToken: string,
        timeoutSeconds = 120
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                assetsApi.credentialsApi.accessToken = accessToken;
                const response = await assetsApi.credentialsApi.listCredentialsAssets(
                    credentialName,
                    undefined,
                    undefined,
                    undefined,
                    undefined,
                    undefined,
                    true,
                    undefined,
                    true
                );
                return response.body.entries[0]?.status?.status?.toString() === 'No issues found';
            },
            `waiting for personal credential "${credentialName}" to be synced on cluster`,
            true,
            timeoutSeconds
        );
    }
}
