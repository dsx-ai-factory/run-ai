/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    flowRootState,
    environmentUtils,
    computeResourceUtils,
    credentialUtils,
    dataSourcesUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { cliV2WorkloadType } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { runaiCliV2 } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    EnvironmentAsset,
    WorkloadSupportedTypes,
    ComputeAsset
} from '../../../../../../modules/api/apis/assetsApi';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';

test.describe(`system admin can update project quota  ${tags.tool.ui} ${tags.resource.workloads} ${tags.runOn.dgxc} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-test');
            namespace = Utils.namespaceFromProjectName(projectName);
        });

        test.afterAll('deleting project and namespace', async ({ }, testInfo) => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });
        test('T3738323: Update project quota', async function () {
            let workloadName;
            let training: any;
            let trainingWorkloadId: string;

            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;

            await test.step('Create a project with disabled over-quota and GPU quota set to 2', async () => {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
            });

            await test.step('Create a GPU training job with 4 GPUs (exceeding quota)', async () => {
                const result = await workloadsCrudUtils.createTraining(
                    'quota-test-1',
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity',
                        compute: {
                            gpuDevicesRequest: 4
                        }
                    }
                );
                trainingWorkloadId = result.body.workloadId;
                workloadName = result.body.name;
            });

            await test.step('Verify the job is in pending state due to quota limits exceeded', async () => {
                await workloadsCrudUtils.waitForWorkloadPhase(
                    trainingWorkloadId,
                    Phase.Pending,
                    WorkloadType.TRAINING
                );

                runaiCliV2.loginAsUser(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
                runaiCliV2.projectSet(project.name);
                await waitUtils.waitForCondition(
                    async () => {
                        const workloadDescription = await runaiCliV2.describeAsync(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        const workloadString = String(workloadDescription);
                        return workloadString.includes(
                            'quota has reached the allowable limit of GPUs. Limit is 2 GPUs'
                        );
                    },
                    'Wait for quota error message to appear in workload description',
                    true,
                    30
                );
            });

            await test.step('Update the project with larger GPU quota (4)', async () => {
                if (project.nodePoolsResources && project.nodePoolsResources.length > 0) {
                    project.nodePoolsResources[0].gpu.deserved = 4;
                    project.nodePoolsResources[0].gpu.maxAllowed = 4;
                    project.nodePoolsResources[0].gpu.overQuotaWeight = 4;
                }
                await projectResourcesUtils.updateProject(project);
            });

            await test.step('Verify the job is now in running state', async () => {
                await workloadsCrudUtils.waitForWorkloadPhase(
                    trainingWorkloadId,
                    Phase.Running,
                    WorkloadType.TRAINING
                );
            });

            await test.step('Revert the GPU quota limit back to 2', async () => {
                if (project.nodePoolsResources && project.nodePoolsResources.length > 0) {
                    project.nodePoolsResources[0].gpu.deserved = 2;
                    project.nodePoolsResources[0].gpu.maxAllowed = 2;
                    project.nodePoolsResources[0].gpu.overQuotaWeight = 2;
                }
                await projectResourcesUtils.updateProject(project);
            });

            await test.step('Verify the job is still in running state', async () => {
                await workloadsCrudUtils.waitForWorkloadPhase(
                    trainingWorkloadId,
                    Phase.Running,
                    WorkloadType.TRAINING
                );
            });

            await test.step('Delete the job', async () => {
                await workloadsCrudUtils.deleteTraining(trainingWorkloadId);
            });

            await test.step('Enable the over-quota feature in the project', async () => {
                if (project.nodePoolsResources && project.nodePoolsResources.length > 0) {
                    project.nodePoolsResources[0].gpu.maxAllowed = -1;
                }
                await projectResourcesUtils.updateProject(project);
            });

            await test.step('Create a GPU job with 4 GPUs (exceeding quota) under over-quota enabled project', async () => {
                const result = await workloadsCrudUtils.createTraining(
                    'quota-test-2',
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep 300',
                        compute: {
                            gpuDevicesRequest: 4
                        }
                    }
                );
                training = result.body;
            });

            await test.step('Verify the job is in running state despite exceeding quota', async () => {
                await workloadsCrudUtils.waitForWorkloadPhase(
                    training.workloadId,
                    Phase.Running,
                    WorkloadType.TRAINING
                );
            });

            await test.step('Clean up - delete the job', async () => {
                await workloadsCrudUtils.deleteTraining(training.workloadId);
            });
        });
    });
});

test.describe(`delete project with running workloads  ${tags.tool.ui} ${tags.resource.workloads} ${tags.runOn.dgxc} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_18)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;
        let auth: AuthPage;
        let browserEngine: BrowserEngine;
        let projectDeleted = false;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-test');
            namespace = Utils.namespaceFromProjectName(projectName);
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
        });

        test.afterAll('deleting project and namespace', async ({ }, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            if (!projectDeleted) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                    throw error;
                }
            }
        });

        test('T4142409: Project deletion with running workloads', async function () {
            test.setTimeout(300 * 1000);
            videoName = `T4142409-${projectName}`;
            let environmentAsset: EnvironmentAsset;
            let computeResource: ComputeAsset;

            const pvcName = objectGenerator.generateNameByPrefix('pvc');
            const credentialName = objectGenerator.generateNameByPrefix('credential');
            const workspaceJobName = objectGenerator.generateNameByPrefix('workspace-job');
            const mpiJobName = objectGenerator.generateNameByPrefix('mpi-job');

            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;

            await test.step('Create a new project with 8 GPUs', async () => {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 8,
                    maxAllowedGpus: 8
                });
            });

            await test.step('Create a distributed MPI Environment', async () => {
                environmentAsset = await environmentUtils.createEnvironmentWithArbitraryFields(
                    project.id,
                    '/bin/bash -c',
                    WorkloadSupportedTypes.DistFrameworkEnum.Mpi,
                    config.quickStartImage.distributedMpi,
                    '"sleep 300"'
                );
            });

            await test.step('Create a compute resource', async () => {
                const computeCreationRequest = {
                    meta: {
                        name: objectGenerator.generateNameByPrefix('compute'),
                        scope: Scope.Project,
                        projectId: project.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 0.01,
                        gpuDevicesRequest: 8,
                        cpuMemoryRequest: '10M'
                    }
                };
                computeResource = await computeResourceUtils.create(computeCreationRequest);
            });

            await test.step('Create a data source (PVC)', async () => {
                await dataSourcesUtils.createPVC(pvcName, Scope.Project, project);
                await dataSourcesUtils.validateExpectedStatus(pvcName, 'No issues found');
            });

            await test.step('Create credential', async () => {
                const passwordCreationRequest = {
                    meta: {
                        name: credentialName,
                        scope: Scope.Project,
                        projectId: project.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        user: 'username',
                        password: 'password'
                    }
                };
                await credentialUtils.createPasswordAsset(passwordCreationRequest);
            });

            await test.step('Create a standard workspace job', async () => {
                const { body: workspace } = await workloadsCrudUtils.createWorkspace(
                    workspaceJobName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity'
                    }
                );
                await workloadsCrudUtils.waitForWorkloadPhase(
                    workspace.workloadId,
                    Phase.Running,
                    WorkloadType.WORKSPACE
                );
            });

            await test.step('Create a distributed MPI job', async () => {
                const { body: training } = await workloadsCrudUtils.createDistributed(
                    mpiJobName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity',
                        numWorkers: 1,
                        distributedFramework: DistributedFramework.Mpi
                    },
                    true
                );
                await workloadsCrudUtils.waitForWorkloadPhase(
                    training.workloadId,
                    Phase.Running,
                    WorkloadType.DISTRIBUTED,
                    undefined,
                    180
                );
            });

            await test.step('Delete project via UI', async () => {
                await auth.signIn(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
                const projectPage = new ProjectPage(browserEngine.currentPage);
                await projectPage.goToPage();

                await projectPage.openDeleteProjectDialog(projectName);
                const deleteDialogText = await projectPage.getDeleteProjectDialogText();
                expect(deleteDialogText).toContain(`Project ${projectName} has`);
                expect(deleteDialogText).toContain('associated workloads');
                expect(deleteDialogText).toContain(
                    'Permanently delete this project and its associated workloads and assets?'
                );

                await projectPage.confirmDeleteProject();
                await projectPage.waitForProjectDeletion(projectName);
            });

            await test.step('Verify resources are deleted', async () => {
                await environmentUtils.waitForEnvironmentToBeDeleted(environmentAsset.meta.name);
                await computeResourceUtils.waitForComputeResourceToBeDeleted(computeResource.meta.name);
                await dataSourcesUtils.waitForDataSourceToBeDeleted(pvcName);
                await credentialUtils.waitForCredentialsToBeDeleted(credentialName);
                await waitUtils.waitForWorkloadToBeDeleted(workspaceJobName);
                await waitUtils.waitForWorkloadToBeDeleted(mpiJobName);
                projectDeleted = true;
            });
        });
    });
});
