/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    config,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    versionUtils,
    waitUtils,
    workloadsUtilsV2,
    projectUtils,
    crds
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { backendClusterApi, projectsDepartmentsApi } from '../../../../../../../modules/api';
import { SupportedVersions } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { kubeCli } from '../../../../../../../modules/cli';

test.describe(
    `Jobs Sync tests, tags: ${tags.resource.jobs} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.services.agent, tags.services.clusterSync, tags.runOn.serialCI] },
    () => {
        let projects = new Array<Project>(0);
        const numberOfProjects = 4;
        const numberOfJobsInEachProject = 6;

        test.beforeAll(async () => {
            projectsDepartmentsApi.projectsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            backendClusterApi.podGroupsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            backendClusterApi.podsApi.accessToken = flowRootState.users.system_administrator.accessToken;

            let promisesToAwait = [];
            for (let index = 0; index < numberOfProjects; index++) {
                const projectName = objectGenerator.generateShortNameByPrefix('p');
                promisesToAwait.push(
                    projectResourcesUtils.createProject(Utils.namespaceFromProjectName(projectName), {
                        name: projectName,
                        maxAllowedGpus: numberOfJobsInEachProject,
                        deservedGpus: numberOfJobsInEachProject
                    })
                );
            }
            projects = await Promise.all(promisesToAwait);

            promisesToAwait = [];
            for (const project of projects) {
                promisesToAwait.push(projectUtils.waitForProjectCrdNonEmptyStatus(project.name, 'Ready'));
            }
            await Promise.all(promisesToAwait);
        });

        test.afterAll(async () => {
            if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.11')) {
                await operations.updateClusterSyncConfig(null, null, true);
            }
            await operations.toggleRunaiOperand(operations.RunaiClusterSyncName, true);

            logger.debug('Deleting all projects from the control plane');
            await operations.deleteAllProjects();
        });

        test.describe('Jobs syncing - from cluster to backend - after cluster-sync restart', () => {
            const jobsByProject = new Map<Project, string[]>();
            let preSyncWorkloadName: string;
            let preSyncWorkloadBackendId: string;

            async function submitClusterBatchJob(project: Project): Promise<string> {
                const workloadName = objectGenerator.generateShortNameByPrefix('j');
                const namespace = Utils.namespaceFromProjectName(project.name);

                await kubeCli.BatchV1Api.createNamespacedJob({
                    namespace,
                    body: {
                        apiVersion: 'batch/v1',
                        kind: 'Job',
                        metadata: { name: workloadName, namespace },
                        spec: {
                            backoffLimit: 0,
                            template: {
                                metadata: { labels: { app: workloadName } },
                                spec: {
                                    schedulerName: operations.RunaiSchedulerName,
                                    restartPolicy: 'Never',
                                    containers: [
                                        {
                                            name: 'job',
                                            image: config.quickStartImage.ubuntu,
                                            command: ['/bin/sh', '-c', 'sleep infinity']
                                        }
                                    ]
                                }
                            }
                        }
                    }
                });
                await waitUtils.waitForBatchJobToBeRunning(workloadName, namespace, 120);
                return workloadName;
            }

            async function deleteClusterBatchJob(workloadName: string, project: Project): Promise<void> {
                const namespace = Utils.namespaceFromProjectName(project.name);
                await kubeCli.BatchV1Api.deleteNamespacedJob({
                    name: workloadName,
                    namespace,
                    // batch/v1 Jobs default to orphaning dependents when the policy is omitted.
                    // Match `kubectl delete` so the Job's Pod and ExternalWorkload are garbage-collected.
                    propagationPolicy: 'Background'
                });
                await waitUtils.waitForBatchJobToExist(workloadName, namespace, false, 60);
                await kubeCli.waitForCrdToBeDeleted(
                    crds.externalWorkloadCrd,
                    `ew-job-${workloadName}`,
                    namespace,
                    60
                );
            }

            test.afterAll(async () => {
                logger.debug('Deleting all workloads from all projects');
                const promises = Array.from(jobsByProject.entries()).flatMap(([project, jobs]) =>
                    jobs.map((workloadName) => deleteClusterBatchJob(workloadName, project))
                );
                if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.11')) {
                    promises.push(operations.updateClusterSyncConfig(null, null, true));
                }
                await Promise.all(promises);
            });

            test('Validate jobs and pods are as expected in the backend', async () => {
                await test.step('Submit a workload and verify it appears in the control plane', async () => {
                    logger.info(`Submitting pre-sync workload to project ${projects[0].name}`);
                    preSyncWorkloadName = await submitClusterBatchJob(projects[0]);
                    logger.info(
                        `Pre-sync workload submitted: ${preSyncWorkloadName}, waiting for it in backend`
                    );
                    const backendWorkload = await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                        preSyncWorkloadName,
                        projects[0].name,
                        120
                    );
                    preSyncWorkloadBackendId = backendWorkload.id;
                    logger.info(
                        `Pre-sync workload confirmed in backend with id: ${preSyncWorkloadBackendId}`
                    );
                });

                await test.step('Disable cluster-sync and configure batch size', async () => {
                    logger.info('Disabling cluster-sync and setting batch size to 3');
                    await operations.updateClusterSyncConfig('3', '3', false);
                    await operations.toggleRunaiOperand(operations.RunaiClusterSyncName, false);
                    logger.info('Cluster-sync disabled');
                });

                await test.step('Submit workloads to all projects while cluster-sync is off', async () => {
                    for (const project of projects) {
                        logger.info(
                            `Submitting ${numberOfJobsInEachProject} workloads to project ${project.name}`
                        );
                        const promisesToAwait = [];
                        for (let index = 0; index < numberOfJobsInEachProject; index++) {
                            promisesToAwait.push(submitClusterBatchJob(project));
                        }

                        const workloads = await Promise.all(promisesToAwait);
                        jobsByProject.set(project, workloads);
                        logger.info(
                            `Submitted ${workloads.length} workloads to project ${project.name}: ${workloads.map((w) => w.name).join(', ')}`
                        );
                    }
                });

                await test.step('Delete the pre-sync workload from the cluster while cluster-sync is off', async () => {
                    logger.info(
                        `Deleting pre-sync workload ${preSyncWorkloadName} from cluster while cluster-sync is off`
                    );
                    await deleteClusterBatchJob(preSyncWorkloadName, projects[0]);
                    logger.info(`Pre-sync workload ${preSyncWorkloadName} deleted from cluster`);
                });

                await test.step('Wait for workloads to be registered in the cluster', async () => {
                    logger.info('Sleeping 5s to let workloads register in the cluster');
                    await waitUtils.sleep(5000);
                });

                await test.step('Re-enable cluster-sync and wait for configuration', async () => {
                    const envVarName = versionUtils.isVersionSupported(
                        flowRootState.cluster.version,
                        SupportedVersions.V2_23
                    )
                        ? 'MAX_NUM_OF_WORKLOADS_PER_REQUEST'
                        : 'MAX_NUM_OF_POD_GROUPS_PER_REQUEST';

                    logger.info(`Re-enabling cluster-sync, expecting env var ${envVarName}=3`);
                    await operations.toggleRunaiOperand(operations.RunaiClusterSyncName, true);
                    await waitUtils.waitForDeploymentToBeAvailable(
                        operations.RunaiClusterSyncName,
                        'runai',
                        90
                    );
                    if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.23')) {
                        await operations.waitForRunaiServiceEnvVarConfiguration(
                            operations.RunaiClusterSyncName,
                            '3',
                            envVarName
                        );
                    }
                    logger.info('Cluster-sync re-enabled and configured');
                });

                for (const [project, jobs] of Array.from(jobsByProject)) {
                    await test.step(`Verify workloads synced for project ${project.name}`, async () => {
                        logger.info(`Verifying ${jobs.length} workloads synced for project ${project.name}`);
                        const promisesToAwait = [];
                        for (const workloadName of jobs) {
                            promisesToAwait.push(
                                workloadsUtilsV2.waitForBackendWorkloadAndPodsWithCorrectProjectAndNamespace(
                                    workloadName,
                                    project,
                                    Utils.namespaceFromProjectName(project.name),
                                    120
                                )
                            );
                        }
                        await Promise.all(promisesToAwait);
                        logger.info(`All workloads synced for project ${project.name}`);
                    });
                }

                await test.step('Verify pre-sync workload was deleted from the control plane', async () => {
                    logger.info(
                        `Waiting for pre-sync workload ${preSyncWorkloadName} (id: ${preSyncWorkloadBackendId}) to appear as deleted in control plane`
                    );
                    await workloadsUtilsV2.waitForDeletedWorkload(
                        preSyncWorkloadBackendId,
                        projects[0].id.toString(),
                        undefined,
                        60
                    );
                    logger.info(
                        `Pre-sync workload ${preSyncWorkloadName} confirmed deleted from control plane`
                    );
                });
            });
        });
    }
);
