/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    tags,
    operations,
    projectResourcesUtils,
    workloadsCrudUtils,
    waitUtils,
    crds,
    computeResourceUtils,
    environmentUtils,
    assetsApiV2Utils,
    objectGenerator,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadStatus } from '../../../../../../modules/helpers/models/baseWorkload';
import { StatusCodes } from 'http-status-codes';
import {
    EnvironmentAsset,
    EnvironmentCreationRequest,
    Scope
} from '../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';
import { kubeCli } from '../../../../../../modules/cli';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

let project: Project;
let workloadId: string;
let workloadName: string;
let environmentName: string;
let environmentId: string;

const ENV_VAR_FOR_ENVIRONMENT = { name: 'env1', value: 'val1' };
const ENV_VAR_SPECIFIC_RUN_PARAMS = { name: 'env2', value: 'val2' };

const WorkloadTypeParams = [
    {
        workloadType: 'training',
        workloadName: 'test-training',
        createFunction: assetsApiV2Utils.createTraining,
        deleteFunction: workloadsCrudUtils.deleteTraining,
        jobKind: crds.runaiTrainingWorkloadCrd
    },
    {
        workloadType: 'interactive',
        workloadName: 'test-interactive',
        createFunction: assetsApiV2Utils.createWorkspace,
        deleteFunction: workloadsCrudUtils.deleteWorkspace,
        jobKind: crds.runaiInteractiveWorkloadCrd
    }
];

WorkloadTypeParams.forEach(function (params) {
    test.describe(`Submit ${params.workloadType} V2 - specificRunParams, tags: ${tags.operations.submit} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit}`, function () {
        test.beforeAll('creating project', async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            environmentName = objectGenerator.generateNameByPrefix('environment');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll('Cleanup', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                logger.error(`Failed to delete project ${project.name} and namespace ${project.namespace}`);
            }
        });

        async function getWorkloadCreationPayload(environmentId: string): Promise<any> {
            return {
                meta: {
                    name: params.workloadName,
                    useGivenNameAsPrefix: true,
                    projectId: project.id.toString(),
                    clusterId: flowRootState.cluster.uuid
                },
                assets: {
                    environment: environmentId,
                    command: 'sleep infinity',
                    compute: await computeResourceUtils.getId('cpu-only')
                },
                specificRunParams: {
                    environmentVariables: [ENV_VAR_SPECIFIC_RUN_PARAMS]
                }
            };
        }

        async function createEnvironment(environmentName: string): Promise<EnvironmentAsset> {
            const environmentCreationRequest: EnvironmentCreationRequest = {
                meta: {
                    name: environmentName,
                    scope: Scope.Project,
                    projectId: project.id,
                    clusterId: project.clusterUuid,
                    workloadSupportedTypes: {
                        workspace: true,
                        training: true
                    }
                },
                spec: {
                    image: 'alpine',
                    command: 'sleep infinity',
                    environmentVariables: [ENV_VAR_FOR_ENVIRONMENT]
                }
            };
            return await environmentUtils.create(environmentCreationRequest);
        }

        test(`Create ${params.workloadType} workload and validate it can be managed using the APIs`, async function () {
            await test.step(`${params.workloadType} Create environment`, async function () {
                environmentId = (await createEnvironment(environmentName)).meta.id;
            });

            await test.step(`Create ${params.workloadType} workload`, async function () {
                const result = await params.createFunction(await getWorkloadCreationPayload(environmentId));
                workloadId = result.body.meta.workloadId;
                workloadName = result.body.meta.name;

                expect(result.response.statusCode).to.equal(StatusCodes.OK);
                expect(result.body.meta.desiredPhase).to.equal(WorkloadStatus.RUNNING);
                await waitUtils.waitForPods(workloadName, project.namespace, 1, 120, ['Running']);
            });

            await test.step(`Validate  ${params.workloadType} workload specificRunParams env var override the environment env var`, async function () {
                if (
                    await isAnyWorkloadSubmissionEnabled(
                        params.workloadType === 'interactive'
                            ? AnyWorkloadSubmissionType.Workspace
                            : AnyWorkloadSubmissionType.Training
                    )
                ) {
                    const job = await kubeCli.getBatchJob(workloadName, project.namespace);
                    const envVars = job.spec?.template.spec?.containers[0].env ?? [];
                    expect(
                        envVars.find((entry) => entry.name === ENV_VAR_SPECIFIC_RUN_PARAMS.name)?.value
                    ).to.equal(ENV_VAR_SPECIFIC_RUN_PARAMS.value);
                    expect(envVars.find((entry) => entry.name === ENV_VAR_FOR_ENVIRONMENT.name)).to.be
                        .undefined;
                } else {
                    const envVars = await workloadUtils.getEnvVars(
                        workloadName,
                        project.namespace,
                        params.jobKind
                    );
                    expect(envVars[ENV_VAR_SPECIFIC_RUN_PARAMS.name].value).to.equal(
                        ENV_VAR_SPECIFIC_RUN_PARAMS.value
                    );
                    expect(envVars?.[ENV_VAR_FOR_ENVIRONMENT.name]).to.undefined;
                }
            });

            await test.step(`Delete ${params.workloadType} workload`, async function () {
                const deleteResult = await params.deleteFunction(workloadId);
                expect(deleteResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);

                await waitUtils.waitForWorkloadToBeDeleted(workloadName);
            });

            await test.step(`${params.workloadType} | Delete environment`, async function () {
                await environmentUtils.waitBeforeAbleToRemoveEnv(environmentName);
                await environmentUtils.delete(environmentId);
            });
        });
    });
});
