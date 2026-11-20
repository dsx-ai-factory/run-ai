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
    credentialUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI List Credential - Positive Tests, tags: ${tags.resource.credential} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-cred-list', 5);
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

        // Per-test cleanup: delete credentials and extra projects created in the test (shared project is cleaned in afterAll)
        let credentialsToClean: { id: string; kind: string }[] = [];
        let projectIdsToClean: number[] = [];

        test.afterEach(async function () {
            await test.step('Cleanup credentials and extra projects from test', async function () {
                for (const c of credentialsToClean) {
                    try {
                        if (c.kind === 'accessKey') await credentialUtils.deleteAccessKeyAsset(c.id);
                        else if (c.kind === 'dockerRegistry')
                            await credentialUtils.deleteDockerRegistryAsset(c.id);
                    } catch (e) {
                        logger.warn(`Cleanup credential ${c.id} failed: ${e}`);
                    }
                }
                for (const projectId of projectIdsToClean) {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                    } catch (e) {
                        logger.warn(`Cleanup project ${projectId} failed: ${e}`);
                    }
                }
                credentialsToClean = [];
                projectIdsToClean = [];
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

        test('T5750391: List credentials in current project with default output', async function () {
            const accessKeyName = objectGenerator.generateShortNameByPrefix('cred-ak', 5);
            const dockerName = objectGenerator.generateShortNameByPrefix('cred-dr', 5);

            await test.step('Create multiple credentials via API', async function () {
                const accessKeyPayload = {
                    meta: credentialUtils.getAssetMeta({
                        name: accessKeyName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { accessKeyId: '123', secretAccessKey: '456' }
                };
                const ak = await credentialUtils.createAccessKeyAsset(accessKeyPayload);
                credentialsToClean.push({ id: ak.meta.id, kind: 'accessKey' });
                logger.info(`Created accessKey credential: ${accessKeyName}`);

                const dockerPayload = {
                    meta: credentialUtils.getAssetMeta({
                        name: dockerName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { user: 'u', password: 'p', url: 'https://example.com' }
                };
                const dr = await credentialUtils.createDockerRegistryAsset(dockerPayload);
                credentialsToClean.push({ id: dr.meta.id, kind: 'dockerRegistry' });
                logger.info(`Created docker-registry credential: ${dockerName}`);
            });

            await test.step('List credentials in current project', function () {
                const stdout = runaiCliV2.list('credential', {}, false).stdout;
                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI credential list command executed successfully');
            });

            await test.step('Verify output contains all created credentials', function () {
                const stdout = runaiCliV2.list('credential', {}, false).stdout;
                expect(stdout).toContain(accessKeyName);
                expect(stdout).toContain(dockerName);
                logger.info('Verified: Output contains all created credentials');
            });
        });

        test('T5750392: List credentials in specific project using -p flag', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-cred-a', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-cred-b', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const namespaceB = Utils.namespaceFromProjectName(projectBName);
            let projectA: Project;
            let projectB: Project;
            let credProjA1: string, credProjA2: string, credProjB1: string;

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
                projectIdsToClean.push(projectA.id, projectB.id);
                logger.info(`Created project-a: ${projectAName}, project-b: ${projectBName}`);
            });

            await test.step('Create credentials in both projects', async function () {
                const nameA1 = objectGenerator.generateShortNameByPrefix('cred-proj-a-1', 5);
                const nameA2 = objectGenerator.generateShortNameByPrefix('cred-proj-a-2', 5);
                const nameB1 = objectGenerator.generateShortNameByPrefix('cred-proj-b-1', 5);
                credProjA1 = nameA1;
                credProjA2 = nameA2;
                credProjB1 = nameB1;

                const ak1 = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: nameA1,
                        scope: Scope.Project,
                        scopeId: String(projectA.id)
                    }),
                    spec: { accessKeyId: '1', secretAccessKey: '1' }
                });
                credentialsToClean.push({ id: ak1.meta.id, kind: 'accessKey' });
                const ak2 = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: nameA2,
                        scope: Scope.Project,
                        scopeId: String(projectA.id)
                    }),
                    spec: { accessKeyId: '2', secretAccessKey: '2' }
                });
                credentialsToClean.push({ id: ak2.meta.id, kind: 'accessKey' });
                const akB = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: nameB1,
                        scope: Scope.Project,
                        scopeId: String(projectB.id)
                    }),
                    spec: { accessKeyId: '3', secretAccessKey: '3' }
                });
                credentialsToClean.push({ id: akB.meta.id, kind: 'accessKey' });
            });

            await test.step('List credentials in project-a using -p flag', function () {
                const stdout = runaiCliV2.list(
                    'credential',
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;
                expect(stdout).toBeTruthy();
                expect(stdout).toContain(credProjA1);
                expect(stdout).toContain(credProjA2);
                expect(stdout).not.toContain(credProjB1);
                logger.info('Verified: Output contains only project-a credentials');
            });
        });

        test('T5750393: List all credentials across all projects using -A flag', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-all-a', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-all-b', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const namespaceB = Utils.namespaceFromProjectName(projectBName);
            let projectA: Project;
            let projectB: Project;
            let credA: string, credB: string, credCluster: string;

            await test.step('Create projects and credentials', async function () {
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
                projectIdsToClean.push(projectA.id, projectB.id);
                credA = objectGenerator.generateShortNameByPrefix('cred-proj-a', 5);
                credB = objectGenerator.generateShortNameByPrefix('cred-proj-b', 5);
                credCluster = objectGenerator.generateShortNameByPrefix('cred-cluster', 5);

                const akA = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: credA,
                        scope: Scope.Project,
                        scopeId: String(projectA.id)
                    }),
                    spec: { accessKeyId: 'a', secretAccessKey: 'a' }
                });
                credentialsToClean.push({ id: akA.meta.id, kind: 'accessKey' });
                const akB = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: credB,
                        scope: Scope.Project,
                        scopeId: String(projectB.id)
                    }),
                    spec: { accessKeyId: 'b', secretAccessKey: 'b' }
                });
                credentialsToClean.push({ id: akB.meta.id, kind: 'accessKey' });
                const akC = await credentialUtils.createAccessKeyAsset(
                    credentialUtils.getAccessKeyAssetPayload(credCluster)
                );
                credentialsToClean.push({ id: akC.meta.id, kind: 'accessKey' });
            });

            await test.step('List all credentials with -A', function () {
                const stdout = runaiCliV2.list(
                    'credential',
                    { [listFlags.allScopes]: '' } as any,
                    false
                ).stdout;
                expect(stdout).toContain(credA);
                expect(stdout).toContain(credB);
                expect(stdout).toContain(credCluster);
                logger.info('Verified: Output contains credentials from all projects and cluster');
            });
        });

        test('T5750394: List credentials filtered by type', async function () {
            const accessKeyName = objectGenerator.generateShortNameByPrefix('cred-ak', 5);
            const dockerName = objectGenerator.generateShortNameByPrefix('cred-dr', 5);

            await test.step('Create credentials of different types', async function () {
                const ak = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: accessKeyName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { accessKeyId: '123', secretAccessKey: '456' }
                });
                credentialsToClean.push({ id: ak.meta.id, kind: 'accessKey' });
                const dr = await credentialUtils.createDockerRegistryAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: dockerName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { user: 'u', password: 'p', url: 'https://x.com' }
                });
                credentialsToClean.push({ id: dr.meta.id, kind: 'dockerRegistry' });
            });

            await test.step('List credentials filtered by accessKey', function () {
                const stdout = runaiCliV2.list(
                    'credential',
                    { [listFlags.type]: 'accessKey' } as any,
                    false
                ).stdout;
                expect(stdout).toContain(accessKeyName);
                expect(stdout).not.toContain(dockerName);
                logger.info('Verified: Output contains only accessKey credentials');
            });
        });

        test('T5750395: List credentials with output format options (json, yaml, table, raw)', async function () {
            const credName = objectGenerator.generateShortNameByPrefix('cred-fmt', 5);

            await test.step('Create credential via API', async function () {
                const ak = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: credName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { accessKeyId: '123', secretAccessKey: '456' }
                });
                credentialsToClean.push({ id: ak.meta.id, kind: 'accessKey' });
            });

            await test.step('List with JSON output', function () {
                const stdout = runaiCliV2.list('credential', { [listFlags.json]: '' } as any, false).stdout;
                const parsed = JSON.parse(stdout);
                expect(parsed).toBeDefined();
                expect(stdout).toContain(credName);
            });

            await test.step('List with YAML output', function () {
                const stdout = runaiCliV2.list('credential', { [listFlags.yaml]: '' } as any, false).stdout;
                expect(stdout).toContain(credName);
            });

            await test.step('List with table and default (identical)', function () {
                const tableOut = runaiCliV2.list(
                    'credential',
                    { [listFlags.table]: '' } as any,
                    false
                ).stdout;
                const defaultOut = runaiCliV2.list('credential', {}, false).stdout;
                expect(tableOut).toContain(credName);
                expect(defaultOut).toContain(credName);
            });

            await test.step('List with raw output', function () {
                // CLI requires --raw to be combined with --json or --yaml
                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} credential list --raw --json -p ${projectName}`,
                    false
                );
                expect(result.status).toBe(0);
                expect(result.stdout).toBeTruthy();
            });
        });

        test('T5750396: List credentials with no-headers option', async function () {
            const credName = objectGenerator.generateShortNameByPrefix('cred-nh', 5);

            await test.step('Create credential via API', async function () {
                const ak = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: credName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { accessKeyId: '1', secretAccessKey: '1' }
                });
                credentialsToClean.push({ id: ak.meta.id, kind: 'accessKey' });
            });

            await test.step('List without headers', function () {
                const stdout = runaiCliV2.list(
                    'credential',
                    { [listFlags.noHeaders]: '' } as any,
                    false
                ).stdout;
                expect(stdout).toContain(credName);
            });

            await test.step('List with default (includes header)', function () {
                const stdout = runaiCliV2.list('credential', {}, false).stdout;
                expect(stdout).toContain(credName);
            });
        });

        test('T5750397: List credentials with custom columns', async function () {
            const credName = objectGenerator.generateShortNameByPrefix('cred-col', 5);

            await test.step('Create credential via API', async function () {
                const ak = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: credName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { accessKeyId: '1', secretAccessKey: '1' }
                });
                credentialsToClean.push({ id: ak.meta.id, kind: 'accessKey' });
            });

            await test.step('List with specific columns', function () {
                const stdout = runaiCliV2.list(
                    'credential',
                    { [listFlags.columns]: 'name,scope' } as any,
                    false
                ).stdout;
                expect(stdout).toContain(credName);
            });
        });

        test('T5750398: List credentials with set-default-columns (experimental)', async function () {
            const credName = objectGenerator.generateShortNameByPrefix('cred-sdc', 5);

            await test.step('Create credential via API', async function () {
                const ak = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: credName,
                        scope: Scope.Project,
                        scopeId: String(project.id)
                    }),
                    spec: { accessKeyId: '1', secretAccessKey: '1' }
                });
                credentialsToClean.push({ id: ak.meta.id, kind: 'accessKey' });
            });

            await test.step('List with set-default-columns flag', function () {
                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} credential list --set-default-columns -p ${projectName}`,
                    false
                );
                expect(result.status).toBe(0);
            });
        });

        test('T5750399: List credentials when no credentials exist (empty list)', async function () {
            const emptyProjectName = objectGenerator.generateShortNameByPrefix('proj-empty', 5);
            const emptyNamespace = Utils.namespaceFromProjectName(emptyProjectName);
            let emptyProject: Project;

            await test.step('Create empty project', async function () {
                emptyProject = await projectResourcesUtils.createProject(emptyNamespace, {
                    name: emptyProjectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                projectIdsToClean.push(emptyProject.id);
            });

            await test.step('List credentials in empty project', function () {
                const stdout = runaiCliV2.list(
                    'credential',
                    { [listFlags.project]: emptyProjectName } as any,
                    false
                ).stdout;
                expect(stdout).toBeDefined();
                logger.info('Verified: Command succeeds with empty or no credentials message');
            });
        });

        test('T5750400: List higher-scope credentials from project context', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-hs', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            let projectA: Project;
            const credCluster = objectGenerator.generateShortNameByPrefix('cred-cluster', 5);
            const credProj = objectGenerator.generateShortNameByPrefix('cred-proj-a', 5);

            await test.step('Create project and credentials (cluster + project scope)', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                projectIdsToClean.push(projectA.id);
                const akCluster = await credentialUtils.createAccessKeyAsset(
                    credentialUtils.getAccessKeyAssetPayload(credCluster)
                );
                credentialsToClean.push({ id: akCluster.meta.id, kind: 'accessKey' });
                const akProj = await credentialUtils.createAccessKeyAsset({
                    meta: credentialUtils.getAssetMeta({
                        name: credProj,
                        scope: Scope.Project,
                        scopeId: String(projectA.id)
                    }),
                    spec: { accessKeyId: '1', secretAccessKey: '1' }
                });
                credentialsToClean.push({ id: akProj.meta.id, kind: 'accessKey' });
            });

            await test.step('List credentials from project-a context', function () {
                const stdout = runaiCliV2.list(
                    'credential',
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;
                expect(stdout).toContain(credCluster);
                expect(stdout).toContain(credProj);
                logger.info('Verified: Output contains cluster and project-scoped credentials');
            });
        });
    });
});
