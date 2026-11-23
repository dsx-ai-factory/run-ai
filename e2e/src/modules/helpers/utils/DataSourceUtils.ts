/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { dataSourcesUtils, objectGenerator, waitUtils } from '..';
import { assetsApi, k8sObjectTrackerApi } from '../../api';
import { expect } from 'chai';
import {
    AssetCreationRequest,
    AssetKind,
    ConfigMapAsset,
    ConfigMapCreationRequest,
    type DatasourceListResponseEntry,
    GitAsset,
    GitCreationRequest,
    HostPathAsset,
    HostPathCreationRequest,
    NFSAsset,
    NFSCreationRequest,
    PVCAsset,
    PVCCreationRequest,
    S3Asset,
    S3CreationRequest,
    Scope,
    SecretAsset,
    SecretAssetCreationRequest
} from '../../api/apis/assetsApi';
import { flowRootState } from '../config/config';
import { logger } from '../logger/logger';
import { kubeCli } from '../../cli';
import { PVCListInfo } from '../../api/apis/k8sObjectsTrackerApi/generated/model/pVCListInfo';
import {
    V1ConfigMap,
    V1ObjectMeta,
    V1PersistentVolumeClaim,
    V1PersistentVolumeClaimSpec,
    V1ResourceRequirements
} from '@kubernetes/client-node';
import { ConfigMapListInfo } from '../../api/apis/k8sObjectsTrackerApi/generated/api';
import { DataSourcePage } from '../../ui/runaiV2/pages/DataSourcePage';
import { StorageClass } from '../../api/apis/backendClusterApi';
import { Project } from '../../api/apis/projectsDepartmentsApi';

export interface AssetMetaOptions {
    name: string;
    scope: Scope;
    scopeId: string;
}

export class DataSourceUtils {
    set accessToken(accessToken: string) {
        assetsApi.datasourcesApi.accessToken = accessToken;
    }

    async list(): Promise<Array<DatasourceListResponseEntry>> {
        assetsApi.datasourcesApi.accessToken = flowRootState.users.editor.accessToken;
        return (await assetsApi.datasourcesApi.listDatasourceAssets(flowRootState.cluster.uuid)).body.entries;
    }

    async getId(dataSourceName: string): Promise<string> {
        const dataSources = await this.list();
        return dataSources.find((ds) => ds.meta.name === dataSourceName).meta.id;
    }

