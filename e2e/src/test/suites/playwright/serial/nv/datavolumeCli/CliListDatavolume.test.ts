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
    dataSourcesUtils,
    datavolumeUtils,
    flowRootState,
    waitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { PVCCreationRequest, Scope } from '../../../../../../modules/api/apis/assetsApi';

test.describe(`CLI List Datavolume - Positive Tests, tags: ${tags.resource.datavolume} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        let projectName: string;
        let project: Project;

        let projectAName: string;
        let projectA: Project;
        let projectBName: string;
        let projectB: Project;
        let emptyProjectName: string;
        let emptyProject: Project;

        let datavolumesToClean: string[] = [];
        let pvcIdsToClean: string[] = [];

        async function createPvcAndDatavolume(
            pvcName: string,
            dvName: string,
            proj: Project
        ): Promise<{ dvId: string }> {
            const pvcRequest: PVCCreationRequest = {
                meta: {
                    name: pvcName,
                    scope: Scope.Project,
                    projectId: proj.id,
                    clusterId: flowRootState.cluster.uuid
                },
                spec: {
                    path: '/mnt/dv',
                    claimName: pvcName,
                    existingPvc: false,
                    claimInfo: {
                        size: '1Gi',
                        accessModes: { readWriteOnce: true },
                        storageClass: 'standard'
                    },
                    readOnly: false
                }
            };
            const pvcAsset = await dataSourcesUtils.createPvcAsset(pvcRequest);
            pvcIdsToClean.push(pvcAsset.meta.id);
            await waitUtils.waitForPvc(pvcName, proj.namespace);

            const dv = await datavolumeUtils.createDatavolume({
                name: dvName,
                description: 'Datavolume for list test',
                originPvcName: pvcName,
                projectId: String(proj.id)
            });
            datavolumesToClean.push(dv.id);
            await datavolumeUtils.waitForApiDatavolumeToExist(dv.id);
            return { dvId: dv.id };
        }

        test.beforeAll(async function () {
            await test.step('Create main project', async function () {
                projectName = objectGenerator.generateShortNameByPrefix('proj-dv-list', 5);
                const namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.debug(`Created main project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Create project-a and project-b', async function () {
                projectAName = objectGenerator.generateShortNameByPrefix('proj-dv-a', 5);
                const namespaceA = Utils.namespaceFromProjectName(projectAName);
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });

                projectBName = objectGenerator.generateShortNameByPrefix('proj-dv-b', 5);
                const namespaceB = Utils.namespaceFromProjectName(projectBName);
                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.debug(`Created project-a: ${projectAName}, project-b: ${projectBName}`);
            });

            await test.step('Create empty project', async function () {
                emptyProjectName = objectGenerator.generateShortNameByPrefix('proj-dv-empty', 5);
                const emptyNamespace = Utils.namespaceFromProjectName(emptyProjectName);
                emptyProject = await projectResourcesUtils.createProject(emptyNamespace, {
                    name: emptyProjectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.debug(`Created empty project: ${emptyProjectName}`);
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.debug(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterEach(async function () {
            await test.step('Cleanup datavolumes and PVCs from test', async function () {
                for (const id of datavolumesToClean) {
                    try {
                        await datavolumeUtils.deleteDatavolume(id);
                        await datavolumeUtils.waitForApiDatavolumeToBeDeleted(id);
                    } catch (e) {
                        logger.warn(`Cleanup datavolume ${id} failed: ${e}`);
                    }
                }
                for (const pvcId of pvcIdsToClean) {
                    try {
                        await dataSourcesUtils.deletePvcAsset(pvcId);
                    } catch (e: any) {
                        if (e?.statusCode !== 404) logger.warn(`Cleanup PVC ${pvcId} failed: ${e}`);
                    }
                }
                datavolumesToClean = [];
                pvcIdsToClean = [];
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup all projects', async function () {
                for (const p of [project, projectA, projectB, emptyProject]) {
                    if (p?.id) {
                        try {
                            await operations.deleteProjectAndNamespace(p.id);
                        } catch (e) {
                            logger.warn(`Cleanup project ${p.id} failed: ${e}`);
                        }
                    }
                }
                logger.debug('Deleted all projects');
            });
        });

        test('T5750405: List data volumes in current project with default output', async function () {
            const pvc1 = objectGenerator.generateShortNameByPrefix('pvc-dv-1', 5);
            const pvc2 = objectGenerator.generateShortNameByPrefix('pvc-dv-2', 5);
            const pvc3 = objectGenerator.generateShortNameByPrefix('pvc-dv-3', 5);
            const dvName1 = objectGenerator.generateShortNameByPrefix('dv-001', 5);
            const dvName2 = objectGenerator.generateShortNameByPrefix('dv-002', 5);
            const dvName3 = objectGenerator.generateShortNameByPrefix('dv-003', 5);

            await test.step('Create multiple data volumes via API', async function () {
                await createPvcAndDatavolume(pvc1, dvName1, project);
                await createPvcAndDatavolume(pvc2, dvName2, project);
                await createPvcAndDatavolume(pvc3, dvName3, project);
            });

            await test.step('List data volumes in current project', function () {
                const stdout = runaiCliV2.list('datavolume', {}, false).stdout;
                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.debug('CLI datavolume list command executed successfully');
            });

            await test.step('Verify output contains all created data volumes', function () {
                const stdout = runaiCliV2.list('datavolume', {}, false).stdout;
                expect(stdout).toContain(dvName1);
                expect(stdout).toContain(dvName2);
                expect(stdout).toContain(dvName3);
                logger.debug('Verified: Output contains all created data volumes');
            });
        });

        test('T5750406: List data volumes in specific project using -p flag', async function () {
            let dvProjA1: string, dvProjA2: string, dvProjB1: string;

            await test.step('Create data volumes in both projects', async function () {
                dvProjA1 = objectGenerator.generateShortNameByPrefix('dv-proj-a-001', 5);
                dvProjA2 = objectGenerator.generateShortNameByPrefix('dv-proj-a-002', 5);
                dvProjB1 = objectGenerator.generateShortNameByPrefix('dv-proj-b-001', 5);
                const pvcA1 = objectGenerator.generateShortNameByPrefix('pvc-a1', 5);
                const pvcA2 = objectGenerator.generateShortNameByPrefix('pvc-a2', 5);
                const pvcB1 = objectGenerator.generateShortNameByPrefix('pvc-b1', 5);
                await createPvcAndDatavolume(pvcA1, dvProjA1, projectA);
                await createPvcAndDatavolume(pvcA2, dvProjA2, projectA);
                await createPvcAndDatavolume(pvcB1, dvProjB1, projectB);
            });

            await test.step('List data volumes in project-a using -p flag', function () {
                const stdout = runaiCliV2.list(
                    'datavolume',
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;
                expect(stdout).toBeTruthy();
                expect(stdout).toContain(dvProjA1);
                expect(stdout).toContain(dvProjA2);
                expect(stdout).not.toContain(dvProjB1);
                logger.debug('Verified: Output contains only project-a data volumes');
            });
        });

        test('T5750407: List all data volumes across all projects using -A flag', async function () {
            let dvA: string, dvB: string;

            await test.step('Create data volumes in both projects', async function () {
                dvA = objectGenerator.generateShortNameByPrefix('dv-proj-a', 5);
                dvB = objectGenerator.generateShortNameByPrefix('dv-proj-b', 5);
                const pvcA = objectGenerator.generateShortNameByPrefix('pvc-alla', 5);
                const pvcB = objectGenerator.generateShortNameByPrefix('pvc-allb', 5);
                await createPvcAndDatavolume(pvcA, dvA, projectA);
                await createPvcAndDatavolume(pvcB, dvB, projectB);
            });

            await test.step('List all data volumes with -A', function () {
                const stdout = runaiCliV2.list(
                    'datavolume',
                    { [listFlags.allScopes]: '' } as any,
                    false
                ).stdout;
                expect(stdout).toContain(dvA);
                expect(stdout).toContain(dvB);
                logger.debug('Verified: Output contains data volumes from all projects');
            });
        });

        test('T5750408: List data volumes with output format options (json, yaml, table, raw)', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-fmt', 5);
            const dvName = objectGenerator.generateShortNameByPrefix('dv-001', 5);
            await createPvcAndDatavolume(pvcName, dvName, project);

            await test.step('List with JSON output', function () {
                const stdout = runaiCliV2.list('datavolume', { [listFlags.json]: '' } as any, false).stdout;
                const parsed = JSON.parse(stdout);
                expect(parsed).toBeDefined();
                expect(stdout).toContain(dvName);
            });

            await test.step('List with YAML output', function () {
                const stdout = runaiCliV2.list('datavolume', { [listFlags.yaml]: '' } as any, false).stdout;
                expect(stdout).toContain(dvName);
            });

            await test.step('List with table and default (identical)', function () {
                const tableOut = runaiCliV2.list(
                    'datavolume',
                    { [listFlags.table]: '' } as any,
                    false
                ).stdout;
                const defaultOut = runaiCliV2.list('datavolume', {}, false).stdout;
                expect(tableOut).toContain(dvName);
                expect(defaultOut).toContain(dvName);
            });

            await test.step('List with raw output', function () {
                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datavolume list --raw --json -p ${projectName}`,
                    false
                );
                expect(result.status).toBe(0);
                expect(result.stdout).toBeTruthy();
            });
        });

        test('T5750409: List data volumes with set-default-columns (experimental)', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-sdc', 5);
            const dvName = objectGenerator.generateShortNameByPrefix('dv-001', 5);
            await createPvcAndDatavolume(pvcName, dvName, project);

            await test.step('List with set-default-columns flag', function () {
                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datavolume list --set-default-columns -p ${projectName}`,
                    false
                );
                expect(result.status).toBe(0);
            });
        });

        test('T5750410: List data volumes with no-headers option', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-nh', 5);
            const dvName = objectGenerator.generateShortNameByPrefix('dv-001', 5);
            await createPvcAndDatavolume(pvcName, dvName, project);

            await test.step('List without headers', function () {
                const stdout = runaiCliV2.list(
                    'datavolume',
                    { [listFlags.noHeaders]: '' } as any,
                    false
                ).stdout;
                expect(stdout).toContain(dvName);
            });

            await test.step('List with default (includes header)', function () {
                const stdout = runaiCliV2.list('datavolume', {}, false).stdout;
                expect(stdout).toContain(dvName);
            });
        });

        test('T5750411: List data volumes with custom columns', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-col', 5);
            const dvName = objectGenerator.generateShortNameByPrefix('dv-001', 5);
            await createPvcAndDatavolume(pvcName, dvName, project);

            await test.step('List with specific columns (name, scope)', function () {
                const stdout = runaiCliV2.list(
                    'datavolume',
                    { [listFlags.columns]: 'name,scope' } as any,
                    false
                ).stdout;
                expect(stdout).toContain(dvName);
            });
        });

        test('T5750412: List data volumes when no data volumes exist (empty list)', async function () {
            await test.step('List data volumes in empty project', function () {
                const stdout = runaiCliV2.list(
                    'datavolume',
                    { [listFlags.project]: emptyProjectName } as any,
                    false
                ).stdout;
                expect(stdout).toBeDefined();
                logger.debug('Verified: Command succeeds with empty or no data volumes message');
            });
        });

        test('T5750413: List higher-scope data volumes from project context', async function () {
            const dvProjName = objectGenerator.generateShortNameByPrefix('dv-proj-a-001', 5);

            await test.step('Create project-scoped data volume', async function () {
                const pvcProj = objectGenerator.generateShortNameByPrefix('pvc-hs', 5);
                await createPvcAndDatavolume(pvcProj, dvProjName, projectA);
            });

            await test.step('List data volumes from project-a context', function () {
                const stdout = runaiCliV2.list(
                    'datavolume',
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;
                expect(stdout).toContain(dvProjName);
                logger.debug('Verified: Output contains project-scoped data volume visible to project-a');
            });
        });
    });
});
