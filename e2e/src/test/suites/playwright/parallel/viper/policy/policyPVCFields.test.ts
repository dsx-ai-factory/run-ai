/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { policyApi as policyApiHandler } from '../../../../../../modules/api';
import { runaiCliV2 } from '../../../../../../modules/cli';
import {
    config,
    crds,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    policyUtils,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    isOCP,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect } from 'chai';
import { WorkloadType } from '../../../../../../modules/helpers/utils/policyUtils';
import { WorkloadType as WorkloadTypev } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import {
    TrainingPolicyDefaultsAndRulesV2Rules,
    WorkspacePolicyDefaultsAndRulesV2Rules
} from '../../../../../../modules/api/apis/policyApi';
import { test, TestInfo } from '@playwright/test';
import {
    cliV2WorkloadType,
    DeleteFlagsMerged,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

const WorkloadTypeParams = [
    {
        policyType: 'training',
        image: config.quickStartImage.training,
        jobKind: crds.runaiTrainingWorkloadCrd.kind,
        crd: crds.runaiTrainingWorkloadCrd,
        kind: WorkloadType.Training,
        isInteractive: false
    },
    {
        policyType: 'interactive',
        image: config.quickStartImage.interactive,
        jobKind: crds.runaiInteractiveWorkloadCrd.kind,
        crd: crds.runaiInteractiveWorkloadCrd,
        kind: WorkloadType.Interactive,
        isInteractive: true
    }
];

WorkloadTypeParams.forEach(function (params) {
    test.describe(`${params.policyType} policy - sanity checks of pvc rules, tags: ${tags.resource.policy} ${tags.tool.api} ${tags.teams.viper} ${tags.runOn.parallelCI} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`, function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let jobName: string;
        let policyName: string;
        let workloadId: string;
        let anyWorkloadSubmissionEnabled: boolean;

        async function _createPolicy(defaults: object, rules?: object): Promise<void> {
            await policyUtils.createPolicy(
                params.isInteractive,
                policyName,
                namespace,
                project.id,
                params.kind,
                defaults,
                rules
            );
        }

        async function _updatePolicy(defaults: object, rules?: object): Promise<void> {
            await policyUtils.updateProjectScopedPolicy(
                params.isInteractive,
                policyName,
                namespace,
                project.id,
                params.kind,
                defaults,
                rules
            );
        }

        async function _submitWorkloadWithPVC(claimName: string): Promise<string> {
            // Submit via CLI with new PVC (creates PVC automatically)
            const wlType = params.isInteractive ? cliV2WorkloadType.workspace : cliV2WorkloadType.training;
            const submitFlags = {
                [SubmitFlagsMerged.image]: params.image,
                [SubmitFlagsMerged.project]: projectName,
                [SubmitFlagsMerged.newPvc]: `claimname=${claimName},size=1Gi,path=/path,accessmode-rwo`
            } as Partial<typeof SubmitFlagsMerged>;
            runaiCliV2.submit(submitFlags, jobName, wlType);

            // Wait for workload to be processed
            await waitUtils.sleep(2000);

            return jobName;
        }

        async function _submitWorkloadWithHostPath(hostPath: string, mountPath: string): Promise<string> {
            // Submit via CLI with host path volume
            const wlType = params.isInteractive ? cliV2WorkloadType.workspace : cliV2WorkloadType.training;
            const submitFlags = {
                [SubmitFlagsMerged.image]: params.image,
                [SubmitFlagsMerged.project]: projectName,
                [SubmitFlagsMerged.hostpath]: `path=${hostPath},mount=${mountPath}`
            } as Partial<typeof SubmitFlagsMerged>;
            runaiCliV2.submit(submitFlags, jobName, wlType);

            // Wait for workload to be processed
            await waitUtils.sleep(2000);

            return jobName;
        }

        async function _waitForWorkloadRunning(): Promise<void> {
            if (anyWorkloadSubmissionEnabled) {
                await waitUtils.waitForBatchJobToExist(jobName, namespace);
                await waitUtils.waitForPods(jobName, namespace, 1, 120, ['Running']);
            } else {
                await waitUtils.waitForJobToBeRunning(jobName, namespace, params.jobKind);
            }
        }

        test.beforeAll(async function () {
            policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
            anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                params.isInteractive
                    ? AnyWorkloadSubmissionType.Workspace
                    : AnyWorkloadSubmissionType.Training
            );

            // Enable control plane policies so they are applied by webhook
            await policyUtils.enableClusterPolicySync();
        });

        test.beforeEach(async function () {
            jobName = objectGenerator.generateShortNameByPrefix('job-for-policy');
            projectName = objectGenerator.generateShortNameByPrefix('project-for-policy');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            policyName = `project-${project.id}`;
        });

        test.afterEach(async function ({}, testInfo) {
            // Delete workload if it was created
            if (workloadId) {
                try {
                    const wlType =
                        params.jobKind === 'InteractiveWorkload'
                            ? WorkloadTypev.WORKSPACE
                            : WorkloadTypev.TRAINING;
                    await workloadsUtilsV2.deleteWorkload(workloadId, wlType);
                } catch {
                    logger.error(`failed to delete workload ${workloadId}`);
                }
                workloadId = undefined;
            }

            // Delete policy
            if (testInfo.status !== 'skipped') {
                try {
                    if (params.isInteractive) {
                        await policyApiHandler.policyApi.deleteWorkspacePolicy(
                            'project',
                            undefined,
                            project.id.toString()
                        );
                    } else {
                        await policyApiHandler.policyApi.deleteTrainingPolicy(
                            'project',
                            undefined,
                            project.id.toString()
                        );
                    }
                } catch {
                    logger.error(`failed to delete policy for project ${project.id}`);
                }
            }

            // Delete project
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });

        test.describe('pvc rules', function () {
            let rules: WorkspacePolicyDefaultsAndRulesV2Rules | TrainingPolicyDefaultsAndRulesV2Rules;
            const claimName = 'regularclaimname';

            test.beforeEach(function ({}, testInfo: TestInfo) {
                rules = {
                    storage: {
                        pvc: {
                            instances: {
                                canAdd: true
                            }
                        }
                    }
                };
            });
            test('can submit a job with pvc, when canAdd=true', async function () {
                await _createPolicy(undefined, rules);
                workloadId = await _submitWorkloadWithPVC(claimName);

                // Wait for workload to be running
                await _waitForWorkloadRunning();
            });

            test('cannot submit a job with pvc, when canAdd=false', async function () {
                rules.storage.pvc.instances.canAdd = false;
                await _createPolicy(undefined, rules);

                try {
                    await _submitWorkloadWithPVC(claimName);
                } catch (error) {
                    expect(error.message).to.contain(
                        'The workload request does not comply with the policy set'
                    );
                    return;
                }
                throw 'submit should fail';
            });

            test('can submit a job with host path, when hostPath.canAdd=true, and pvc.canAdd.=false', async function () {
                // Skip on OpenShift: HostPath volumes at root level (e.g., /sgw-2023-03-08-f99b5) require
                // creating directories on the node, which OpenShift's Security Context Constraints (SCC)
                // block by default. This is expected OpenShift security behavior.
                if (isOCP()) {
                    return;
                }

                await _createPolicy(undefined, {
                    storage: {
                        pvc: {
                            instances: {
                                canAdd: false
                            }
                        },
                        hostPath: {
                            instances: {
                                canAdd: true
                            }
                        }
                    }
                });
                workloadId = await _submitWorkloadWithHostPath('/sgw-2023-03-08-f99b5', '/volume1path');

                // Wait for workload to be running
                await _waitForWorkloadRunning();
            });

            test('can update pvc rule', async function () {
                await _createPolicy(undefined, rules);

                const newValue = false;
                rules.storage.pvc.instances.canAdd = newValue;

                await _updatePolicy(undefined, rules);

                await waitUtils.sleep(2000);

                try {
                    await _submitWorkloadWithPVC(claimName);
                } catch (error) {
                    expect(error.message).to.contain(
                        'The workload request does not comply with the policy set'
                    );
                    return;
                }
                throw 'submit should fail';
            });
        });
    });
});