    async deleteAllDataSources(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.datasourcesApi.accessToken = accessToken;
            const dataSources: DatasourceListResponseEntry[] = (
                await assetsApi.datasourcesApi.listDatasourceAssets()
            ).body.entries;

            const dataSourceKindMap = {
                [AssetKind.Git]: this.deleteGitAsset,
                [AssetKind.S3]: this.deleteS3Asset,
                [AssetKind.Nfs]: this.deleteNfsAsset,
                [AssetKind.Pvc]: this.deletePvcAsset,
                [AssetKind.HostPath]: this.deleteHostPathAsset,
                [AssetKind.ConfigMap]: this.deleteConfigMapAsset,
                [AssetKind.SecretVolume]: this.deleteSecretAsset
            };
            await Promise.all(
                dataSources.map((ds: DatasourceListResponseEntry) => {
                    if (dataSourceKindMap[ds.meta.kind]) {
                        return dataSourceKindMap[ds.meta.kind](ds.meta.id, accessToken);
                    } else {
                        throw new Error(`Failed to delete unknown data source kind: ${ds.meta.kind}`);
                    }
                })
            );
            if (dataSources.length > 0) {
                logger.info('deleted all data sources successfully');
            }
        } catch (error) {
            logger.warn(`Failed to delete all data sources: ${error}`);
        }
    }

    async deleteAllPvcs(accessToken = flowRootState.users.editor.accessToken): Promise<void> {
        assetsApi.pVCApi.accessToken = accessToken;
        const pvcAssets = (await assetsApi.pVCApi.listPvcAssets()).body.entries;
        logger.info(`Found ${pvcAssets.length} PVC assets to delete`);
        await Promise.all(pvcAssets.map((pvc) => this.deletePvcAsset(pvc.meta.id, accessToken)));
    }

    waitForDataSourceToBeDeleted(name: string): Promise<void> {
        return waitUtils.waitForCondition(
            async () => {
                const dataSourcesNames = (await this.list()).map((dataSource) => dataSource.meta.name);
                return !dataSourcesNames.includes(name);
            },
            `Waiting for data source: ${name} to be deleted`,
            true,
            30
        );
    }

    getExistingPVCPayload(
        hasClusterWideLabel: boolean,
        departmentId?: string,
        namespace = 'runai'
    ): V1PersistentVolumeClaim {
        let namePrefix = 'pvc';
        const labels = {};
        if (hasClusterWideLabel) labels['run.ai/cluster-wide'] = 'true';
        if (departmentId) {
            labels['run.ai/department'] = departmentId;
            namePrefix += `-dep-${departmentId}`;
        }
        return {
            apiVersion: 'v1',
            kind: 'PersistentVolumeClaim',
            metadata: {
                name: objectGenerator.generateShortNameByPrefix(namePrefix),
                namespace: namespace,
                labels: labels
            },
            spec: {
                accessModes: ['ReadWriteOnce'],
                resources: {
                    requests: {
                        storage: '10M'
                    }
                },
                storageClassName: 'default',
                volumeMode: 'Filesystem'
            }
        };
    }

    async createExistingPVC(
        hasClusterWideLabel: boolean,
        departmentId?: string,
        namespace = 'runai'
    ): Promise<V1ObjectMeta> {
        const secret = this.getExistingPVCPayload(hasClusterWideLabel, departmentId, namespace);
        await kubeCli.CoreV1Api.createNamespacedPersistentVolumeClaim({ namespace, body: secret });
        logger.info(`created pvc named: ${secret.metadata.name}`);
        return secret.metadata;
    }

    async getPvcInfo(name: string, scopeType: string, scopeId: string): Promise<Array<PVCListInfo>> {
        k8sObjectTrackerApi.pvcsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        return (await k8sObjectTrackerApi.pvcsApi.listPvcs(scopeType, scopeId, name)).body.pvcs;
    }

    async getPvcByName(
        name: string,
        accessToken = flowRootState.users.editor.accessToken
    ): Promise<PVCAsset | undefined> {
        assetsApi.pVCApi.accessToken = accessToken;
        const pvcAssets: PVCAsset[] = (await assetsApi.pVCApi.listPvcAssets()).body.entries;
        const pvc = pvcAssets.find((pvc) => pvc.meta.name === name);
        return pvc;
    }

    async getPvcsByProjectId(
        projectId: number,
        workloadType?: 'Workspace' | 'Training' | 'Distributed' | 'Inference'
    ): Promise<Array<PVCAsset>> {
        assetsApi.pVCApi.accessToken = flowRootState.users.system_administrator.accessToken;
        return (
            await assetsApi.pVCApi.listPvcAssets(
                undefined,
                undefined,
                projectId,
                undefined,
                undefined,
                false,
                projectId,
                workloadType
            )
        ).body.entries;
    }

    async createConfigMapResource(name: string, namespace: string): Promise<V1ObjectMeta> {
        name = objectGenerator.generateShortNameByPrefix(name);
        const cmToCreate = new V1ConfigMap();
        cmToCreate.metadata = new V1ObjectMeta();
        cmToCreate.metadata.name = name;
        cmToCreate.metadata.namespace = namespace;
        cmToCreate.metadata.labels = {
            'run.ai/resource': 'resource'
        };
        cmToCreate.data = {
            key1: 'value1',
            key2: 'value2'
        };
        try {
            await kubeCli.CoreV1Api.createNamespacedConfigMap({ namespace, body: cmToCreate });
        } catch (error) {
            logger.error(`failed to create config-map ${name} ns ${namespace} error: [${error}]`);
            throw error;
        }
        return cmToCreate.metadata;
    }

    async getConfigMapInfo(
        name: string,
        scopeType: string,
        scopeId: string
    ): Promise<Array<ConfigMapListInfo>> {
        k8sObjectTrackerApi.configMapsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        return (await k8sObjectTrackerApi.configMapsApi.listConfigMaps(scopeType, scopeId, name)).body
            .configMaps;
    }

    async waitForConfigMapToBeVisible(
        name: string,
        scopeType: string,
        scopeId: string,
        timeout = 60
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const configMaps = await this.getConfigMapInfo(name, scopeType, scopeId);
                    return configMaps.some((configMap) => configMap.name === name);
                } catch (error) {
                    logger.debug(`Failed checking ConfigMap ${name} visibility: ${error}`);
                    return false;
                }
            },
            `ConfigMap ${name} to be visible in ${scopeType} scope ${scopeId}`,
            true,
            timeout,
            2
        );
    }

    async getStorageClassInfo(name: string, clusterId: string): Promise<Array<StorageClass>> {
        k8sObjectTrackerApi.storageClassesApi.accessToken =
            flowRootState.users.system_administrator.accessToken;
        const s = await k8sObjectTrackerApi.storageClassesApi.getStorageClasses(clusterId, name);
        return s.body.items;
    }

    async waitForStorageClassInfo(
        name: string,
        clusterId: string,
        waitDeleted = false,
        predicate: (sc: StorageClass) => boolean = undefined
    ): Promise<StorageClass> {
        const getScs = async (): Promise<StorageClass> => {
            const scs = await this.getStorageClassInfo(name, clusterId);
            return scs && scs.length >= 0 && (!predicate || predicate(scs[0])) ? scs[0] : undefined;
        };

        return (await waitUtils.waitForObject(
            getScs,
            `wait for storage class: ${name}`,
            120,
            1,
            waitDeleted
        )) as StorageClass;
    }

    async createClusterPvc(name: string, namespace: string): Promise<V1PersistentVolumeClaim> {
        const pvcToCreate = new V1PersistentVolumeClaim();
        pvcToCreate.metadata = new V1ObjectMeta();
        pvcToCreate.spec = new V1PersistentVolumeClaimSpec();

        name = objectGenerator.generateShortNameByPrefix(name);

        pvcToCreate.metadata.name = name;
        pvcToCreate.metadata.namespace = namespace;
        pvcToCreate.spec.accessModes = ['ReadWriteOnce'];
        pvcToCreate.spec.resources = new V1ResourceRequirements();
        pvcToCreate.spec.resources.requests = { storage: '8Gi' };
        pvcToCreate.spec.volumeMode = 'Filesystem';

        logger.debug(`Creating pvc ${name} in namesapce ${namespace}`);

        let res;
        try {
            res = await kubeCli.CoreV1Api.createNamespacedPersistentVolumeClaim({
                namespace,
                body: pvcToCreate
            });
        } catch (e) {
            logger.debug(`Failed to create pvc ${name} in namespace ${namespace}, err: ${e}`);
        }

        return res;
    }

    async createClusterConfigMap(name: string, namespace: string): Promise<V1ConfigMap> {
        const cmToCreate = new V1ConfigMap();
        cmToCreate.metadata = new V1ObjectMeta();
        cmToCreate.data = {
            key1: 'value1',
            key2: 'value2'
        };

        name = objectGenerator.generateShortNameByPrefix(name);

        cmToCreate.metadata.name = name;
        cmToCreate.metadata.namespace = namespace;

        logger.debug(`Creating ConfigMap ${name} in namesapce ${namespace}`);

        let res;
        try {
            res = await kubeCli.CoreV1Api.createNamespacedConfigMap({ namespace, body: cmToCreate });
        } catch (e) {
            logger.debug(`Failed to create ConfigMap ${name} in namespace ${namespace}, err: ${e}`);
        }

        return res;
    }

    async waitForDataSourceStatusInTable(
        dsName: string,
        desiredStatus: string,
        page: DataSourcePage
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const dsStatus = await page.findDataSourceStatus(dsName);
                return dsStatus === desiredStatus;
            },
            'waiting for status to update in the table',
            true,
            30
        );
    }

    // ASSETS DELETION UTILS //
    async deleteHostPathAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.hostPathApi.accessToken = accessToken;
            await assetsApi.hostPathApi.deleteHostPathById(id);
            logger.info(`deleted hostPath ${id} successfully`);
        } catch (error) {
            logger.error(`failed to delete hostPath ${id} error: [${error}]`);
            throw error;
        }
    }

    async deleteGitAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.gitApi.accessToken = accessToken;
            await assetsApi.gitApi.deleteGitAssetById(id);
            logger.info(`deleted git ${id} successfully`);
        } catch (error) {
            logger.error(`failed to delete git ${id} error: [${error}]`);
            throw error;
        }
    }

    async deleteS3Asset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        assetsApi.s3Api.accessToken = accessToken;
        await waitUtils.waitForCondition(
            async () => {
                try {
                    await assetsApi.s3Api.deleteS3AssetById(id);
                    logger.info(`Deleted S3 asset ${id} successfully`);
                    return true; // Successfully deleted
                } catch (error: any) {
                    if (error?.body?.code === 412) {
                        // Retry only if the error is 412 (asset in use)
                        logger.warn(`Retrying delete for S3 asset ${id} due to 412 error...`);
                        return false; // Continue waiting
                    } else {
                        logger.error(`Failed to delete S3 asset ${id}, error: [${error}]`);
                        throw error;
                    }
                }
            },
            `Delete S3 asset with ID ${id}`,
            true,
            40,
            2
        );
    }

    async deleteNfsAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.nFSApi.accessToken = accessToken;
            await assetsApi.nFSApi.deleteNfsAssetById(id);
            logger.info(`deleted nfs ${id} successfully`);
        } catch (error) {
            logger.error(`failed to delete nfs ${id} error: [${error}]`);
            throw error;
        }
    }

    async deletePvcAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.pVCApi.accessToken = accessToken;
            await assetsApi.pVCApi.deletePvcAssetById(id);
            logger.info(`deleted pvc ${id} successfully`);
        } catch (error) {
            logger.error(`failed to delete pvc ${id} error: [${error}]`);
            throw error;
        }
    }

    async deleteConfigMapAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        assetsApi.configMapApi.accessToken = accessToken;

        await waitUtils.waitForCondition(
            async () => {
                try {
                    await assetsApi.configMapApi.deleteConfigMapAssetById(id);
                    logger.info(`Deleted ConfigMap asset ${id} successfully`);
                    return true; // Successfully deleted
                } catch (error: any) {
                    if (error?.body?.code === 412) {
                        // Retry only if the error is 412 (asset in use)
                        logger.warn(`Retrying delete for ConfigMap asset ${id} due to 412 error...`);
                        return false; // Continue waiting
                    } else {
                        logger.error(`Failed to delete ConfigMap asset ${id}, error: [${error}]`);
                        throw error;
                    }
                }
            },
            `Delete ConfigMap asset with ID ${id}`,
            true,
            40,
            2
        );
    }

    async deleteSecretAsset(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.secretApi.accessToken = accessToken;
            await assetsApi.secretApi.deleteSecretAssetById(id);
            logger.info(`deleted secret ${id} successfully`);
        } catch (error) {
            logger.error(`failed to delete secret ${id} error: [${error}]`);
            throw error;
        }
    }

    // ASSETS CREATION UTILS //
    async createPvcAsset(
        pvcAsset: PVCCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<PVCAsset> {
        try {
            assetsApi.pVCApi.accessToken = accessToken;
            const res = await assetsApi.pVCApi.createPvcAsset(pvcAsset);
            logger.info(`created pvc ${pvcAsset.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`failed to create pvc ${pvcAsset.meta.name} error: [${error}]`);
            throw error;
        }
    }

    async createPVC(dataSourceName: string, scope: Scope, project: Project): Promise<PVCAsset> {
        let departmentId: number, projectId: number, clusterId: string;
        if (scope === Scope.Department) {
            departmentId = project.departmentId;
            clusterId = undefined;
            projectId = undefined;
        } else if (scope === Scope.Project) {
            projectId = project.id;
            clusterId = flowRootState.cluster.uuid;
            departmentId = undefined;
        } else if (scope === Scope.Cluster) {
            clusterId = flowRootState.cluster.uuid;
        }
        const pvcCreationRequest: PVCCreationRequest = {
            meta: {
                name: dataSourceName,
                scope: scope,
                clusterId: clusterId,
                projectId: projectId,
                departmentId: departmentId?.toString()
            },
            spec: {
                path: '/home',
                claimName: dataSourceName,
                existingPvc: false,
                claimInfo: {
                    size: '1M',
                    accessModes: {
                        readWriteOnce: true
                    },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };
        return await dataSourcesUtils.createPvcAsset(pvcCreationRequest);
    }

    async createGitAsset(
        git: GitCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<GitAsset> {
        try {
            assetsApi.gitApi.accessToken = accessToken;
            const res = await assetsApi.gitApi.createGitAsset(git);
            logger.info(`created git ${git.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`failed to create git ${git.meta.name} error: [${error}]`);
            throw error;
        }
    }

    async createS3Asset(
        s3: S3CreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<S3Asset> {
        try {
            assetsApi.s3Api.accessToken = accessToken;
            const res = await assetsApi.s3Api.createS3Asset(s3);
            logger.info(`created s3 ${s3.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`failed to create s3 ${s3.meta.name} error: [${error}]`);
            throw error;
        }
    }

    async createNfsAsset(
        nfs: NFSCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<NFSAsset> {
        try {
            assetsApi.nFSApi.accessToken = accessToken;
            const res = await assetsApi.nFSApi.createNfsAsset(nfs);
            logger.info(`created nfs ${nfs.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`failed to create nfs ${nfs.meta.name} error: [${error}]`);
            throw error;
        }
    }

    async createHostPathAsset(
        hostPath: HostPathCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<HostPathAsset> {
        try {
            assetsApi.hostPathApi.accessToken = accessToken;
            const res = await assetsApi.hostPathApi.createHostPath(hostPath);
            logger.info(`created hostPath ${hostPath.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`failed to create hostPath ${hostPath.meta.name} error: [${error}]`);
            throw error;
        }
    }

    createHostPathPayload(projectId: number): HostPathCreationRequest {
        return {
            meta: {
                name: objectGenerator.generateNameByPrefix('host-path'),
                scope: Scope.Project,
                projectId: projectId
            },
            spec: {
                path: objectGenerator.generateShortNameByPrefix('/tmp/test'),
                readOnly: false,
                mountPath: objectGenerator.generateShortNameByPrefix('/home')
            }
        };
    }

    async createConfigMapAsset(
        configMap: ConfigMapCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<ConfigMapAsset> {
        try {
            assetsApi.configMapApi.accessToken = accessToken;
            const res = await assetsApi.configMapApi.createConfigMapAsset(configMap);
            logger.info(`created configMap ${configMap.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`failed to create configMap ${configMap.meta.name} error: [${error}]`);
            throw error;
        }
    }

    async createConfigMapWithArbitraryFields(meta: V1ObjectMeta, projectId: number): Promise<ConfigMapAsset> {
        const configMapCreationRequest: ConfigMapCreationRequest = {
            meta: {
                name: `cm-asset-${meta.name}`,
                scope: Scope.Project,
                projectId: projectId
            },
            spec: {
                configMap: meta.name,
                mountPath: objectGenerator.generateNameByPrefix('/var/path')
            }
        };
        return await dataSourcesUtils.createConfigMapAsset(configMapCreationRequest);
    }

    async createSecretAsset(
        secret: SecretAssetCreationRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<SecretAsset> {
        try {
            assetsApi.secretApi.accessToken = accessToken;
            const res = await assetsApi.secretApi.createSecretAsset(secret);
            logger.info(`created secret ${secret.meta.name} successfully`);
            return res.body;
        } catch (error) {
            logger.error(`failed to create secret ${secret.meta.name} error: [${error}]`);
            throw error;
        }
    }

    getAssetMeta(options: AssetMetaOptions): AssetCreationRequest {
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
            scope: options.scope,
            clusterId,
            projectId,
            departmentId
        };
    }

    // ASSETS CREATION PAYLOADS //
    getPvcCreationPayload(options: AssetMetaOptions): PVCCreationRequest {
        return {
            meta: this.getAssetMeta(options),
            spec: {
                path: '/home',
                claimName: options.name,
                existingPvc: false,
                claimInfo: {
                    size: '1M',
                    accessModes: {
                        readWriteOnce: true
                    },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };
    }

    getGitCreationPayload(options: AssetMetaOptions, credentialsId?: string): GitCreationRequest {
        return {
            meta: this.getAssetMeta(options),
            spec: {
                repository: 'https://www.com',
                branch: null,
                revision: null,
                path: '/mnt',
                passwordAssetId: credentialsId
            }
        };
    }

    getNfsCreationPayload(options: AssetMetaOptions): NFSCreationRequest {
        return {
            meta: this.getAssetMeta(options),
            spec: {
                path: '/home',
                server: 'test-server',
                mountPath: '/mnt',
                readOnly: false
            }
        };
    }

    getHostPathCreationPayload(options: AssetMetaOptions): HostPathCreationRequest {
        return {
            meta: this.getAssetMeta(options),
            spec: {
                path: '/user',
                readOnly: false,
                mountPath: '/mnt'
            }
        };
    }

    getS3CreationPayload(options: AssetMetaOptions, credentialsId?: string): S3CreationRequest {
        return {
            meta: this.getAssetMeta(options),
            spec: {
                bucket: 'my-s3',
                path: '/home',
                url: 'https://www.com',
                accessKeyAssetId: credentialsId
            }
        };
    }

    getConfigMapCreationPayload(options: AssetMetaOptions): ConfigMapCreationRequest {
        return {
            meta: this.getAssetMeta(options),
            spec: {
                configMap: options.name,
                mountPath: '/mnt'
            }
        };
    }

    getSecretCreationPayload(options: AssetMetaOptions): SecretAssetCreationRequest {
        return {
            meta: this.getAssetMeta(options),
            spec: {
                credentialAssetId: null,
                mountPath: '/mnt'
            }
        };
    }

    async validateExpectedStatus(dataSourceName: string, expectedStatus: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                assetsApi.datasourcesApi.accessToken = flowRootState.users.editor.accessToken;
                const dataSource = (
                    await assetsApi.datasourcesApi.listDatasourceAssets(
                        dataSourceName,
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        true
                    )
                ).body.entries[0];
                return dataSource?.status?.status.toString() === expectedStatus;
            },
            `waiting for ${dataSourceName} status to be ${expectedStatus}`,
            true
        );
    }

    async verifyDataSourceRW(
        podName: string,
        namespace: string,
        containerPath: string,
        canWrite = true
    ): Promise<void> {
        const dataToWrite = 'test123';
        if (canWrite) {
            // write data
            await kubeCli.execInPod(podName, namespace, [
                '/bin/sh',
                '-c',
                `echo ${dataToWrite} > ${containerPath}/tmp.txt`
            ]);
            // read data
            const dataResult = await kubeCli.execInPod(podName, namespace, [
                'cat',
                `${containerPath}/tmp.txt`
            ]);
            expect(dataResult.stdout).equal(dataToWrite);
        } else {
            // write data（expect failure）
            try {
                await kubeCli.execInPod(podName, namespace, [
                    '/bin/sh',
                    '-c',
                    `echo ${dataToWrite} > ${containerPath}/tmp.txt`
                ]);
            } catch (error) {
                expect(String(error)).contain('command terminated with exit code');
            }
        }
        // delete file
        await kubeCli.execInPod(podName, namespace, ['/bin/sh', '-c', `rm ${containerPath}/tmp.txt`]);
        const outputResult = await kubeCli.execInPod(podName, namespace, [
            '/bin/sh',
            '-c',
            `cat ${containerPath}/tmp.txt || true`
        ]);
        expect(outputResult.stderr).contain('No such file or directory');
    }

    async waitForPvcAssetToBeReady(
        pvcAssetName: string,
        pvcAssetId: string,
        projectId: number,
        workloadType: 'Workspace' | 'Training' | 'Distributed' | 'Inference'
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const pvcAsset = await assetsApi.pVCApi.getPvcAssetById(
                    pvcAssetId,
                    false,
                    projectId,
                    workloadType
                );
                return pvcAsset.body.compliance.compliance === true;
            },
            `waiting for pvc asset ${pvcAssetName} to be ready`,
            true
        );
    }

    async createArbitraryConfigMap(
        dataSourceName: string,
        scope: Scope,
        configMapName: string,
        scopeId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<string> {
        const options: AssetMetaOptions = {
            name: dataSourceName,
            scope: scope,
            scopeId: scopeId
        };

        const configMapCreationRequest = this.getConfigMapCreationPayload(options);
        configMapCreationRequest.spec.mountPath = '/mnt/data';
        configMapCreationRequest.spec.configMap = configMapName;

        const dataSource = await this.createConfigMapAsset(configMapCreationRequest, accessToken);
        return dataSource.meta.id;
    }
}
