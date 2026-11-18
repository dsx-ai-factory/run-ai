/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    dataSourcesUtils,
    credentialUtils,
    orgUnitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI Describe Datasource - Positive Tests, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-datasource-cli', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`Deleted project: ${projectName}`);
                }
            });
        });

        test('T5585410: Describe PVC data source with default output', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-pvc-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe PVC data source with CLI', function () {
                logger.info(`Describing PVC data source: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);

                logger.info('CLI describe command executed successfully');
                logger.info('=== CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== CLI Output End ===');
            });

            await test.step('Verify output contains data source details', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc');

                // Verify key fields in output
                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('Pvc'); // Type
                expect(stdout).toContain('Project'); // Scope

                // Verify PVC-specific fields
                expect(stdout).toContain('Claim Name:'); // Claim name
                expect(stdout).toContain('Container Path:'); // Mount path

                logger.info('Verified: Output contains expected data source information');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });
        });

        test('T5585411: Describe NFS data source with default output', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-nfs-ds', 5);
            let dataSourceId: string;

            await test.step('Create NFS data source via API', async function () {
                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.example.com',
                            path: '/data',
                            mountPath: '/mnt/nfs'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = nfsAsset.meta.id;
                logger.info(`Created NFS data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe NFS data source with CLI and print output', function () {
                logger.info(`Describing NFS data source: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'nfs');

                logger.info('=== NFS CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== NFS CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify output contains NFS-specific details', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'nfs');

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('Nfs'); // Type
                expect(stdout).toContain('Project'); // Scope
                expect(stdout).toContain('NFS Specification'); // Section header
                expect(stdout).toContain('Server:'); // Server field
                expect(stdout).toContain('nfs.example.com'); // Server value

                logger.info('Verified: Output contains expected NFS data source information');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted NFS data source: ${dataSourceId}`);
            });
        });

        test('T5585412: Describe S3 data source with default output', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-s3-ds', 5);
            let dataSourceId: string;

            await test.step('Create S3 data source via API', async function () {
                const s3Asset = await dataSourcesUtils.createS3Asset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            url: 'https://s3.amazonaws.com',
                            bucket: 'my-bucket',
                            path: '/mnt/s3'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = s3Asset.meta.id;
                logger.info(`Created S3 data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe S3 data source with CLI and print output', function () {
                logger.info(`Describing S3 data source: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 's3');

                logger.info('=== S3 CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== S3 CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify output contains S3-specific details', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 's3');

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('S3'); // Type
                expect(stdout).toContain('Project'); // Scope
                expect(stdout).toContain('S3 Specification'); // Section header
                expect(stdout).toContain('Bucket:'); // Bucket field
                expect(stdout).toContain('my-bucket'); // Bucket value

                logger.info('Verified: Output contains expected S3 data source information');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteS3Asset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted S3 data source: ${dataSourceId}`);
            });
        });

        test('T5585413: Describe Git data source with default output', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-git-ds', 5);
            let dataSourceId: string;

            await test.step('Create Git data source via API', async function () {
                const gitAsset = await dataSourcesUtils.createGitAsset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            repository: 'https://github.com/example/repo.git',
                            branch: 'main',
                            path: '/mnt/git'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = gitAsset.meta.id;
                logger.info(`Created Git data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe Git data source with CLI and print output', function () {
                logger.info(`Describing Git data source: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'git');

                logger.info('=== Git CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== Git CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify output contains Git-specific details', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'git');

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('Git'); // Type
                expect(stdout).toContain('Project'); // Scope
                expect(stdout).toContain('Git Specification'); // Section header
                expect(stdout).toContain('Repository:'); // Repository field
                expect(stdout).toContain('github.com/example/repo.git'); // Repository value

                logger.info('Verified: Output contains expected Git data source information');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteGitAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted Git data source: ${dataSourceId}`);
            });
        });

        test('T5585414: Describe hostPath data source with default output', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-hostpath-ds', 5);
            let dataSourceId: string;

            await test.step('Create hostPath data source via API', async function () {
                const hostPathAsset = await dataSourcesUtils.createHostPathAsset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            path: '/mnt/data',
                            mountPath: '/data'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = hostPathAsset.meta.id;
                logger.info(`Created hostPath data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe hostPath data source with CLI and print output', function () {
                logger.info(`Describing hostPath data source: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'hostPath');

                logger.info('=== HostPath CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== HostPath CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify output contains hostPath-specific details', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'hostPath');

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('HostPath'); // Type
                expect(stdout).toContain('Project'); // Scope
                expect(stdout).toContain('Host Path Specification'); // Section header
                expect(stdout).toContain('Host Path:'); // Host path field
                expect(stdout).toContain('/mnt/data'); // Host path value

                logger.info('Verified: Output contains expected hostPath data source information');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteHostPathAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted hostPath data source: ${dataSourceId}`);
            });
        });

        test('T5585415: Describe configMap data source with default output', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-configmap-ds', 5);
            let configMapName: string;
            let dataSourceId: string;

            await test.step('Create configMap resource and data source via API', async function () {
                // Create configMap resource first and get the actual name
                const configMapMetadata = await dataSourcesUtils.createConfigMapResource(
                    'app-config',
                    namespace
                );
                configMapName = configMapMetadata.name;
                logger.info(`Created K8s configMap resource: ${configMapName}`);

                const configMapAsset = await dataSourcesUtils.createConfigMapAsset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            configMap: configMapName,
                            mountPath: '/config'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = configMapAsset.meta.id;
                logger.info(`Created configMap data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe configMap data source with CLI and print output', function () {
                logger.info(`Describing configMap data source: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'configMap');

                logger.info('=== ConfigMap CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== ConfigMap CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify output contains configMap-specific details', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'configMap');

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('ConfigMap'); // Type
                expect(stdout).toContain('Project'); // Scope
                expect(stdout).toContain('ConfigMap Specification'); // Section header (no space in "ConfigMap")
                expect(stdout).toContain('Mount Path:'); // Mount path field
                expect(stdout).toContain('ConfigMap Name:'); // ConfigMap name field

                logger.info('Verified: Output contains expected configMap data source information');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteConfigMapAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted configMap data source: ${dataSourceId}`);
            });
        });

        test('T5585416: Describe secretVolume data source with default output', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-secret-ds', 5);
            const secretName = objectGenerator.generateShortNameByPrefix('app-secrets', 5);
            let dataSourceId: string;

            await test.step('Create credential and secret data source via API', async function () {
                // Create a generic secret credential first
                const credentialAsset = await credentialUtils.createGenericSecretAsset({
                    meta: {
                        name: secretName,
                        scope: Scope.Project,
                        projectId: project.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        existingSecretName: null,
                        keyValuePairs: [
                            { key: 'username', value: 'testuser' },
                            { key: 'password', value: 'testpassword' }
                        ]
                    }
                });

                const secretAsset = await dataSourcesUtils.createSecretAsset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            credentialAssetId: credentialAsset.meta.id,
                            mountPath: '/secrets'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = secretAsset.meta.id;
                logger.info(`Created secret data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe secretVolume data source with CLI and print output', function () {
                logger.info(`Describing secretVolume data source: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'secretVolume');

                logger.info('=== SecretVolume CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== SecretVolume CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify output contains secretVolume-specific details', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'secretVolume');

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('Project'); // Scope
                expect(stdout).toContain('Secret Specification'); // Section header
                expect(stdout).toContain('Mount Path:'); // Mount path field
                // Note: SecretVolume output does NOT include a "Type:" field like other data sources

                logger.info('Verified: Output contains expected secretVolume data source information');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteSecretAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted secret data source: ${dataSourceId}`);
            });
        });

        test('T5585418: Describe data source in specific project using -p flag', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-project-ds', 5);
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-ds-a', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            let dataSourceId: string;
            let projectA: Project;

            await test.step('Create project-a via API', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-a: ${projectAName} (ID: ${projectA.id})`);
            });

            await test.step('Create PVC data source in project-a via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, projectA);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source in project-a: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe data source using -p flag', function () {
                logger.info(`Describing data source in project-a using -p flag: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-p': projectAName
                } as any);

                logger.info('=== CLI Output with -p flag Start ===');
                logger.info(stdout);
                logger.info('=== CLI Output with -p flag End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command with -p flag executed successfully');
            });

            await test.step('Verify data source details are displayed', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-p': projectAName
                } as any);

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('Project'); // Scope

                logger.info('Verified: Data source details displayed correctly with -p flag');
            });

            await test.step('Cleanup data source and project', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);

                await operations.deleteProjectAndNamespace(projectA.id);
                logger.info(`Deleted project-a: ${projectAName}`);
            });
        });

        test('T5585419: Describe data source with JSON output format', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-json-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe data source with JSON output', function () {
                logger.info(`Describing data source with JSON output: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-o': 'json'
                } as any);

                logger.info('=== JSON CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== JSON CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command with JSON output executed successfully');
            });

            await test.step('Verify output is valid JSON format', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-o': 'json'
                } as any);

                // Parse JSON to verify it's valid
                let jsonData;
                expect(() => {
                    jsonData = JSON.parse(stdout);
                }).not.toThrow();

                // Verify JSON contains data source metadata
                expect(jsonData).toHaveProperty('meta');
                expect(jsonData.meta).toHaveProperty('name', dataSourceName);
                expect(jsonData).toHaveProperty('spec');

                logger.info('Verified: Output is valid JSON format with data source metadata');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });
        });

        test('T5585420: Describe data source with YAML output format', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-yaml-ds', 5);
            let dataSourceId: string;

            await test.step('Create NFS data source via API', async function () {
                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.example.com',
                            path: '/data',
                            mountPath: '/mnt/nfs'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = nfsAsset.meta.id;
                logger.info(`Created NFS data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe data source with YAML output', function () {
                logger.info(`Describing data source with YAML output: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'nfs', {
                    '-o': 'yaml'
                } as any);

                logger.info('=== YAML CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== YAML CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command with YAML output executed successfully');
            });

            await test.step('Verify output is valid YAML format', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'nfs', {
                    '-o': 'yaml'
                } as any);

                // Verify YAML contains key fields
                expect(stdout).toContain('meta:');
                expect(stdout).toContain('name:');
                expect(stdout).toContain('spec:');
                expect(stdout).toContain(dataSourceName);

                logger.info('Verified: Output is valid YAML format with data source metadata');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted NFS data source: ${dataSourceId}`);
            });
        });

        test('T5585421: Describe data source with table output format (explicit)', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-table-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe data source with explicit table output', function () {
                logger.info(`Describing data source with explicit table output: ${dataSourceName}`);
                const stdoutTable = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-o': 'table'
                } as any);

                expect(stdoutTable).toBeTruthy();
                expect(stdoutTable.length).toBeGreaterThan(0);
                logger.info('CLI describe command with -o table executed successfully');
            });

            await test.step('Compare with default output (no -o flag)', function () {
                const stdoutTable = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-o': 'table'
                } as any);
                const stdoutDefault = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc');

                // Both outputs should be identical
                expect(stdoutTable).toEqual(stdoutDefault);

                logger.info('Verified: Both outputs are identical - table is the default output format');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });
        });

        test('T5585422: Describe cluster-scoped data source from project context', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-cluster-ds', 5);
            let dataSourceId: string;

            await test.step('Create cluster-scoped NFS data source via API', async function () {
                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: dataSourceName,
                            scope: Scope.Cluster,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.cluster.com',
                            path: '/cluster-data',
                            mountPath: '/mnt/cluster'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = nfsAsset.meta.id;
                logger.info(
                    `Created cluster-scoped NFS data source: ${dataSourceName} (ID: ${dataSourceId})`
                );
            });

            await test.step('Describe cluster-scoped data source from project context', function () {
                logger.info(`Describing cluster-scoped data source from project: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'nfs', {
                    '-p': projectName
                } as any);

                logger.info('=== Cluster-scoped CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== Cluster-scoped CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify scope is shown as cluster', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'nfs', {
                    '-p': projectName
                } as any);

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('Cluster'); // Scope

                logger.info('Verified: Scope shows cluster, project can access cluster-scoped data sources');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted cluster-scoped NFS data source: ${dataSourceId}`);
            });
        });

        test('T5585423: Describe department-scoped data source from project', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-dept-ds', 5);
            const departmentName = objectGenerator.generateShortNameByPrefix('dept-datasource', 5);
            const projectDeptName = objectGenerator.generateShortNameByPrefix('proj-in-dept', 5);
            const namespaceDept = Utils.namespaceFromProjectName(projectDeptName);
            let dataSourceId: string;
            let departmentId: number;
            let projectInDept: Project;

            await test.step('Create department via API', async function () {
                const department = await orgUnitUtils.createDepartment({
                    name: departmentName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                departmentId = Number(department.id);
                logger.info(`Created department: ${departmentName} (ID: ${departmentId})`);
            });

            await test.step('Create project in department via API', async function () {
                projectInDept = await projectResourcesUtils.createProject(namespaceDept, {
                    name: projectDeptName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentId
                });
                logger.info(`Created project in department: ${projectDeptName} (ID: ${projectInDept.id})`);
            });

            await test.step('Create department-scoped PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(
                    dataSourceName,
                    Scope.Department,
                    projectInDept
                );
                dataSourceId = pvcAsset.meta.id;
                logger.info(
                    `Created department-scoped PVC data source: ${dataSourceName} (ID: ${dataSourceId})`
                );
            });

            await test.step('Describe data source from project in department', function () {
                logger.info(`Describing department-scoped data source from project: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-p': projectDeptName
                } as any);

                logger.info('=== Department-scoped CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== Department-scoped CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify scope is shown as department', function () {
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-p': projectDeptName
                } as any);

                expect(stdout).toContain(dataSourceName); // Name
                expect(stdout).toContain('Department'); // Scope

                logger.info(
                    'Verified: Scope shows department, project can access department-scoped data sources'
                );
            });

            await test.step('Cleanup data source, project and department', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted department-scoped PVC data source: ${dataSourceId}`);

                await operations.deleteProjectAndNamespace(projectInDept.id);
                logger.info(`Deleted project in department: ${projectDeptName}`);

                await orgUnitUtils.deleteDepartment(departmentId.toString());
                logger.info(`Deleted department: ${departmentName}`);
            });
        });

        test('T5585424: Describe data source in default project without -p flag', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-default-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source in current project via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Describe data source without -p flag', function () {
                logger.info(`Describing data source without -p flag: ${dataSourceName}`);
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                expect(stdout).toContain(dataSourceName);

                logger.info('Verified: Data source is found using default project context');
            });

            await test.step('Compare with explicit -p flag', function () {
                const stdoutDefault = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc');
                const stdoutExplicit = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc', {
                    '-p': projectName
                } as any);

                // Both outputs should be identical
                expect(stdoutDefault).toEqual(stdoutExplicit);

                logger.info('Verified: Both outputs are identical when using default project context');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });
        });

        test('T5585433: Verify datasource describe help output', function () {
            test.step('Execute describe help command and verify output', function () {
                logger.info('Executing datasource describe --help command');
                const stdout = runaiCliV2['bashCli'].exec(
                    `${runaiCliV2['cliPath']} datasource describe --help`
                );

                logger.info('=== Help Output Start ===');
                logger.info(stdout);
                logger.info('=== Help Output End ===');

                // Verify the entire help output contains all expected content
                const expectedContent = [
                    'describe a data source asset',
                    'runai datasource describe',
                    'DATASOURCE_NAME',
                    '--type',
                    'Examples:',
                    '-h, --help',
                    '-o, --output',
                    '-p, --project',
                    'table',
                    'json',
                    'yaml'
                ];

                // Check all expected content is present in the output
                expectedContent.forEach((content) => {
                    expect(stdout).toContain(content);
                });

                logger.info('Verified: Help output contains all expected information');
            });
        });
    });
});
