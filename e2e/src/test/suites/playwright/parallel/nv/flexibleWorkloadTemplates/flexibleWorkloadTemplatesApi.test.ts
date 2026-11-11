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
    templateUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    ScopeType,
    SupersetTemplate,
    WorkloadType
} from '../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(`Flexible Workload Template API Tests, tags: ${tags.tool.api} ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectNameList: Array<string>;
        let namespaceList: Array<string>;
        let projectIdList: Array<number>;
        let templateId: string;
        let secondTemplateId: string;
        let workloadType: WorkloadType;
        let templateCreated: Array<string> = [];
        type SortOrder = 'asc' | 'desc';
        type SortBy =
            | 'name'
            | 'workloadType'
            | 'runnable'
            | 'scopeType'
            | 'clusterId'
            | 'projectId'
            | 'departmentId'
            | 'createdAt'
            | 'updatedAt'
            | 'createdBy'
            | 'updatedBy';

        const alpineImage = 'alpine:latest';
        const ubuntuImage = 'ubuntu:latest';
        const sleep10s = 'sleep 10';
        const sleep20s = 'sleep 20';
        const sleep30s = 'sleep 30';
        const cpuOnly = {
            gpuDevicesRequest: 0,
            cpuCoreRequest: 0.1,
            cpuMemoryRequest: '100M'
        };
        const oneGpu = {
            gpuDevicesRequest: 1,
            cpuCoreRequest: 0.1,
            cpuMemoryRequest: '100M'
        };

        const creationSpec = {
            image: alpineImage,
            imagePullPolicy: 'IfNotPresent',
            backoffLimit: 6,
            command: sleep10s,
            compute: cpuOnly
        };

        const distMPISpec = {
            distributedFramework: 'MPI',
            numWorkers: 1,
            image: alpineImage,
            imagePullPolicy: 'IfNotPresent',
            backoffLimit: 6,
            command: sleep10s,
            compute: cpuOnly
        };

        const updatedSpec = {
            command: sleep20s,
            image: alpineImage,
            imagePullPolicy: 'IfNotPresent',
            backoffLimit: 6,
            compute: oneGpu
        };

        const updatedDistMPISpec = {
            distributedFramework: 'MPI',
            numWorkers: 1,
            command: sleep20s,
            image: alpineImage,
            imagePullPolicy: 'IfNotPresent',
            backoffLimit: 6,
            compute: oneGpu
        };

        const ubuntuSpec = {
            image: ubuntuImage,
            imagePullPolicy: 'IfNotPresent',
            backoffLimit: 6,
            command: sleep10s,
            compute: oneGpu
        };

        const distPtSpec = {
            ...distMPISpec,
            distributedFramework: 'PyTorch'
        };

        const updatedDistPtSpec = {
            ...updatedDistMPISpec,
            distributedFramework: 'PyTorch'
        };

        const distTfSpec = {
            ...distMPISpec,
            distributedFramework: 'TF'
        };

        const updatedDistTfSpec = {
            ...updatedDistMPISpec,
            distributedFramework: 'TF'
        };

        const distXGBoostSpec = {
            ...distMPISpec,
            distributedFramework: 'XGBoost'
        };

        const updatedDistXGBoostSpec = {
            ...updatedDistMPISpec,
            distributedFramework: 'XGBoost'
        };

        const distJAXSpec = {
            ...distMPISpec,
            distributedFramework: 'JAX'
        };

        const updatedDistJAXSpec = {
            ...updatedDistMPISpec,
            distributedFramework: 'JAX'
        };

        test.beforeAll(async function () {
            projectNameList = Array.from({ length: 2 }, () =>
                objectGenerator.generateShortNameByPrefix('proj-flex-template', 5)
            );
            namespaceList = projectNameList.map((name) => Utils.namespaceFromProjectName(name));

            const projectsToCreate = [
                { namespace: namespaceList[0], name: projectNameList[0] },
                { namespace: namespaceList[1], name: projectNameList[1] }
            ];
            const projResponse = await Promise.all(
                projectsToCreate.map((config) =>
                    projectResourcesUtils.createProject(config.namespace, {
                        name: config.name,
                        deservedGpus: 0,
                        maxAllowedGpus: 0
                    })
                )
            );
            projectIdList = projResponse.map((proj) => proj.id);
        });

        test.afterAll('deleting project and namespace', async () => {
            await Promise.all(
                projectIdList.map(async (projectId, i) => {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                        logger.info(
                            `deleted project ${projectNameList[i]} and namespace ${namespaceList[i]}`
                        );
                    } catch (error) {
                        logger.error(
                            `failed to delete project ${projectNameList[i]} and namespace ${namespaceList[i]}`,
                            error
                        );
                    }
                })
            );
        });

        function isSorted(entries: Array<SupersetTemplate>, sortOrder: SortOrder, sortBy: SortBy): boolean {
            const getValue = (item: SupersetTemplate): any => item.meta[sortBy];

            if (sortOrder === 'asc') {
                return entries.every(
                    (item, idx, array) => idx === 0 || getValue(array[idx - 1]) <= getValue(item)
                );
            } else {
                return entries.every(
                    (item, idx, array) => idx === 0 || getValue(array[idx - 1]) >= getValue(item)
                );
            }
        }

        test.describe('Test Different types of templates CRUD operations', function () {
            let templateName: string;
            let updatedTemplateName: string;
            let patchedTemplateName: string;
            let secondTemplateName: string;
            let description: string;

            test.beforeEach(function () {
                templateName = objectGenerator.generateShortNameByPrefix('alpine-templ', 5);
                updatedTemplateName = `${templateName}-2`;
                patchedTemplateName = `${templateName}-3`;
                secondTemplateName = objectGenerator.generateShortNameByPrefix('alpine-templ', 5);
                description = objectGenerator.generateShortNameByPrefix('alpine-templ-descripton', 5);
            });

            test.afterEach(async function () {
                if (templateCreated.length !== 0) {
                    await Promise.all(
                        templateCreated.map(async (templId) => {
                            try {
                                await templateUtils.deleteWorkloadTemplateV2(templId, workloadType);
                            } catch (error) {
                                logger.error(error);
                            }
                        })
                    );
                }
            });

            const dataSets = [
                {
                    testCase: 'T5193623: test standard training template CRUD operations',
                    workloadType: WorkloadType.Training,
                    workerSpec: creationSpec,
                    updatedSpec: updatedSpec,
                    workersOnly: undefined,
                    testEnable: true
                },
                {
                    testCase: 'T5193624: test distributed MPI training template CRUD operations',
                    workloadType: WorkloadType.Distributed,
                    workerSpec: distMPISpec,
                    updatedSpec: updatedDistMPISpec,
                    workersOnly: undefined,
                    testEnable: true
                },
                {
                    testCase: 'T5193625: test workspace template CRUD operations',
                    workloadType: WorkloadType.Workspace,
                    workerSpec: creationSpec,
                    updatedSpec: updatedSpec,
                    workersOnly: undefined,
                    testEnable: true
                },
                {
                    testCase:
                        'T5197476: test distributed PyTorch Workers & masters training template CRUD operations',
                    workloadType: WorkloadType.Distributed,
                    workerSpec: distPtSpec,
                    updatedSpec: updatedDistPtSpec,
                    workersOnly: undefined,
                    testEnable: true
                },
                {
                    testCase:
                        'T5197477: test distributed PyTorch Workers only training template CRUD operations',
                    workloadType: WorkloadType.Distributed,
                    workerSpec: distPtSpec,
                    updatedSpec: updatedDistPtSpec,
                    workersOnly: true,
                    testEnable: true
                },
                {
                    testCase:
                        'T5197478: test distributed TensorFlow Workers & masters training template CRUD operations',
                    workloadType: WorkloadType.Distributed,
                    workerSpec: distTfSpec,
                    updatedSpec: updatedDistTfSpec,
                    workersOnly: undefined,
                    testEnable: true
                },
                {
                    testCase:
                        'T5197479: test distributed TensorFlow Workers only training template CRUD operations',
                    workloadType: WorkloadType.Distributed,
                    workerSpec: distTfSpec,
                    updatedSpec: updatedDistTfSpec,
                    workersOnly: undefined,
                    testEnable: true
                },
                {
                    testCase: 'T5197480: test distributed XGBoost training template CRUD operations',
                    workloadType: WorkloadType.Distributed,
                    workerSpec: distXGBoostSpec,
                    updatedSpec: updatedDistXGBoostSpec,
                    workersOnly: undefined,
                    testEnable: true
                },
                {
                    testCase: 'T5197481: test distributed JAX training template CRUD operations',
                    workloadType: WorkloadType.Distributed,
                    workerSpec: distJAXSpec,
                    updatedSpec: updatedDistJAXSpec,
                    workersOnly: true,
                    testEnable: true
                }
            ];

            dataSets.forEach((caseData) => {
                if (!caseData.testEnable) return;

                test(`${caseData.testCase}`, async function () {
                    workloadType = caseData.workloadType;

                    await test.step(`use L1 researcher to create ${caseData.workloadType} template`, async function () {
                        const response = await templateUtils.createWorkloadTemplateV2(
                            {
                                workloadType: caseData.workloadType,
                                scopeType: ScopeType.Cluster,
                                scopeId: flowRootState.cluster.uuid,
                                templateName: templateName,
                                description: description
                            },
                            caseData.workerSpec,
                            caseData.workersOnly ? undefined : { masterSpecSameAsWorker: true },
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);
                        templateId = response.body.meta.id;
                        templateCreated.push(templateId);
                    });

                    await test.step('get the template and verify the related fields', async function () {
                        const response = await templateUtils.getWorkloadTemplateV2(
                            templateId,
                            caseData.workloadType,
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);

                        expect(response.body.meta.name).toEqual(templateName);
                        expect(response.body.meta.description).toEqual(description);
                        expect(response.body.meta.scopeType).toEqual(ScopeType.Cluster);
                        expect(response.body.meta.scopeId).toEqual(flowRootState.cluster.uuid);

                        expect(response.body.spec.image).toEqual(caseData.workerSpec.image);
                        expect(response.body.spec.command).toEqual(caseData.workerSpec.command);
                        expect(response.body.spec.compute).toEqual(caseData.workerSpec.compute);
                    });

                    await test.step('update the template', async function () {
                        const response = await templateUtils.updateWorkloadTemplateV2(
                            templateId,
                            caseData.workloadType,
                            {
                                name: updatedTemplateName
                            },
                            caseData.updatedSpec,
                            caseData.workersOnly ? undefined : { masterSpecSameAsWorker: true },
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);
                    });

                    await test.step('get the updated template and verify the related fields', async function () {
                        const response = await templateUtils.getWorkloadTemplateV2(
                            templateId,
                            caseData.workloadType,
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);

                        expect(response.body.meta.name).toEqual(updatedTemplateName);
                        expect(response.body.meta.scopeType).toEqual(ScopeType.Cluster);
                        expect(response.body.meta.scopeId).toEqual(flowRootState.cluster.uuid);

                        expect(response.body.spec.image).toEqual(caseData.updatedSpec.image);
                        expect(response.body.spec.command).toEqual(caseData.updatedSpec.command);
                        expect(response.body.spec.compute).toEqual(caseData.updatedSpec.compute);
                    });

                    await test.step('patch the template', async function () {
                        const response = await templateUtils.patchWorkloadTemplateV2(
                            templateId,
                            caseData.workloadType,
                            {
                                name: patchedTemplateName
                            },
                            {
                                command: sleep30s
                            },
                            undefined,
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);
                    });

                    await test.step('get the patched template and verify the related fields', async function () {
                        const response = await templateUtils.getWorkloadTemplateV2(
                            templateId,
                            caseData.workloadType,
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);

                        expect(response.body.meta.name).toEqual(patchedTemplateName);
                        expect(response.body.meta.scopeType).toEqual(ScopeType.Cluster);
                        expect(response.body.meta.scopeId).toEqual(flowRootState.cluster.uuid);

                        expect(response.body.spec.image).toEqual(caseData.updatedSpec.image);
                        expect(response.body.spec.command).toEqual(sleep30s);
                        expect(response.body.spec.compute).toEqual(caseData.updatedSpec.compute);
                    });

                    await test.step(`create another ${caseData.workloadType} template`, async function () {
                        const response = await templateUtils.createWorkloadTemplateV2(
                            {
                                workloadType: caseData.workloadType,
                                scopeType: ScopeType.Cluster,
                                scopeId: flowRootState.cluster.uuid,
                                templateName: secondTemplateName,
                                description: description
                            },
                            creationSpec,
                            caseData.workersOnly ? undefined : { masterSpecSameAsWorker: true },
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);
                        secondTemplateId = response.body.meta.id;
                        templateCreated.push(secondTemplateId);
                    });

                    await test.step('list the created templates and verify created templates can be seen', async function () {
                        const response = await templateUtils.listWorkloadTemplateV2(
                            caseData.workloadType,
                            undefined,
                            flowRootState.users.l1_researcher.accessToken
                        );

                        expect(response.response.statusCode).toEqual(200);
                        expect(
                            response.body.entries.some((entry) => entry.meta.id === templateId)
                        ).toBeTruthy();
                        expect(
                            response.body.entries.some((entry) => entry.meta.id === secondTemplateId)
                        ).toBeTruthy();
                    });

                    await test.step('delete the templates', async function () {
                        const idsToDelete = [templateId, secondTemplateId];

                        for (const id of idsToDelete) {
                            const response = await templateUtils.deleteWorkloadTemplateV2(
                                id,
                                caseData.workloadType,
                                flowRootState.users.l1_researcher.accessToken
                            );
                            expect(response.response.statusCode).toEqual(204);
                            templateCreated = templateCreated.filter((item) => item !== id);
                            try {
                                await templateUtils.getWorkloadTemplateV2(id, caseData.workloadType);
                            } catch (error) {
                                expect(error.response.statusCode).toEqual(404);
                            }
                        }
                    });
                });
            });
        });

        test.describe('Test the generic workload template api and filtering/sorting/pagination', function () {
            let templateId1: string;
            let templateId2: string;
            let templateId3: string;

            const trainingTemplateName = objectGenerator.generateShortNameByPrefix('training-templ', 5);
            const distributedTemplateName = objectGenerator.generateShortNameByPrefix('distributed-templ', 5);
            const workspaceTemplateName = objectGenerator.generateShortNameByPrefix('workspace-templ', 5);

            test.beforeAll('create training, distributed and workspace templates', async function () {
                const templates = [
                    {
                        workloadType: WorkloadType.Training,
                        templateName: trainingTemplateName,
                        spec: creationSpec,
                        scopeId: projectIdList[0].toString(),
                        token: flowRootState.users.l1_researcher.accessToken
                    },
                    {
                        workloadType: WorkloadType.Distributed,
                        templateName: distributedTemplateName,
                        spec: distMPISpec,
                        scopeId: projectIdList[0].toString(),
                        token: flowRootState.users.l1_researcher.accessToken
                    },
                    {
                        workloadType: WorkloadType.Workspace,
                        templateName: workspaceTemplateName,
                        spec: ubuntuSpec,
                        scopeId: projectIdList[1].toString(),
                        token: flowRootState.users.system_administrator.accessToken
                    }
                ];

                const commonParams = {
                    scopeType: ScopeType.Project
                };
                const [resp1, resp2, resp3] = await Promise.all(
                    templates.map(({ workloadType, templateName, spec, scopeId, token }) =>
                        templateUtils.createWorkloadTemplateV2(
                            { ...commonParams, workloadType, templateName, scopeId },
                            spec,
                            {
                                masterSpecSameAsWorker: true
                            },
                            token
                        )
                    )
                );

                const templateIds = [];

                [resp1, resp2, resp3].forEach((resp, idx) => {
                    expect(resp.response.statusCode).toEqual(200);
                    templateIds[idx] = resp.body.meta.id;
                });

                [templateId1, templateId2, templateId3] = templateIds;
            });

            test.afterAll('delete training, distributed and workspace templates', async function () {
                const templatesToDelete = [
                    { id: templateId1, workloadType: WorkloadType.Training },
                    { id: templateId2, workloadType: WorkloadType.Distributed },
                    { id: templateId3, workloadType: WorkloadType.Workspace }
                ];
                for (const templInfo of templatesToDelete) {
                    try {
                        const response = await templateUtils.deleteWorkloadTemplateV2(
                            templInfo.id,
                            templInfo.workloadType
                        );
                        expect(response.response.statusCode).toEqual(204);
                    } catch (error) {
                        logger.error(`failed to delete workload template with id: ${templInfo.id}`, error);
                    }
                }
            });

            test('T5193626: Test the generic template API to count, list and get template', async function () {
                await test.step('get total template count and it should be gte 3', async function () {
                    const response = await templateUtils.getTemplateCountV2();

                    expect(response.response.statusCode).toEqual(200);

                    logger.info(`Total template count: ${response.body.count}`);
                    expect(response.body.count).toBeGreaterThanOrEqual(3);
                });

                await test.step('list all types of templates', async function () {
                    const response = await templateUtils.listAllWorkloadTemplatesV2();

                    expect(response.response.statusCode).toEqual(200);

                    const templateNames = response.body.entries.map((entry) => entry.meta.name);
                    logger.info(`All templates listed: ${templateNames}`);
                    expect(templateNames).toContain(trainingTemplateName);
                    expect(templateNames).toContain(distributedTemplateName);
                    expect(templateNames).toContain(workspaceTemplateName);
                });

                await test.step('get different types of templates using the generic template api', async function () {
                    const [resp1, resp2, resp3] = await Promise.all(
                        [templateId1, templateId2, templateId3].map((templId) =>
                            templateUtils.genericGetWorkloadTemplateV2(templId)
                        )
                    );

                    const responses = [
                        { resp: resp1, name: trainingTemplateName },
                        { resp: resp2, name: distributedTemplateName },
                        { resp: resp3, name: workspaceTemplateName }
                    ];

                    responses.forEach(({ resp, name }) => {
                        expect(resp.response.statusCode).toEqual(200);
                        expect(resp.body.meta.name).toEqual(name);
                    });
                });
            });

            test('T5193627: test workload template filtering/sorting/pagination', async function () {
                await test.step('test filtering: template field contains function', async function () {
                    const response = await templateUtils.listAllWorkloadTemplatesV2(
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        [`name=@${workspaceTemplateName}`]
                    );

                    expect(response.response.statusCode).toEqual(200);
                    expect(response.body.entries.length).toEqual(1);
                    expect(response.body.entries[0].meta.name).toEqual(workspaceTemplateName);
                });

                await test.step('test filtering: workload type field equals to training', async function () {
                    const response = await templateUtils.listAllWorkloadTemplatesV2(
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        [
                            `workloadType!=${WorkloadType.Distributed},workloadType!=${WorkloadType.Workspace},workloadType!=${WorkloadType.Inference}`
                        ]
                    );

                    expect(response.response.statusCode).toEqual(200);
                    response.body.entries.forEach((entry) =>
                        expect(entry.meta.workloadType).toEqual(WorkloadType.Training)
                    );
                });

                await test.step('test filtering: create by field equals to system admin', async function () {
                    const response = await templateUtils.listAllWorkloadTemplatesV2(
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        [`createdBy==${flowRootState.users.system_administrator.email}`]
                    );

                    expect(response.response.statusCode).toEqual(200);
                    response.body.entries.forEach((entry) =>
                        expect(entry.meta.createdBy).toEqual(
                            flowRootState.users.system_administrator.email
                        )
                    );
                });

                // the scope of sorting test data is decided by the flexible template UI column fields
                const sortingData = [
                    { sortBy: 'name', sortOrder: 'asc' },
                    { sortBy: 'name', sortOrder: 'desc' },
                    { sortBy: 'scopeType', sortOrder: 'asc' },
                    { sortBy: 'scopeType', sortOrder: 'desc' },
                    // { sortBy: 'image', sortOrder: 'asc' },
                    // { sortBy: 'image', sortOrder: 'desc' },
                    // { sortBy: 'gpuComputeRequest', sortOrder: 'asc' },
                    // { sortBy: 'gpuComputeRequest', sortOrder: 'desc' },
                    { sortBy: 'workloadType', sortOrder: 'asc' },
                    { sortBy: 'workloadType', sortOrder: 'desc' },
                    // { sortBy: 'workloadArchitecture', sortOrder: 'asc' },
                    // { sortBy: 'workloadArchitecture', sortOrder: 'desc' },
                    { sortBy: 'createdBy', sortOrder: 'asc' },
                    { sortBy: 'createdBy', sortOrder: 'desc' },
                    { sortBy: 'createdAt', sortOrder: 'asc' },
                    { sortBy: 'createdAt', sortOrder: 'desc' },
                    { sortBy: 'updatedAt', sortOrder: 'asc' },
                    { sortBy: 'updatedAt', sortOrder: 'desc' }
                ];

                for (const data of sortingData) {
                    await test.step(`test sorting: sort by ${data.sortBy} with ${data.sortOrder} order`, async function () {
                        const response = await templateUtils.listAllWorkloadTemplatesV2(
                            undefined,
                            undefined,
                            data.sortOrder as SortOrder,
                            data.sortBy as SortBy,
                            [`clusterId==${flowRootState.cluster.uuid}`]
                        );

                        expect(response.response.statusCode).toEqual(200);
                        const respData = response.body.entries;

                        expect(
                            isSorted(respData, data.sortOrder as SortOrder, data.sortBy as SortBy)
                        ).toBeTruthy();
                    });
                }

                await test.step('test workload template pagination', async function () {
                    const response1 = await templateUtils.listAllWorkloadTemplatesV2(0, 2);

                    expect(response1.response.statusCode).toEqual(200);
                    expect(response1.body.entries.length).toEqual(2);

                    const response2 = await templateUtils.listAllWorkloadTemplatesV2(2, 5);

                    expect(response1.response.statusCode).toEqual(200);
                    expect(response1.body.entries.length).toBeGreaterThanOrEqual(1);
                    expect(response1.body.entries.length).toBeLessThanOrEqual(5);
                });
            });
        });
    });
});
