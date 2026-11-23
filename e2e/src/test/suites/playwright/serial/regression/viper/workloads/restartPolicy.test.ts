/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assert } from 'chai';
import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    getWorkloadTypeToCrd,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { bashCli, kubeCli, runaiCliV2 } from '../../../../../../../modules/cli';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { workloadCrdParamsMapper } from './restartPolicy.consts';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

test.describe(`Workload restart policy   ${tags.teams.viper}`, { tag: [tags.resource.trainingWorkload] }, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        const runningCommand = '-- sleep 99999';
        const masterWorkloadType = [
            cliV2WorkloadType.mpi,
            cliV2WorkloadType.pyTorch,
            cliV2WorkloadType.tensorFlow,
            cliV2WorkloadType.xgBoost
        ];
        const noMasterWorkloadTypes = [
            cliV2WorkloadType.workspace,
            cliV2WorkloadType.training,
            cliV2WorkloadType.jax
        ];
        const allTypes = [...masterWorkloadType, ...noMasterWorkloadTypes];
        const runSuffix = objectGenerator.generateShortNameByPrefix('run-rp', 3);
        const redirectedRestartPolicyTypes = new Set<cliV2WorkloadType>();

        function getSubmissionType(
            workloadType: cliV2WorkloadType
        ): AnyWorkloadSubmissionType | undefined {
            if (workloadType === cliV2WorkloadType.workspace) {
                return AnyWorkloadSubmissionType.Workspace;
            }
            if (workloadType === cliV2WorkloadType.training) {
                return AnyWorkloadSubmissionType.Training;
            }
            return undefined;
        }

        async function deleteProject(projectObj: Project): Promise<void> {
            try {
                await operations.deleteProjectAndNamespace(projectObj.id);
            } catch (error) {
                logger.error(`Failed to delete project ${projectObj.name}`);
            }
        }

        function setFlags(
            restartPolicy = '',
            masterRestartPolicy = '',
            backoffLimit = ''
        ): Partial<typeof SubmitFlagsMerged> {
            const submitFlag = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.command]: runningCommand
            } as Partial<typeof SubmitFlagsMerged>;
            if (backoffLimit) {
                submitFlag[SubmitFlagsMerged.backoffLimit] = backoffLimit;
            }
            if (restartPolicy) {
                submitFlag[SubmitFlagsMerged.restartPolicy] = restartPolicy;
            }
            if (masterRestartPolicy) {
                submitFlag[SubmitFlagsMerged.masterRestartPolicy] = masterRestartPolicy;
            }
            return submitFlag;
        }

        function validateWorkloadResource(
            workloadType: cliV2WorkloadType,
            jobResource: any,
            expectedWorkerValue: string,
            expectedLeaderValue: string
        ): void {
            const crdParam = workloadCrdParamsMapper[workloadType];
            if (crdParam.workerName) {
                const workerValue =
                    jobResource.spec[crdParam.specName][crdParam.workerName]['template']['spec'][
                        'restartPolicy'
                    ];
                expect(workerValue).toBe(expectedWorkerValue);
                if (crdParam.leaderName) {
                    const leaderValue =
                        jobResource.spec[crdParam.specName][crdParam.leaderName]['template']['spec'][
                            'restartPolicy'
                        ];
                    expect(leaderValue).toBe(expectedLeaderValue);
                }
            } else {
                const rpValue =
                    jobResource.spec.template?.spec?.restartPolicy ?? jobResource.spec.restartPolicy;
                const normalizedRpValue = rpValue?.value ?? rpValue;
                expect(normalizedRpValue, 'Restart policy value is invalid').toBe(expectedWorkerValue);
            }
        }

        async function getJobResource(workloadType: cliV2WorkloadType, jobName: string): Promise<any> {
            const submissionType = getSubmissionType(workloadType);
            if (
                submissionType !== undefined &&
                (await isAnyWorkloadSubmissionEnabled(submissionType))
            ) {
                return kubeCli.getBatchJob(jobName, project.namespace);
            }
            return JSON.parse(
                bashCli.exec(
                    `kubectl get ${getWorkloadTypeToCrd()[workloadType].plural} -n ${project.namespace} ${jobName} -ojson`
                )
            );
        }

        async function generateWorkloads(
            workloadTypes: cliV2WorkloadType[],
            jobNameSuffix: string,
            expectedPhase: Phase = undefined,
            restartPolicy = '',
            masterRestartPolicy = '',
            backoffLimit = ''
        ): Promise<void> {
            const submitFlag = setFlags(restartPolicy, masterRestartPolicy, backoffLimit);

            Object.values(workloadTypes).forEach((type) => {
                const jobName = `${type}-${jobNameSuffix}`;
                runaiCliV2.submit(submitFlag, jobName, type as cliV2WorkloadType);
            });
            for (const type of Object.values(workloadTypes)) {
                const jobName = `${type}-${jobNameSuffix}`;
                await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, expectedPhase);
            }
        }

        test.beforeAll(async function () {
            cliLoginAsSystemAdmin();
            project = await projectResourcesUtils.createProjectAndNamespace('wl-issues');
            runaiCliV2.projectSet(project.name);
            for (const type of [cliV2WorkloadType.workspace, cliV2WorkloadType.training]) {
                const submissionType = getSubmissionType(type);
                if (submissionType && (await isAnyWorkloadSubmissionEnabled(submissionType))) {
                    redirectedRestartPolicyTypes.add(type);
                }
            }
        });

        test.afterAll(async function () {
            await deleteProject(project);
        });

        test.describe('validate restartPolicy workload resource', function () {
            const rpValue = 'Always';
            test.beforeAll(async function () {
                await generateWorkloads(
                    allTypes,
                    runSuffix,
                    Phase.Running,
                    rpValue
                );
            });
            Object.values(allTypes).forEach((type) => {
                test(`validate ${type} restart policy in workload resource`, async function () {
                    const jobName = `${type}-${runSuffix}`;
                    const jobResource = await getJobResource(type, jobName);
                    // Native CRDs retain Always. Redirected IW/TW workloads become batch Jobs,
                    // where the composer translates the deprecated value to valid OnFailure.
                    const expectedRestartPolicy = redirectedRestartPolicyTypes.has(type)
                        ? 'OnFailure'
                        : rpValue;
                    validateWorkloadResource(
                        type,
                        jobResource,
                        expectedRestartPolicy,
                        expectedRestartPolicy
                    );
                });
            });
        });
        test.describe('validate master restartPolicy ', function () {
            test('should fail to submit, restartPolicy=never backofflimit>0 ', function () {
                Object.values(masterWorkloadType).forEach((type) => {
                    const jobName = 'should-fail-to-submit';
                    const submitFlag = setFlags('Never', undefined, '1');
                    assert.throw(
                        () => runaiCliV2.submit(submitFlag, jobName, type as cliV2WorkloadType),
                        'cannot set backoff limit when restart policy is Never'
                    );
                });
            });
            test('should succeed to submit, restartPolicy=never,masterRestart=Always backofflimit>0 ', async function () {
                const masterNeverSuffix = objectGenerator.generateShortNameByPrefix('rp-first', 3);
                const masterAlwaysSuffix = objectGenerator.generateShortNameByPrefix('rp-second', 3);
                await Promise.all([
                    generateWorkloads(
                        masterWorkloadType,
                        `${masterNeverSuffix}`,
                        Phase.Running,
                        'Always',
                        'Never',
                        '1'
                    ),
                    generateWorkloads(
                        masterWorkloadType,
                        `${masterAlwaysSuffix}`,
                        Phase.Running,
                        'Never',
                        'Always',
                        '1'
                    )
                ]);
                let leaderValue = 'Never';
                let workerValue = 'Always';
                for (const suffix of [masterNeverSuffix, masterAlwaysSuffix]) {
                    for (const type of Object.values(masterWorkloadType)) {
                        const jobName = `${type}-${suffix}`;
                        const jobResource = await getJobResource(type, jobName);
                        validateWorkloadResource(type, jobResource, workerValue, leaderValue);
                    }
                    leaderValue = 'Always';
                    workerValue = 'Never';
                }
            });
        });
    });
});
