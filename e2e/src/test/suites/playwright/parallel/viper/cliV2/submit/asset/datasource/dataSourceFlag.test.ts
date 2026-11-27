/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';

import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../../../modules/helpers';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    PVCCreationRequest,
    NFSCreationRequest,
    GitCreationRequest,
    S3CreationRequest,
    HostPathCreationRequest,
    ConfigMapCreationRequest,
    SecretAssetCreationRequest,
    PasswordCreationRequest,
    Scope
} from '../../../../../../../../../modules/api/apis/assetsApi';
import { credentialUtils } from '../../../../../../../../../modules/helpers';
import { submitWorkload, validateWorkload, deleteWorkload, StorageValidation } from '../../submitTestUtils';

// Test case configuration for single datasource tests
interface DatasourceTestCase {
    type: string;
    getAssetName: () => string;
    getStorageValidation: () => StorageValidation;
    requiresExternalInfra?: boolean; // Skip if true (NFS, S3, Git need real servers)
}

test.describe(`CLI V2 Submit with Datasource Asset, ${tags.teams.viper}, ${tags.runOn.parallelCI}`, () => {
    let project: Project;
    let workloadName: string;

    // PVC datasource
    let pvcAssetId: string;
    let pvcAssetName: string;
    const pvcMountPath = '/data/pvc';

    // NFS datasources
    let nfsAssetId1: string;
    let nfsAssetName1: string;
    const nfsMountPath1 = '/data/nfs1';
    const nfsServer = 'nfs-server.local';
    const nfsExportPath1 = '/exports/data1';

    let nfsAssetId2: string;
    let nfsAssetName2: string;
    const nfsMountPath2 = '/data/nfs2';
    const nfsExportPath2 = '/exports/data2';

    // Git datasource
    let gitAssetId: string;
    let gitAssetName: string;
    const gitMountPath = '/data/git';
    const gitRepository = 'https://github.com/octocat/Hello-World';
    const gitBranch = 'master';

    // S3 datasource
    let s3AssetId: string;
    let s3AssetName: string;
    const s3MountPath = '/data/s3';
    const s3Bucket = 'test-bucket';
    const s3Url = 'https://s3.amazonaws.com';

    // HostPath datasource
    let hostPathAssetId: string;
    let hostPathAssetName: string;
    const hostPath = '/tmp/hostdata';
    const hostPathMountPath = '/data/hostpath';

    // ConfigMap datasource
    let configMapAssetId: string;
    let configMapAssetName: string;
    let configMapK8sName: string;
    const configMapMountPath = '/data/configmap';

    // SecretVolume datasource
    let secretAssetId: string;
    let secretAssetName: string;
    let credentialId: string;
    const secretMountPath = '/data/secret';

    // Single datasource test cases
    const singleDatasourceTests: DatasourceTestCase[] = [
        {
            type: 'pvc',
            getAssetName: () => pvcAssetName,
            getStorageValidation: () => ({
                pvc: [{ claimName: pvcAssetName, path: pvcMountPath, existingPvc: true, readOnly: false }]
            })
        },
        {
            type: 'nfs',
            getAssetName: () => nfsAssetName1,
            getStorageValidation: () => ({
                nfs: [{ server: nfsServer, path: nfsExportPath1, mountPath: nfsMountPath1, readOnly: false }]
            }),
            requiresExternalInfra: true // Requires NFS server
        },
        {
            type: 'git',
            getAssetName: () => gitAssetName,
            getStorageValidation: () => ({
                git: [{ repository: gitRepository, path: gitMountPath, branch: gitBranch }]
            }),
            requiresExternalInfra: true // Requires accessible git repo
        },
        {
            type: 's3',
            getAssetName: () => s3AssetName,
            getStorageValidation: () => ({
                s3: [{ bucket: s3Bucket, path: s3MountPath, url: s3Url }]
            }),
            requiresExternalInfra: true // Requires S3 bucket
        },
        {
            type: 'hostPath',
            getAssetName: () => hostPathAssetName,
            getStorageValidation: () => ({
                hostPath: [{ path: hostPath, mountPath: hostPathMountPath, readOnly: false }]
            })
        },
        {
            type: 'configMap',
            getAssetName: () => configMapAssetName,
            getStorageValidation: () => ({
                configMapVolume: [{ configMap: configMapK8sName, mountPath: configMapMountPath }]
            })
        },
        {
            type: 'secretVolume',
            getAssetName: () => secretAssetName,
            getStorageValidation: () => ({
                secretVolume: [{ mountPath: secretMountPath }]
            })
        }
    ];

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-datasource');

        // Create PVC datasource
        pvcAssetName = objectGenerator.generateNameByPrefix('pvc');
        const pvcCreationRequest: PVCCreationRequest = {
            meta: {
                name: pvcAssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: pvcMountPath,
                claimName: pvcAssetName,
                existingPvc: false,
                claimInfo: {
                    size: '1Gi',
                    accessModes: { readWriteOnce: true },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };
        const pvcAsset = await dataSourcesUtils.createPvcAsset(pvcCreationRequest);
        pvcAssetId = pvcAsset.meta.id;
        await waitUtils.waitForPvc(pvcAssetName, project.namespace);

        // Create first NFS datasource
        nfsAssetName1 = objectGenerator.generateNameByPrefix('nfs1');
        const nfsCreationRequest1: NFSCreationRequest = {
            meta: {
                name: nfsAssetName1,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: nfsExportPath1,
                server: nfsServer,
                mountPath: nfsMountPath1,
                readOnly: false
            }
        };
        const nfsAsset1 = await dataSourcesUtils.createNfsAsset(nfsCreationRequest1);
        nfsAssetId1 = nfsAsset1.meta.id;

        // Create second NFS datasource
        nfsAssetName2 = objectGenerator.generateNameByPrefix('nfs2');
        const nfsCreationRequest2: NFSCreationRequest = {
            meta: {
                name: nfsAssetName2,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: nfsExportPath2,
                server: nfsServer,
                mountPath: nfsMountPath2,
                readOnly: false
            }
        };
        const nfsAsset2 = await dataSourcesUtils.createNfsAsset(nfsCreationRequest2);
        nfsAssetId2 = nfsAsset2.meta.id;

        // Create Git datasource
        gitAssetName = objectGenerator.generateNameByPrefix('git');
        const gitCreationRequest: GitCreationRequest = {
            meta: {
                name: gitAssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                repository: gitRepository,
                branch: gitBranch,
                revision: null,
                path: gitMountPath
            }
        };
        const gitAsset = await dataSourcesUtils.createGitAsset(gitCreationRequest);
        gitAssetId = gitAsset.meta.id;

        // Create S3 datasource
        s3AssetName = objectGenerator.generateNameByPrefix('s3');
        const s3CreationRequest: S3CreationRequest = {
            meta: {
                name: s3AssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                bucket: s3Bucket,
                path: s3MountPath,
                url: s3Url
            }
        };
        const s3Asset = await dataSourcesUtils.createS3Asset(s3CreationRequest);
        s3AssetId = s3Asset.meta.id;

        // Create HostPath datasource
        hostPathAssetName = objectGenerator.generateNameByPrefix('hostpath');
        const hostPathCreationRequest: HostPathCreationRequest = {
            meta: {
                name: hostPathAssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: hostPath,
                mountPath: hostPathMountPath,
                readOnly: false
            }
        };
        const hostPathAsset = await dataSourcesUtils.createHostPathAsset(hostPathCreationRequest);
        hostPathAssetId = hostPathAsset.meta.id;

        // Create ConfigMap in k8s first, then create ConfigMap datasource asset
        const configMapMeta = await dataSourcesUtils.createConfigMapResource('cm', project.namespace);
        configMapK8sName = configMapMeta.name;
        configMapAssetName = objectGenerator.generateNameByPrefix('configmap');
        const configMapCreationRequest: ConfigMapCreationRequest = {
            meta: {
                name: configMapAssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                configMap: configMapK8sName,
                mountPath: configMapMountPath
            }
        };
        const configMapAsset = await dataSourcesUtils.createConfigMapAsset(configMapCreationRequest);
        configMapAssetId = configMapAsset.meta.id;

        // Create credential first, then create SecretVolume datasource asset
        const passwordCreationRequest: PasswordCreationRequest = {
            meta: {
                name: objectGenerator.generateNameByPrefix('cred'),
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                user: 'testuser',
                password: 'testpassword'
            }
        };
        const credential = await credentialUtils.createPasswordAsset(passwordCreationRequest);
        credentialId = credential.meta.id;

        secretAssetName = objectGenerator.generateNameByPrefix('secret');
        const secretCreationRequest: SecretAssetCreationRequest = {
            meta: {
                name: secretAssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                credentialAssetId: credentialId,
                mountPath: secretMountPath
            }
        };
        const secretAsset = await dataSourcesUtils.createSecretAsset(secretCreationRequest);
        secretAssetId = secretAsset.meta.id;
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl');
    });

    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            if (pvcAssetId) await dataSourcesUtils.deletePvcAsset(pvcAssetId);
            if (nfsAssetId1) await dataSourcesUtils.deleteNfsAsset(nfsAssetId1);
            if (nfsAssetId2) await dataSourcesUtils.deleteNfsAsset(nfsAssetId2);
            if (gitAssetId) await dataSourcesUtils.deleteGitAsset(gitAssetId);
            if (s3AssetId) await dataSourcesUtils.deleteS3Asset(s3AssetId);
            if (hostPathAssetId) await dataSourcesUtils.deleteHostPathAsset(hostPathAssetId);
            if (configMapAssetId) await dataSourcesUtils.deleteConfigMapAsset(configMapAssetId);
            if (secretAssetId) await dataSourcesUtils.deleteSecretAsset(secretAssetId);
            if (credentialId) await credentialUtils.deletePasswordAsset(credentialId);
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    for (const ds of singleDatasourceTests) {
        test(`submit training workload with single ${ds.type} datasource`, async () => {
            submitWorkload(
                workloadName,
                {
                    [SubmitFlagsMerged.datasource]: [`type=${ds.type},name=${ds.getAssetName()}`],
                    [SubmitFlagsMerged.project]: project.name,
                    [SubmitFlagsMerged.image]: 'ubuntu',
                    [SubmitFlagsMerged.command]: '-- sleep infinity'
                } as Partial<typeof SubmitFlagsMerged>,
                cliV2WorkloadType.training
            );

            await validateWorkload(
                workloadName,
                project.id,
                {
                    fields: { image: 'ubuntu' },
                    storage: ds.getStorageValidation()
                },
                cliV2WorkloadType.training
            );

            deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
        });
    }

    test('submit training workload with pvc + nfs + nfs + git datasources', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.datasource]: [
                    `type=pvc,name=${pvcAssetName}`,
                    `type=nfs,name=${nfsAssetName1}`,
                    `type=nfs,name=${nfsAssetName2}`,
                    `type=git,name=${gitAssetName}`
                ],
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'ubuntu',
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: { image: 'ubuntu' },
                storage: {
                    pvc: [
                        { claimName: pvcAssetName, path: pvcMountPath, existingPvc: true, readOnly: false }
                    ],
                    nfs: [
                        {
                            server: nfsServer,
                            path: nfsExportPath1,
                            mountPath: nfsMountPath1,
                            readOnly: false
                        },
                        { server: nfsServer, path: nfsExportPath2, mountPath: nfsMountPath2, readOnly: false }
                    ],
                    git: [{ repository: gitRepository, path: gitMountPath, branch: gitBranch }]
                }
            },
            cliV2WorkloadType.training
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
});
