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
    orgUnitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI List Datasource - Positive Tests, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-ds-list', 5);
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

        test('T5585438: List data sources in current project with default output', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('ds-pvc', 5);
            const nfsName = objectGenerator.generateShortNameByPrefix('ds-nfs', 5);
            const s3Name = objectGenerator.generateShortNameByPrefix('ds-s3', 5);
            const dataSourceIds: string[] = [];

            await test.step('Create multiple data sources via API', async function () {
                // Create PVC
                const pvcAsset = await dataSourcesUtils.createPVC(pvcName, Scope.Project, project);
                dataSourceIds.push(pvcAsset.meta.id);
                logger.info(`Created PVC data source: ${pvcName}`);

                // Create NFS
                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: nfsName,
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
                dataSourceIds.push(nfsAsset.meta.id);
                logger.info(`Created NFS data source: ${nfsName}`);

                // Create S3
                const s3Asset = await dataSourcesUtils.createS3Asset(
                    {
                        meta: {
                            name: s3Name,
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
                dataSourceIds.push(s3Asset.meta.id);
                logger.info(`Created S3 data source: ${s3Name}`);
            });

            await test.step('List data sources in current project', function () {
                logger.info('Listing data sources in current project');
                const stdout = runaiCliV2.list('datasource', {}, false).stdout;

                logger.info('=== List Output Start ===');
                logger.info(stdout);
                logger.info('=== List Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI list command executed successfully');
            });

            await test.step('Verify output contains all created data sources', function () {
                const stdout = runaiCliV2.list('datasource', {}, false).stdout;

                // Verify all data sources are present
                expect(stdout).toContain(pvcName);
                expect(stdout).toContain(nfsName);
                expect(stdout).toContain(s3Name);

                logger.info('Verified: Output contains all three data sources');
            });

            await test.step('Cleanup data sources', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[0],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[1],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteS3Asset(
                    dataSourceIds[2],
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Deleted all data sources');
            });
        });

        test('T5585439: List data sources in specific project using -p flag', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-list-a', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-list-b', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const namespaceB = Utils.namespaceFromProjectName(projectBName);
            let projectA: Project;
            let projectB: Project;
            const dataSourceIds: string[] = [];

            await test.step('Create projects via API', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-a: ${projectAName}`);

                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-b: ${projectBName}`);
            });

            await test.step('Create data sources in both projects', async function () {
                // Create in project-a
                const dsA1 = await dataSourcesUtils.createPVC(
                    objectGenerator.generateShortNameByPrefix('ds-proj-a-1', 5),
                    Scope.Project,
                    projectA
                );
                dataSourceIds.push(dsA1.meta.id);

                const dsA2 = await dataSourcesUtils.createPVC(
                    objectGenerator.generateShortNameByPrefix('ds-proj-a-2', 5),
                    Scope.Project,
                    projectA
                );
                dataSourceIds.push(dsA2.meta.id);

                // Create in project-b
                const dsB1 = await dataSourcesUtils.createPVC(
                    objectGenerator.generateShortNameByPrefix('ds-proj-b-1', 5),
                    Scope.Project,
                    projectB
                );
                dataSourceIds.push(dsB1.meta.id);

                logger.info('Created data sources in both projects');
            });

            await test.step('List data sources in project-a using -p flag', function () {
                logger.info(`Listing data sources in project-a: ${projectAName}`);
                const stdout = runaiCliV2.list(
                    'datasource',
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;

                logger.info('=== List Output for Project-A Start ===');
                logger.info(stdout);
                logger.info('=== List Output for Project-A End ===');

                // Should contain project-a data sources
                expect(stdout).toContain('ds-proj-a');
                // Should NOT contain project-b data sources
                expect(stdout).not.toContain('ds-proj-b');

                logger.info('Verified: Output contains only project-a data sources');
            });

            await test.step('Cleanup data sources and projects', async function () {
                for (const id of dataSourceIds) {
                    await dataSourcesUtils.deletePvcAsset(
                        id,
                        flowRootState.users.system_administrator.idToken
                    );
                }
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                logger.info('Deleted all data sources and projects');
            });
        });

        test('T5585440: List all data sources across all projects using -A flag', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-all-a', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-all-b', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const namespaceB = Utils.namespaceFromProjectName(projectBName);
            let projectA: Project;
            let projectB: Project;
            const dataSourceIds: string[] = [];
            const dsNames: string[] = [];

            await test.step('Create projects via API', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info('Created both projects');
            });

            await test.step('Create data sources in different scopes', async function () {
                // Project-a data source
                const dsA = objectGenerator.generateShortNameByPrefix('ds-proj-a', 5);
                const pvcA = await dataSourcesUtils.createPVC(dsA, Scope.Project, projectA);
                dataSourceIds.push(pvcA.meta.id);
                dsNames.push(dsA);

                // Project-b data source
                const dsB = objectGenerator.generateShortNameByPrefix('ds-proj-b', 5);
                const pvcB = await dataSourcesUtils.createPVC(dsB, Scope.Project, projectB);
                dataSourceIds.push(pvcB.meta.id);
                dsNames.push(dsB);

                // Cluster-scoped data source
                const dsCluster = objectGenerator.generateShortNameByPrefix('ds-cluster', 5);
                const nfsCluster = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: dsCluster,
                            scope: Scope.Cluster,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.cluster.com',
                            path: '/data',
                            mountPath: '/mnt/cluster'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceIds.push(nfsCluster.meta.id);
                dsNames.push(dsCluster);

                logger.info('Created data sources in different scopes');
            });

            await test.step('List all data sources using -A flag', function () {
                logger.info('Listing all data sources across all projects');
                const stdout = runaiCliV2.list(
                    'datasource',
                    { [listFlags.allScopes]: '' } as any,
                    false
                ).stdout;

                logger.info('=== List All Output Start ===');
                logger.info(stdout);
                logger.info('=== List All Output End ===');

                // Verify all data sources are present
                dsNames.forEach((name) => {
                    expect(stdout).toContain(name);
                });

                logger.info('Verified: Output contains data sources from all projects and cluster scope');
            });

            await test.step('Cleanup data sources and projects', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[0],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[1],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[2],
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                logger.info('Deleted all data sources and projects');
            });
        });

        test('T5585441: List data sources filtered by type (PVC)', async function () {
            const dataSourceIds: string[] = [];
            const pvcNames: string[] = [];

            await test.step('Create multiple data sources of different types', async function () {
                // Create PVC data sources
                const pvc1 = objectGenerator.generateShortNameByPrefix('ds-pvc-1', 5);
                const pvcAsset1 = await dataSourcesUtils.createPVC(pvc1, Scope.Project, project);
                dataSourceIds.push(pvcAsset1.meta.id);
                pvcNames.push(pvc1);

                const pvc2 = objectGenerator.generateShortNameByPrefix('ds-pvc-2', 5);
                const pvcAsset2 = await dataSourcesUtils.createPVC(pvc2, Scope.Project, project);
                dataSourceIds.push(pvcAsset2.meta.id);
                pvcNames.push(pvc2);

                // Create NFS data source
                const nfsName = objectGenerator.generateShortNameByPrefix('ds-nfs', 5);
                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: nfsName,
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
                dataSourceIds.push(nfsAsset.meta.id);

                logger.info('Created PVC and NFS data sources');
            });

            await test.step('List only PVC data sources', function () {
                logger.info('Listing only PVC data sources');
                const stdout = runaiCliV2.list(
                    'datasource',
                    { [listFlags.type]: 'pvc' } as any,
                    false
                ).stdout;

                logger.info('=== List PVC Output Start ===');
                logger.info(stdout);
                logger.info('=== List PVC Output End ===');

                // Should contain PVC data sources
                pvcNames.forEach((name) => {
                    expect(stdout).toContain(name);
                });

                // Should NOT contain NFS data source
                expect(stdout).not.toContain('ds-nfs');

                logger.info('Verified: Output contains only PVC data sources');
            });

            await test.step('Cleanup data sources', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[0],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[1],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[2],
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Deleted all data sources');
            });
        });

        test('T5585442: List data sources filtered by type (NFS)', async function () {
            const dataSourceIds: string[] = [];
            const nfsNames: string[] = [];

            await test.step('Create multiple data sources of different types', async function () {
                // Create PVC data source
                const pvcName = objectGenerator.generateShortNameByPrefix('ds-pvc', 5);
                const pvcAsset = await dataSourcesUtils.createPVC(pvcName, Scope.Project, project);
                dataSourceIds.push(pvcAsset.meta.id);

                // Create NFS data sources
                const nfs1 = objectGenerator.generateShortNameByPrefix('ds-nfs-1', 5);
                const nfsAsset1 = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: nfs1,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs1.example.com',
                            path: '/data1',
                            mountPath: '/mnt/nfs1'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceIds.push(nfsAsset1.meta.id);
                nfsNames.push(nfs1);

                const nfs2 = objectGenerator.generateShortNameByPrefix('ds-nfs-2', 5);
                const nfsAsset2 = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: nfs2,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs2.example.com',
                            path: '/data2',
                            mountPath: '/mnt/nfs2'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceIds.push(nfsAsset2.meta.id);
                nfsNames.push(nfs2);

                logger.info('Created PVC and NFS data sources');
            });

            await test.step('List only NFS data sources', function () {
                logger.info('Listing only NFS data sources');
                const stdout = runaiCliV2.list(
                    'datasource',
                    { [listFlags.type]: 'nfs' } as any,
                    false
                ).stdout;

                logger.info('=== List NFS Output Start ===');
                logger.info(stdout);
                logger.info('=== List NFS Output End ===');

                // Should contain NFS data sources
                nfsNames.forEach((name) => {
                    expect(stdout).toContain(name);
                });

                // Should NOT contain PVC data source
                expect(stdout).not.toContain('ds-pvc');

                logger.info('Verified: Output contains only NFS data sources');
            });

            await test.step('Cleanup data sources', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[0],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[1],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[2],
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Deleted all data sources');
            });
        });

        // Output Format Tests - Shared Resources
        // These tests only verify output format, so they can share the same data sources
        test.describe('Output Format Tests - Shared Resources', () => {
            const sharedDataSourceIds: string[] = [];
            const sharedDsNames: string[] = [];

            test.beforeAll(async function () {
                await test.step('Create shared data sources for output format tests', async function () {
                    const pvcName = objectGenerator.generateShortNameByPrefix('ds-format-pvc', 5);
                    const pvcAsset = await dataSourcesUtils.createPVC(pvcName, Scope.Project, project);
                    sharedDataSourceIds.push(pvcAsset.meta.id);
                    sharedDsNames.push(pvcName);

                    const nfsName = objectGenerator.generateShortNameByPrefix('ds-format-nfs', 5);
                    const nfsAsset = await dataSourcesUtils.createNfsAsset(
                        {
                            meta: {
                                name: nfsName,
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
                    sharedDataSourceIds.push(nfsAsset.meta.id);
                    sharedDsNames.push(nfsName);

                    logger.info('Created shared PVC and NFS data sources for output format tests');
                });
            });

            test.afterAll(async function () {
                await test.step('Cleanup shared data sources', async function () {
                    if (sharedDataSourceIds.length > 0) {
                        await dataSourcesUtils.deletePvcAsset(
                            sharedDataSourceIds[0],
                            flowRootState.users.system_administrator.idToken
                        );
                        await dataSourcesUtils.deleteNfsAsset(
                            sharedDataSourceIds[1],
                            flowRootState.users.system_administrator.idToken
                        );
                        logger.info('Deleted shared data sources for output format tests');
                    }
                });
            });

            test('T5585443: List data sources with JSON output format', async function () {
                await test.step('List data sources with JSON output', function () {
                    logger.info('Listing data sources with JSON output (using shared resources)');
                    const stdout = runaiCliV2.list(
                        'datasource',
                        { [listFlags.json]: '' } as any,
                        false
                    ).stdout;

                    logger.info('=== List JSON Output Start ===');
                    logger.info(stdout);
                    logger.info('=== List JSON Output End ===');

                    // Parse JSON to verify it's valid
                    const jsonData = JSON.parse(stdout);

                    logger.info(`JSON structure: ${JSON.stringify(jsonData, null, 2)}`);

                    // Based on YAML output structure, JSON should have datasources array
                    expect(jsonData.datasources).toBeDefined();
                    expect(Array.isArray(jsonData.datasources)).toBe(true);
                    expect(jsonData.datasources.length).toBeGreaterThanOrEqual(2);

                    // Verify data source names are present
                    sharedDsNames.forEach((name) => {
                        const found = jsonData.datasources.some((ds: any) => ds.name === name);
                        expect(found).toBe(true);
                    });

                    logger.info('Verified: Output is valid JSON format with data sources');
                });
            });

            test('T5585444: List data sources with YAML output format', async function () {
                await test.step('List data sources with YAML output', function () {
                    logger.info('Listing data sources with YAML output (using shared resources)');
                    const stdout = runaiCliV2.list(
                        'datasource',
                        { [listFlags.yaml]: '' } as any,
                        false
                    ).stdout;

                    logger.info('=== List YAML Output Start ===');
                    logger.info(stdout);
                    logger.info('=== List YAML Output End ===');

                    // Verify YAML contains key fields based on actual output structure
                    expect(stdout).toContain('datasources:');
                    expect(stdout).toContain('name:');
                    expect(stdout).toContain('type:');
                    expect(stdout).toContain('scope:');

                    // Verify data source names are present
                    sharedDsNames.forEach((name) => {
                        expect(stdout).toContain(name);
                    });

                    logger.info('Verified: Output is valid YAML format with data sources');
                });
            });

            test('T5585445: List data sources with table output format (explicit)', async function () {
                await test.step('Compare table output with default output', function () {
                    logger.info('Comparing table output with default (using shared resources)');
                    const stdoutTable = runaiCliV2.list(
                        'datasource',
                        { [listFlags.table]: '' } as any,
                        false
                    ).stdout;
                    const stdoutDefault = runaiCliV2.list('datasource', {}, false).stdout;

                    // Both outputs should be identical
                    expect(stdoutTable).toEqual(stdoutDefault);

                    logger.info('Verified: Both outputs are identical - table is the default output format');
                });
            });

            test('T5585446: List data sources with no-headers option', async function () {
                await test.step('List with no-headers and compare with default', function () {
                    logger.info('Testing no-headers option (using shared resources)');
                    const stdoutNoHeaders = runaiCliV2.list(
                        'datasource',
                        { [listFlags.noHeaders]: '' } as any,
                        false
                    ).stdout;
                    const stdoutDefault = runaiCliV2.list('datasource', {}, false).stdout;

                    logger.info('=== No-Headers Output Start ===');
                    logger.info(stdoutNoHeaders);
                    logger.info('=== No-Headers Output End ===');

                    logger.info('=== Default Output Start ===');
                    logger.info(stdoutDefault);
                    logger.info('=== Default Output End ===');

                    // No-headers should be shorter than default (no header row)
                    expect(stdoutNoHeaders.split('\n').length).toBeLessThan(stdoutDefault.split('\n').length);

                    logger.info('Verified: No-headers output displays data without header row');
                });
            });
        });

        test('T5585447: List data sources when no data sources exist (empty list)', function () {
            test.step('List data sources in empty project', function () {
                logger.info('Listing data sources in project with no data sources');
                const stdout = runaiCliV2.list('datasource', {}, false).stdout;

                logger.info('=== Empty List Output Start ===');
                logger.info(stdout);
                logger.info('=== Empty List Output End ===');

                expect(stdout).toBeTruthy();
                logger.info('Verified: Command exits successfully with empty or no data message');
            });
        });

        test('T5585448: List cluster-scoped data sources from project context', async function () {
            const dataSourceIds: string[] = [];
            const dsNames: string[] = [];

            await test.step('Create cluster and project-scoped data sources', async function () {
                // Cluster-scoped
                const clusterDS1 = objectGenerator.generateShortNameByPrefix('ds-cluster-1', 5);
                const nfs1 = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: clusterDS1,
                            scope: Scope.Cluster,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.cluster1.com',
                            path: '/data1',
                            mountPath: '/mnt/cluster1'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceIds.push(nfs1.meta.id);
                dsNames.push(clusterDS1);

                const clusterDS2 = objectGenerator.generateShortNameByPrefix('ds-cluster-2', 5);
                const nfs2 = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: clusterDS2,
                            scope: Scope.Cluster,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.cluster2.com',
                            path: '/data2',
                            mountPath: '/mnt/cluster2'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceIds.push(nfs2.meta.id);
                dsNames.push(clusterDS2);

                // Project-scoped
                const projDS = objectGenerator.generateShortNameByPrefix('ds-proj', 5);
                const pvc = await dataSourcesUtils.createPVC(projDS, Scope.Project, project);
                dataSourceIds.push(pvc.meta.id);
                dsNames.push(projDS);

                logger.info('Created cluster and project-scoped data sources');
            });

            await test.step('List from project context', function () {
                logger.info('Listing data sources from project context');
                const stdout = runaiCliV2.list(
                    'datasource',
                    { [listFlags.project]: projectName } as any,
                    false
                ).stdout;

                logger.info('=== List with Cluster Scope Output Start ===');
                logger.info(stdout);
                logger.info('=== List with Cluster Scope Output End ===');

                // Should contain both cluster and project data sources
                dsNames.forEach((name) => {
                    expect(stdout).toContain(name);
                });

                logger.info('Verified: Output contains both cluster-scoped and project-scoped data sources');
            });

            await test.step('Cleanup data sources', async function () {
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[0],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[1],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[2],
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Deleted all data sources');
            });
        });

        test('T5585449: List department-scoped data sources from project in department', async function () {
            const departmentName = objectGenerator.generateShortNameByPrefix('dept-list', 5);
            const projectDeptName = objectGenerator.generateShortNameByPrefix('proj-in-dept', 5);
            const namespaceDept = Utils.namespaceFromProjectName(projectDeptName);
            let departmentId: number;
            let projectInDept: Project;
            const dataSourceIds: string[] = [];
            const dsNames: string[] = [];

            await test.step('Create department and project', async function () {
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
                logger.info(`Created department: ${departmentName}`);

                projectInDept = await projectResourcesUtils.createProject(namespaceDept, {
                    name: projectDeptName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentId
                });
                logger.info(`Created project in department: ${projectDeptName}`);
            });

            await test.step('Create department and project-scoped data sources', async function () {
                // Department-scoped
                const deptDS = objectGenerator.generateShortNameByPrefix('ds-dept', 5);
                const pvcDept = await dataSourcesUtils.createPVC(deptDS, Scope.Department, projectInDept);
                dataSourceIds.push(pvcDept.meta.id);
                dsNames.push(deptDS);

                // Project-scoped
                const projDS = objectGenerator.generateShortNameByPrefix('ds-proj', 5);
                const pvcProj = await dataSourcesUtils.createPVC(projDS, Scope.Project, projectInDept);
                dataSourceIds.push(pvcProj.meta.id);
                dsNames.push(projDS);

                logger.info('Created department and project-scoped data sources');
            });

            await test.step('List from project in department', function () {
                logger.info('Listing data sources from project in department');
                const stdout = runaiCliV2.list(
                    'datasource',
                    { [listFlags.project]: projectDeptName } as any,
                    false
                ).stdout;

                logger.info('=== List with Department Scope Output Start ===');
                logger.info(stdout);
                logger.info('=== List with Department Scope Output End ===');

                // Should contain both department and project data sources
                dsNames.forEach((name) => {
                    expect(stdout).toContain(name);
                });

                logger.info(
                    'Verified: Output contains both department-scoped and project-scoped data sources'
                );
            });

            await test.step('Cleanup data sources, project and department', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[0],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[1],
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectInDept.id);
                await orgUnitUtils.deleteDepartment(departmentId.toString());
                logger.info('Deleted all data sources, project and department');
            });
        });

        test('T5585450: List data sources with combined filters (type + project)', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-filter', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            let projectA: Project;
            const dataSourceIds: string[] = [];
            const pvcName = objectGenerator.generateShortNameByPrefix('ds-pvc', 5);
            const nfsName = objectGenerator.generateShortNameByPrefix('ds-nfs', 5);

            await test.step('Create project via API', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project: ${projectAName}`);
            });

            await test.step('Create PVC and NFS data sources', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(pvcName, Scope.Project, projectA);
                dataSourceIds.push(pvcAsset.meta.id);

                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: nfsName,
                            scope: Scope.Project,
                            projectId: projectA.id,
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
                dataSourceIds.push(nfsAsset.meta.id);

                logger.info('Created PVC and NFS data sources');
            });

            await test.step('List only PVC data sources in project', function () {
                logger.info('Listing only PVC data sources in specific project');
                const stdout = runaiCliV2.list(
                    'datasource',
                    { [listFlags.project]: projectAName, [listFlags.type]: 'pvc' } as any,
                    false
                ).stdout;

                logger.info('=== List with Combined Filters Output Start ===');
                logger.info(stdout);
                logger.info('=== List with Combined Filters Output End ===');

                // Should contain PVC only
                expect(stdout).toContain(pvcName);
                // Should NOT contain NFS
                expect(stdout).not.toContain(nfsName);

                logger.info('Verified: Output contains only PVC data sources from specified project');
            });

            await test.step('Cleanup data sources and project', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceIds[0],
                    flowRootState.users.system_administrator.idToken
                );
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceIds[1],
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectA.id);
                logger.info('Deleted all data sources and project');
            });
        });

        test('T5585455: Verify datasource list help output', function () {
            test.step('Execute list help command and verify output', function () {
                logger.info('Executing datasource list --help command');
                const stdout = runaiCliV2.datasourceHelp('list --help');

                logger.info('=== List Help Output Start ===');
                logger.info(stdout);
                logger.info('=== List Help Output End ===');

                // Verify the entire help output contains all expected content
                const expectedContent = [
                    'list data source assets',
                    'runai datasource list',
                    'Examples:',
                    '-A, --all',
                    '--columns',
                    '-h, --help',
                    '--json',
                    '--no-headers',
                    '-p, --project',
                    '--table',
                    '--type',
                    '--yaml'
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
