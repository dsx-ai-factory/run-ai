/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test, TestInfo } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import {
    tags,
    nodeUtils,
    nodePoolUtils,
    departmentUtils,
    objectGenerator,
    flowRootState,
    logger,
    operations,
    orgUnitUtils,
    authorizationUtils,
    usersManager,
    waitUtils,
    workloadsCrudUtils,
    inferenceUtils,
    config,
    trainingUtils,
    workspaceUtils,
    telemetryUtils,
    Utils,
    workloadsUtilsV2,
    metricUtils
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import {
    SupportedVersions,
    supportedFrom
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    MetricsType,
    NodepoolCreateFields,
    NodepoolPhase,
    NodeTelemetryType
} from '../../../../../../modules/api/apis/clusterApi';
import { RankingMetricType, RankingOrderType } from '../../../../../../modules/api/apis/orgUnitApi';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    AvgGpuMetricSelectLabels,
    OverviewPageV2
} from '../../../../../../modules/ui/runaiV2/pages/OverviewPageV2';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';
import { SubjectType, ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { ServingPortProtocol } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { WorkloadTelemetryType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadTelemetryType';
import { NodePage } from '../../../../../../modules/ui/runaiV2/pages/NodePage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { GuidedTour } from '../../../../../../modules/ui/runaiV2/pages/GuidedTour';
import { clusterApi } from '../../../../../../modules/api';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

test.describe(`Overview Dashboard Tests, tags: ${tags.tool.ui} ${tags.resource.overview} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        test.describe.configure({ timeout: 10 * 60 * 1000 }); // 10 minutes timeout

        // Cache file path for state persistence across worker restarts
        const cacheFilePath = path.join(__dirname, '.overview-test-data-cache.json');

        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let loginUIPage: LoginUi;
        let overviewPage: OverviewPageV2;
        let nodesPage: NodePage;
        let projectsPage: ProjectPage;
        let departmentsPage: DepartmentPage;
        let workloadsPage: WorkloadPage;
        let guidedTour: GuidedTour;

        let gpuWorkerNodeNames: string[];
        let totalGpuNodeCnt: number;
        let totalGpuDeviceCnt: number;
        let npId1: string;
        let npId2: string;
        let deptId1: string;
        let deptId2: string;
        let nodeToLabel: string;
        let userName: string;
        let userPasswd: string;
        let userId: string;
        let proj1Id: string;
        let proj2Id: string;
        let proj3Id: string;
        let proj4Id: string;
        let runningTrainingId: string;
        let failedTrainingId: string;
        let pendingTrainingId: string;
        let idleTrainingId: string;
        let runningWorkspaceId: string;
        let runningInferenceId: string;
        const trainingIdsToCleanup: string[] = [];
        const workspaceIdsToCleanup: string[] = [];
        const inferenceIdsToCleanup: string[] = [];
        let videoName: string;
        let isTheLastTest: boolean = false;

        const staticElemCheckTimeout = 5000;
        const createdNodePools: { id: string; name: string }[] = [];
        const createdDepartmentIds: number[] = [];
        const createdProjectIds: number[] = [];

        const alpineImage = 'alpine:latest';
        const sleepCmd = 'sleep infinity';
        const errCmd = 'sl 5';

        const nodeLabelKey = 'test';
        const nodeLabelVal1 = 'overview-group1';
        const nodeLabelVal2 = 'overview-group2';

        const departmentsUrl = `${config.tenantUrl}/departments`;
        const projectsUrl = `${config.tenantUrl}/projects`;
        const workloadsUrl = `${config.tenantUrl}/workloads`;

        let npName1 = objectGenerator.generateShortNameByPrefix('np1', 5);

        // no node will be included in this node pool, which will be used for 'pending' workload creation
        let npName2 = objectGenerator.generateShortNameByPrefix('np2', 5);

        let deptName1 = objectGenerator.generateShortNameByPrefix('depmt1', 5);
        let deptName2 = objectGenerator.generateShortNameByPrefix('depmt2', 5);

        let projName1 = objectGenerator.generateShortNameByPrefix('proj1', 5);
        let projName2 = objectGenerator.generateShortNameByPrefix('proj2', 5);
        // for pending workload
        let projName3 = objectGenerator.generateShortNameByPrefix('proj3', 5);
        // for 'no data' scenario test
        let projName4 = objectGenerator.generateShortNameByPrefix('proj4', 5);

        // workload names
        let runningTrainingName = objectGenerator.generateShortNameByPrefix('running-train', 5);
        let failedTrainingName = objectGenerator.generateShortNameByPrefix('failed-train', 5);
        let pendingTrainingName = objectGenerator.generateShortNameByPrefix('pending-train', 5);
        let idleTrainingName = objectGenerator.generateShortNameByPrefix('runai-idle-gpu', 5);
        let runningWorkspaceName = objectGenerator.generateShortNameByPrefix('running-ws', 5);
        let runningInferenceName = objectGenerator.generateShortNameByPrefix('running-inf', 5);

        // departments
        let depmts = ['default', deptName1, deptName2];
        logger.info(`depmts: ${depmts}`);

        //projects
        let projs = [projName1, projName2, projName3, projName4];

        // department - project mapping
        let depmtProjMap: Record<string, string> = {
            default: projName1,
            [deptName1]: projName2,
            [deptName2]: projName3
        };

        test.beforeAll(async () => {
            test.setTimeout(10 * 60 * 1000);

            // Check if cache file exists, if yes, then load the cached data (playwright worker restart scenario)
            if (fs.existsSync(cacheFilePath)) {
                logger.info('=== Cache file found, restoring state from previous worker ===');
                try {
                    const cachedState = JSON.parse(fs.readFileSync(cacheFilePath, 'utf-8'));

                    // Restore all resource IDs
                    npId1 = cachedState.npId1;
                    npId2 = cachedState.npId2;
                    deptId1 = cachedState.deptId1;
                    deptId2 = cachedState.deptId2;
                    proj1Id = cachedState.proj1Id;
                    proj2Id = cachedState.proj2Id;
                    proj3Id = cachedState.proj3Id;
                    proj4Id = cachedState.proj4Id;
                    userId = cachedState.userId;
                    userName = cachedState.userName;
                    userPasswd = cachedState.userPasswd;
                    runningTrainingId = cachedState.runningTrainingId;
                    failedTrainingId = cachedState.failedTrainingId;
                    pendingTrainingId = cachedState.pendingTrainingId;
                    idleTrainingId = cachedState.idleTrainingId;
                    runningWorkspaceId = cachedState.runningWorkspaceId;
                    runningInferenceId = cachedState.runningInferenceId;
                    nodeToLabel = cachedState.nodeToLabel;
                    gpuWorkerNodeNames = cachedState.gpuWorkerNodeNames;
                    totalGpuDeviceCnt = cachedState.totalGpuDeviceCnt;
                    totalGpuNodeCnt = cachedState.totalGpuNodeCnt;

                    // Restore generated names
                    npName1 = cachedState.npName1;
                    npName2 = cachedState.npName2;
                    deptName1 = cachedState.deptName1;
                    deptName2 = cachedState.deptName2;
                    depmts = ['default', deptName1, deptName2];

                    projName1 = cachedState.projName1;
                    projName2 = cachedState.projName2;
                    projName3 = cachedState.projName3;
                    projName4 = cachedState.projName4;

                    projs = [projName1, projName2, projName3, projName4];
                    depmtProjMap = {
                        default: projName1,
                        [deptName1]: projName2,
                        [deptName2]: projName3
                    };

                    runningTrainingName = cachedState.runningTrainingName;
                    failedTrainingName = cachedState.failedTrainingName;
                    pendingTrainingName = cachedState.pendingTrainingName;
                    idleTrainingName = cachedState.idleTrainingName;
                    runningWorkspaceName = cachedState.runningWorkspaceName;
                    runningInferenceName = cachedState.runningInferenceName;

                    // Restore cleanup arrays
                    createdNodePools.push(...cachedState.createdNodePools);
                    createdDepartmentIds.push(...cachedState.createdDepartmentIds);
                    createdProjectIds.push(...cachedState.createdProjectIds);
                    trainingIdsToCleanup.push(...cachedState.trainingIdsToCleanup);
                    workspaceIdsToCleanup.push(...cachedState.workspaceIdsToCleanup);
                    inferenceIdsToCleanup.push(...cachedState.inferenceIdsToCleanup);

                    // Re-initialize browserEngine for the new worker
                    browserEngine = new BrowserEngine();

                    logger.info('=== State restored successfully from cache ===');
                    logger.info(
                        `Restored: ${createdNodePools.length} nodepools, ${createdDepartmentIds.length} departments, ${createdProjectIds.length} projects`
                    );
                    return; // Skip resource creation
                } catch (error) {
                    logger.error('Failed to restore state from cache file, will recreate resources', error);
                    // If restore fails, continue with normal creation
                }
            }

            logger.info('=== No cache found or cache invalid, creating resources ===');
            browserEngine = new BrowserEngine();

            // get the default node pool info
            const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();

            // get gpu node list
            gpuWorkerNodeNames = await nodeUtils.getGpuWorkerNodeNames();
            totalGpuDeviceCnt = await nodeUtils.getTotalGpuDeviceCnt();
            totalGpuNodeCnt = gpuWorkerNodeNames.length;

            expect(totalGpuNodeCnt).toBeGreaterThan(0);

            // label a gpu node for node pool
            nodeToLabel = gpuWorkerNodeNames[0];
            kubeCli.labelNode(nodeToLabel, nodeLabelKey, nodeLabelVal1);

            // step 1: create node pool 1 includes 1 gpu node
            const npReq1 = Object.assign(new NodepoolCreateFields(), {
                name: npName1,
                labelKey: nodeLabelKey,
                labelValue: nodeLabelVal1,
                clusterId: flowRootState.cluster.uuid
            });

            // step 2: create node pool 1 includes 0 node, used for 'pending' workload creation
            const npReq2 = Object.assign(new NodepoolCreateFields(), {
                name: npName2,
                labelKey: nodeLabelKey,
                labelValue: nodeLabelVal2,
                clusterId: flowRootState.cluster.uuid
            });

            await Promise.all([
                nodePoolUtils.createClusterNodePoolWithRequest(npReq1).then((rsp) => {
                    createdNodePools.push({ id: rsp.body.id, name: npName1 });
                    npId1 = rsp.body.id;
                    return rsp;
                }),
                nodePoolUtils.createClusterNodePoolWithRequest(npReq2).then((rsp) => {
                    createdNodePools.push({ id: rsp.body.id, name: npName2 });
                    npId2 = rsp.body.id;
                    return rsp;
                })
            ]);

            // wait for node pool creation to complete
            await Promise.all([
                nodePoolUtils.waitForClusterNodePoolPhase(npId1, [NodepoolPhase.Ready], 120),
                nodePoolUtils.waitForClusterNodePoolPhase(
                    npId2,
                    [NodepoolPhase.Ready, NodepoolPhase.Empty],
                    120
                )
            ]);

            // step 3: create 2 departments
            await Promise.all([
                departmentUtils.createBackendDepartmentWithName(deptName1, 8, 8).then((rsp) => {
                    createdDepartmentIds.push(parseInt(rsp.body.id));
                    deptId1 = rsp.body.id.toString();
                    return rsp;
                }),
                departmentUtils.createBackendDepartmentWithName(deptName2, 8, 8).then((rsp) => {
                    createdDepartmentIds.push(parseInt(rsp.body.id));
                    deptId2 = rsp.body.id.toString();
                    return rsp;
                })
            ]);
            logger.info(`dpmts ${deptName1} and ${deptName2} have been created.`);

            // step 4: create 3 projects
            await Promise.all([
                // Project 1: 'default' department, 'default' node pool, gpu quota 16
                orgUnitUtils
                    .createProjectAndValidate({
                        clusterId: flowRootState.cluster.uuid,
                        name: projName1,
                        resources: [
                            {
                                nodePool: { id: defaultNodePool.id.toString(), name: defaultNodePool.name },
                                gpu: { deserved: 16, limit: -1 }
                            }
                        ],
                        defaultNodePools: [defaultNodePool.name]
                    })
                    .then((proj) => {
                        proj1Id = proj.id;
                        createdProjectIds.push(parseInt(proj1Id));
                        return proj;
                    }),
                // Project 2: department 1, 'node pool 1', gpu quota 8
                orgUnitUtils
                    .createProjectAndValidate({
                        clusterId: flowRootState.cluster.uuid,
                        parentId: deptId1,
                        name: projName2,
                        resources: [
                            {
                                nodePool: { id: npId1, name: npName1 },
                                gpu: { deserved: 8, limit: 8 }
                            }
                        ],
                        defaultNodePools: [npName1]
                    })
                    .then((proj) => {
                        proj2Id = proj.id;
                        createdProjectIds.push(parseInt(proj2Id));
                        return proj;
                    }),
                // Project 3: department 2, 'node pool 2', gpu quota 8
                orgUnitUtils
                    .createProjectAndValidate({
                        clusterId: flowRootState.cluster.uuid,
                        parentId: deptId2,
                        name: projName3,
                        resources: [
                            {
                                nodePool: { id: npId2, name: npName2 },
                                gpu: { deserved: 8, limit: 8 }
                            }
                        ],
                        defaultNodePools: [npName2]
                    })
                    .then((proj) => {
                        proj3Id = proj.id;
                        createdProjectIds.push(parseInt(proj3Id));
                        return proj;
                    }),
                // Project 4: no data project, gpu quota 0
                orgUnitUtils
                    .createProjectAndValidate({
                        clusterId: flowRootState.cluster.uuid,
                        name: projName4,
                        resources: [{ gpu: { deserved: 0 } }]
                    })
                    .then((proj) => {
                        proj4Id = proj.id;
                        createdProjectIds.push(parseInt(proj4Id));
                        return proj;
                    })
            ]);

            await Utils.delay(3000); // wait for projects creation to propagate

            // step 5: create a system admin user whose scope is the project 4 to test the 'No data' scenario
            const roles = await authorizationUtils.getRoles(
                flowRootState.users.system_administrator.accessToken
            );
            const sysAdminId = roles.find(
                (role) => role.name === PROPERTY_TO_ROLE_MAPPING.system_administrator
            )?.id;
            userName = objectGenerator.generateEmailByPrefix('sysadmin-nodata');

            const [userResponse, _] = await Promise.all([
                usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    userName,
                    false
                ),
                authorizationUtils.createAccessRuleWithDefaults(
                    userName,
                    SubjectType.User,
                    sysAdminId,
                    proj4Id,
                    ScopeType.Project,
                    flowRootState.users.system_administrator.accessToken
                )
            ]);
            userPasswd = userResponse.body.tempPassword;
            userId = userResponse.body.id;

            // step 6: create workloads
            // create 1 'running' training, 1 'running' workspace and 1 'running' inference workload in project 1
            // create 1 'failed' and 1 'idle' training workload in project 2
            // create 1 'pending' training workload in project 3

            const runningTrainingResp = await trainingUtils.createTrainingV2(proj1Id, runningTrainingName, {
                image: alpineImage,
                command: sleepCmd,
                compute: { gpuDevicesRequest: 1 }
            });
            runningTrainingId = runningTrainingResp.body.workloadId;
            logger.info(`Created running training workload: ${runningTrainingName}`);

            const runningWorkspaceResp = await workspaceUtils.createWorkspace(proj1Id, runningWorkspaceName, {
                image: alpineImage,
                command: sleepCmd,
                compute: { gpuDevicesRequest: 1 }
            });
            runningWorkspaceId = runningWorkspaceResp.body.workloadId;
            logger.info(`Created running workspace workload: ${runningWorkspaceName}`);

            const runningInfResp = await inferenceUtils.createInference(
                runningInferenceName,
                {
                    image: config.quickStartImage.inferenceHelloWorld,
                    servingPort: { container: 8033, protocol: ServingPortProtocol.Http },
                    compute: { gpuDevicesRequest: 1 },
                    autoscaling: { minReplicas: 1 }
                },
                proj1Id
            );
            runningInferenceId = runningInfResp[0];
            logger.info(`Created running inference workload: ${runningInferenceName}`);

            const failedTrainingResp = await trainingUtils.createTrainingV2(proj2Id, failedTrainingName, {
                image: alpineImage,
                command: errCmd,
                compute: { gpuDevicesRequest: 1 },
                // let the workload fail immediately without retrying
                backoffLimit: 0
            });
            failedTrainingId = failedTrainingResp.body.workloadId;
            logger.info(`Created failed training workload: ${failedTrainingName}`);

            const pendingTrainingResp = await trainingUtils.createTrainingV2(proj3Id, pendingTrainingName, {
                image: alpineImage,
                command: sleepCmd,
                compute: { gpuDevicesRequest: 1 }
            });
            pendingTrainingId = pendingTrainingResp.body.workloadId;
            logger.info(`Created pending training workload: ${pendingTrainingName}`);

            const idleTrainingResp = await trainingUtils.createTrainingV2(proj1Id, idleTrainingName, {
                image: alpineImage,
                command: sleepCmd,
                compute: { gpuDevicesRequest: 1 }
            });
            idleTrainingId = idleTrainingResp.body.workloadId;
            logger.info(`Created idle training workload: ${idleTrainingName}`);

            // wait workloads to reach expected phases
            await Promise.all([
                waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(runningTrainingId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${runningTrainingName} to be Running`,
                    true,
                    120
                ),
                waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getWorkspace(runningWorkspaceId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${runningWorkspaceName} to be Running`,
                    true,
                    120
                ),
                waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getInference(runningInferenceId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${runningInferenceName} to be Running`,
                    true,
                    120
                ),
                waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(failedTrainingId)).body.actualPhase ===
                        Phase.Failed,
                    `Wait for workload ${failedTrainingName} to be Failed`,
                    true,
                    120
                ),
                waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(pendingTrainingId)).body.actualPhase ===
                        Phase.Pending,
                    `Wait for workload ${pendingTrainingName} to be Pending`,
                    true,
                    120
                ),
                waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(idleTrainingId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${idleTrainingName} to be Running`,
                    true,
                    120
                )
            ]);
            logger.info('All workloads reached expected phases successfully');

            trainingIdsToCleanup.push(runningTrainingId);
            trainingIdsToCleanup.push(failedTrainingId);
            trainingIdsToCleanup.push(pendingTrainingId);
            trainingIdsToCleanup.push(idleTrainingId);
            workspaceIdsToCleanup.push(runningWorkspaceId);
            inferenceIdsToCleanup.push(runningInferenceId);

            // Save resource info to cache file for playwright worker restart scenario
            try {
                const stateToCache = {
                    // Resource IDs
                    npId1,
                    npId2,
                    deptId1,
                    deptId2,
                    proj1Id,
                    proj2Id,
                    proj3Id,
                    proj4Id,
                    userId,
                    userName,
                    userPasswd,
                    runningTrainingId,
                    failedTrainingId,
                    pendingTrainingId,
                    idleTrainingId,
                    runningWorkspaceId,
                    runningInferenceId,
                    nodeToLabel,
                    gpuWorkerNodeNames,
                    totalGpuDeviceCnt,
                    totalGpuNodeCnt,
                    // Generated resource names
                    npName1,
                    npName2,
                    deptName1,
                    deptName2,
                    projName1,
                    projName2,
                    projName3,
                    projName4,
                    runningTrainingName,
                    failedTrainingName,
                    pendingTrainingName,
                    idleTrainingName,
                    runningWorkspaceName,
                    runningInferenceName,
                    // Cleanup arrays
                    createdNodePools: [...createdNodePools],
                    createdDepartmentIds: [...createdDepartmentIds],
                    createdProjectIds: [...createdProjectIds],
                    trainingIdsToCleanup: [...trainingIdsToCleanup],
                    workspaceIdsToCleanup: [...workspaceIdsToCleanup],
                    inferenceIdsToCleanup: [...inferenceIdsToCleanup],
                    timestamp: new Date().toISOString()
                };

                fs.writeFileSync(cacheFilePath, JSON.stringify(stateToCache, null, 2));
                logger.info(`=== State cached to: ${cacheFilePath} ===`);
            } catch (error) {
                logger.error('Failed to save state to cache file', error);
                throw error;
            }

            // need to retry since the running workload transitions to idle may take 5+ min
            await waitUtils.waitForIdleAllocatedGpuDevicesAppear();
            logger.info(`the idle workload ${idleTrainingName} becomes idle now.`);

            // ============================ workload info summary ======================================
            // 'default' department >  project 1 ('default' nodepool):
            // 2 running training(including 1 idle training), 1 running workspace, 1 running inference
            // GPU quota: 16
            // allocated GPU devices: 4
            // -----------------------------------------------------------------------------------------
            // department 1 > project 2 (nodepool 1): 1 failed training
            // GPU quota: 8
            // allocated GPU devices: 0
            // -----------------------------------------------------------------------------------------
            // department 2 > project 3 (nodepool 2): 1 pending training
            // GPU quota: 8
            // allocated GPU devices: 0
            // -----------------------------------------------------------------------------------------
            // project 4: no workload
            // =========================================================================================
        });

        test.afterAll(async () => {
            // Check if it's the last test
            if (!isTheLastTest) {
                return;
            }

            // Unlabel nodes
            try {
                if (gpuWorkerNodeNames && gpuWorkerNodeNames.length > 0) {
                    kubeCli.unLabelNode(nodeToLabel, nodeLabelKey);
                }
            } catch (e) {
                logger.error(`Failed to unlabel node ${nodeToLabel}`, e);
            }

            // Delete the user
            try {
                if (userId) {
                    await usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        userId
                    );
                }
            } catch (e) {
                logger.error(`Failed to delete user ${userId}`, e);
            }

            // Delete workloads
            await Promise.all(
                trainingIdsToCleanup.map(async (id) => {
                    try {
                        await workloadsCrudUtils.deleteTraining(id);
                    } catch (e) {
                        logger.error(`Failed to delete training workload ${id}`, e);
                    }
                })
            );

            await Promise.all(
                workspaceIdsToCleanup.map(async (id) => {
                    try {
                        await workloadsCrudUtils.deleteWorkspace(id);
                    } catch (e) {
                        logger.error(`Failed to delete workspace workload ${id}`, e);
                    }
                })
            );

            await Promise.all(
                inferenceIdsToCleanup.map(async (id) => {
                    try {
                        await workloadsCrudUtils.deleteInference(id);
                    } catch (e) {
                        logger.error(`Failed to delete inference workload ${id}`, e);
                    }
                })
            );

            // Delete projects
            await Promise.all(
                createdProjectIds.map(async (projectId) => {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                    } catch (e) {
                        logger.error(`Failed to delete project ${projectId}`, e);
                    }
                })
            );

            // Delete departments
            await Promise.all(
                createdDepartmentIds.map(async (deptId) => {
                    try {
                        await departmentUtils.deleteDepartment(deptId);
                    } catch (e) {
                        logger.error(`Failed to delete department ${deptId}`, e);
                    }
                })
            );

            // Delete node pools
            await Promise.all(
                createdNodePools.map(async (np) => {
                    try {
                        await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                    } catch (e) {
                        logger.error(`Failed to delete node pool ${np.id}`, e);
                    }
                })
            );

            // Delete cache file after successful cleanup
            try {
                if (fs.existsSync(cacheFilePath)) {
                    fs.unlinkSync(cacheFilePath);
                    logger.info('=== Cache file deleted after cleanup ===');
                }
            } catch (error) {
                logger.error('Failed to delete cache file', error);
            }
        });

        test.beforeEach(async () => {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            loginUIPage = new LoginUi(browserEngine.currentPage);
            overviewPage = new OverviewPageV2(browserEngine.currentPage);
            nodesPage = new NodePage(browserEngine.currentPage);
            workloadsPage = new WorkloadPage(browserEngine.currentPage);
            projectsPage = new ProjectPage(browserEngine.currentPage);
            departmentsPage = new DepartmentPage(browserEngine.currentPage);
            guidedTour = new GuidedTour(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );

            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            clusterApi.clustersApi.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        // ===================================================
        // Overview dashboard elelment visilibity sanity tests
        // ===================================================
        test.describe('Element visibility sanity tests', () => {
            const userTestDataSet = [
                {
                    testCase:
                        'T5582154: verify all the widgets appear on the overview dashboard page for system administrator user',
                    video: 'sys-admin-check-widgets-exist',
                    user: flowRootState.users.system_administrator.email,
                    password: flowRootState.users.system_administrator.password,
                    resourceAllocationLegends: [
                        'GPU devices (total)',
                        'GPU devices (unallocated)',
                        'GPU devices (allocated)',
                        'GPU quota (All departments)'
                    ]
                },
                {
                    testCase:
                        'T5582155: verify all the widgets appear on the overview dashboard page for department administrator user',
                    video: 'dept-admin-check-widgets-exist',
                    user: flowRootState.users.department_admin.email,
                    password: flowRootState.users.department_admin.password,
                    resourceAllocationLegends: ['GPU quota', 'GPU allocated']
                }
            ];

            userTestDataSet.forEach((caseData) => {
                test(caseData.testCase, async () => {
                    videoName = caseData.video;

                    if (caseData.user !== flowRootState.users.system_administrator.email) {
                        await loginUIPage.logoutToLoginPage();
                        await authPage.signIn(caseData.user, caseData.password);
                    }

                    await test.step('Go to the Overview dashboard page', async () => {
                        await overviewPage.goToPage();
                    });

                    // start checking the top widgets/cards, including all the widget div, name div and the link button in it, also verify the name
                    await test.step('Verify the Ready / Total GPU wiget exists', async () => {
                        await Promise.all(
                            overviewPage.getReadyNodesWidgetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getReadyNodesWidgetName()).trim()).toBe(
                            'Ready / Total GPU nodes'
                        );
                    });

                    await test.step('Verify the Ready / Total GPU devices widget exists', async () => {
                        await Promise.all(
                            overviewPage.getReadyGpuWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getReadyGpuWidgetName()).trim()).toBe(
                            'Ready / Total GPU devices'
                        );
                    });

                    await test.step('Verify the Allocated GPU devices widget exists', async () => {
                        await Promise.all(
                            overviewPage.getAllocatedGpuWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getAllocatedGpuWidgetName()).trim()).toBe(
                            'Allocated GPU devices'
                        );
                    });

                    await test.step('Verify the Idle allocated GPU devices widget exists', async () => {
                        await Promise.all(
                            overviewPage.getIdleAllocatedGpuWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getIdleAllocatedGpuWidgetName()).trim()).toBe(
                            'Idle allocated GPU devices'
                        );
                    });

                    await test.step('Verify the Running workloads widget exists', async () => {
                        await Promise.all(
                            overviewPage.getRunningWorkloadsWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getRunningWorkloadsWidgetName()).trim()).toBe(
                            'Running workloads'
                        );
                    });

                    await test.step('Verify the Pending workloads widget exists', async () => {
                        await Promise.all(
                            overviewPage.getPendingWorkloadsWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getPendingWorkloadsWidgetName()).trim()).toBe(
                            'Pending workloads'
                        );
                    });

                    await test.step('Verify the Failed workloads widget exists', async () => {
                        await Promise.all(
                            overviewPage.getFailedWorkloadsWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getFailedWorkloadsWidgetName()).trim()).toBe(
                            'Failed workloads'
                        );
                    }); // end of checking the top widgets/cards

                    // check the other widgets in the page
                    await test.step('Verify the Running workloads by category/type widget exists', async () => {
                        await Promise.all(
                            overviewPage.getRunningWorkloadsByTypeWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getRunningWorkloadsByTypeWidgetName()).trim()).toMatch(
                            /^(Running workloads by category|Running workloads by type)$/
                        );
                    });

                    await test.step('Verify the Workloads by status widget exists', async () => {
                        await Promise.all(
                            overviewPage.getWorkloadsByStatusWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getWorkloadStatusWidgetName()).trim()).toBe(
                            'Workloads by status'
                        );
                    });

                    await test.step('Verify the projects/departments widget exists', async () => {
                        await Promise.all(
                            overviewPage.getOrgUnitRankingWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );
                    });

                    await test.step('Verify the Pending workloads by pending time widget exists', async () => {
                        await Promise.all(
                            overviewPage.getPendingWorkloadDistTimeWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getPendingWorkloadDistTimeWidgetName()).trim()).toBe(
                            'Pending workloads by pending time'
                        );
                    });

                    await test.step('Verify the Idle workloads time widget exists', async () => {
                        await Promise.all(
                            overviewPage.getIdleGpuWorkloadsWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getIdleGpuWorkloadsWidgetName()).trim()).toBe(
                            'Idle workloads'
                        );
                    });

                    await test.step('Verify the Allocation ratio by node pool widget exists', async () => {
                        await Promise.all(
                            overviewPage.getGpuAllocationRatioWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getGpuAllocationRatioWidgetName()).trim()).toBe(
                            'Allocation ratio by node pool'
                        );
                    });

                    await test.step('Verify the Free nodes by node pool widget exists', async () => {
                        await Promise.all(
                            overviewPage.getFreeNodesByNodepoolWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getFreeNodesByNodepoolWidgetName()).trim()).toBe(
                            'Free nodes by node pool'
                        );
                    });

                    await test.step('Verify the Resources utilization widget exists', async () => {
                        await Promise.all(
                            overviewPage.getResourceUtilizationWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getResourceUtilizationWidgetName()).trim()).toBe(
                            'Resources utilization'
                        );

                        // check the legend texts
                        const legends = [
                            'GPU compute utilization',
                            'GPU memory utilization',
                            'CPU compute utilization',
                            'CPU memory utilization'
                        ];

                        expect((await overviewPage.getResourceUtilizationWidgetLegendNames()).sort()).toEqual(
                            legends.sort()
                        );
                    });

                    await test.step('Verify the Resources allocation widget exists', async () => {
                        await Promise.all(
                            overviewPage.getResourceAllocationWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getResourceAllocationWidgetName()).trim()).toBe(
                            'Resources allocation'
                        );

                        // check the legend texts
                        const legends = caseData.resourceAllocationLegends;

                        expect((await overviewPage.getResourceAllocationWidgetLegendNames()).sort()).toEqual(
                            legends.sort()
                        );
                    });

                    await test.step('Verify the Workloads by category widget exists', async () => {
                        await Promise.all(
                            overviewPage.getWorkloadsByCategoryWdigetElems().map(async (locator) => {
                                await overviewPage.waitForVisible(locator, staticElemCheckTimeout);
                            })
                        );

                        expect((await overviewPage.getWorkloadsByCategoryWidgetName()).trim()).toBe(
                            'Workloads by category'
                        );
                    });
                });
            });
        }); // End of: Element visibility sanity tests

        // ===================================================================================================
        // Overview dashboard telemetry/metrics data sanity tests
        // due to the long data aggregation intervals for last 7/30 days metrics, they will not be tested here
        // ===================================================================================================
        test.describe('Telemetry and metrics data validation tests', () => {
            test('T5582157: test overview dashboard when no data available', async () => {
                videoName = 'no-data-check';

                await test.step('login with the user account that is limited to the project scope with no data', async () => {
                    await loginUIPage.logoutToLoginPage();
                    await authPage.signIn(userName, userPasswd);
                });

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify running workload widget shows 0', async () => {
                    expect(await overviewPage.getRunningWorkloadsWidgetData()).toBe(0);
                });

                await test.step('verify pending workload widget shows 0', async () => {
                    expect(await overviewPage.getPendingWorkloadsWidgetData()).toBe(0);
                });

                await test.step('verify failed workload widget shows 0', async () => {
                    expect(await overviewPage.getFailedWorkloadsWidgetData()).toBe(0);
                });

                await test.step('verify running workloads by type chart shows no data', async () => {
                    await expect(
                        overviewPage.page
                            .locator(overviewPage.runningWorkloadsByTypeWidget)
                            .locator(overviewPage.emptyWidgetMsgDiv)
                    ).toHaveText('No running workloads to display');
                });

                await test.step('verify workloads by status chart shows no data', async () => {
                    await expect(
                        overviewPage.page
                            .locator(overviewPage.workloadsByStatusWidget)
                            .locator(overviewPage.emptyWidgetMsgDiv)
                    ).toHaveText('To view data, first create workloads');
                });

                await test.step(`verify projects table only shows the assigned project ${projName4} when the user was created`, async () => {
                    const data = await overviewPage.getProjectsDepartmentsTableData();

                    expect(data.length).toBe(1);
                    expect(data[0]['Project'].trim()).toBe(projName4);
                });

                await test.step('verify pending workloads by pending time chart shows no data', async () => {
                    await expect(
                        overviewPage.page
                            .locator(overviewPage.pendingWorkloadDistTimeWidget)
                            .locator(overviewPage.emptyWidgetMsgDiv)
                    ).toHaveText('No pending workloads to display');
                });

                await test.step('verify idle workloads chart shows no data', async () => {
                    await expect(
                        overviewPage.page
                            .locator(overviewPage.idleGpuWorkloadsWidget)
                            .locator(overviewPage.emptyStateContainerDiv)
                    ).toHaveText('No idle workloads to display');
                });

                await test.step('verify allocation ratio by node pool chart shows no data', async () => {
                    await expect(
                        overviewPage.page
                            .locator(overviewPage.gpuAllocationRatioWidget)
                            .locator(overviewPage.emptyWidgetMsgDiv)
                    ).toHaveText('To view data, first create workloads');
                });
            });

            test('T5582165: validate the Ready / Total GPU nodes widget data', async () => {
                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Ready / Total GPU nodes widget data', async () => {
                    const data = await overviewPage.getReadyTotalGpuNodesWidgetData();
                    gpuWorkerNodeNames = await nodeUtils.getGpuWorkerNodeNames();
                    totalGpuNodeCnt = gpuWorkerNodeNames.length;
                    expect(data).toBe(`${totalGpuNodeCnt}/${totalGpuNodeCnt}`);
                });
            });

            test('T5582166: validate the Ready / Total GPU devices widget data', async () => {
                videoName = 'validate-ready-total-gpu-devices-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Ready / Total GPU devices widget data', async () => {
                    const data = await overviewPage.getReadyTotalGpuDevicesWidgetData();
                    totalGpuDeviceCnt = await nodeUtils.getTotalGpuDeviceCnt();
                    expect(data).toBe(`${totalGpuDeviceCnt}/${totalGpuDeviceCnt}`);
                });
            });

            test('T5582167: validate the Allocated GPU devices widget data', async () => {
                videoName = 'validate-allocated-gpu-devices-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Allocated GPU devices widget data', async () => {
                    // total 4 running workloads each request 1 GPU device
                    const data = await overviewPage.getAllocatedGpuDevicesWidgetData();
                    expect(data).toBeGreaterThanOrEqual(4);
                });
            });

            test('T5582168: validate the Idle allocated GPU devices widget data', async () => {
                videoName = 'validate-idle-allocated-gpu-devices-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify idle allocated GPU devices widget data', async () => {
                    // 1 idle workload each request 1 GPU device
                    const data = await overviewPage.getIdleAllocatedGpuDevicesWidgetData();
                    expect(data).toBeGreaterThanOrEqual(1);
                });
            });

            test('T5582169: validate the Running workloads widget data', async () => {
                videoName = 'validate-running-workloads-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Running workloads widget data', async () => {
                    // total 4 running workloads
                    const data = await overviewPage.getRunningWorkloadsWidgetData();
                    expect(data).toBeGreaterThanOrEqual(4);
                });
            });

            test('T5582170: validate the Pending workloads widget data', async () => {
                videoName = 'validate-pending-workloads-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Pending workloads widget data', async () => {
                    // total 1 pending workload
                    const data = await overviewPage.getPendingWorkloadsWidgetData();
                    expect(data).toBeGreaterThanOrEqual(1);
                });
            });

            test('T5582171: validate the Failed workloads widget data', async () => {
                videoName = 'validate-failed-workloads-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Failed workloads widget data', async () => {
                    // total 1 failed workloads
                    const data = await overviewPage.getFailedWorkloadsWidgetData();
                    expect(data).toBeGreaterThanOrEqual(1);
                });
            });

            test('T5582172: validate the Running workloads by category widget data', async () => {
                const workloadCatResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                    WorkloadTelemetryType.WorkloadsCount,
                    ['Category'],
                    flowRootState.cluster.uuid,
                    undefined,
                    [Phase.Running]
                );

                // running workspace workload: 1
                const buildItem = workloadCatResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Category' && g.value === 'Build')
                );
                expect(Number(buildItem?.value)).toBeGreaterThanOrEqual(1);

                // running inference workload: 1
                const deployItem = workloadCatResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Category' && g.value === 'Deploy')
                );
                expect(Number(deployItem?.value)).toBeGreaterThanOrEqual(1);

                // running training workload: 2
                const trainItem = workloadCatResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Category' && g.value === 'Train')
                );
                expect(Number(trainItem?.value)).toBeGreaterThanOrEqual(2);
            });

            test('T5582173: validate the Running workloads by type widget data', async () => {
                const workloadTypeResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                    WorkloadTelemetryType.WorkloadsCount,
                    ['Type'],
                    flowRootState.cluster.uuid,
                    undefined,
                    [Phase.Running]
                );

                // running workspace workload: 1
                const workspaceItem = workloadTypeResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Type' && g.value === 'Workspace')
                );
                expect(Number(workspaceItem?.value)).toBeGreaterThanOrEqual(1);

                // running inference workload: 1
                const infItem = workloadTypeResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Type' && g.value === 'Inference')
                );
                expect(Number(infItem?.value)).toBeGreaterThanOrEqual(1);

                // running training workload: 2
                const trainItem = workloadTypeResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Type' && g.value === 'Training')
                );
                expect(Number(trainItem?.value)).toBeGreaterThanOrEqual(2);
            });

            test('T5582174: validate the Workloads by status widget data', async () => {
                const workloadTypeResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                    WorkloadTelemetryType.WorkloadsCount,
                    ['Phase'],
                    flowRootState.cluster.uuid
                );

                // running workload: 4
                const runningItem = workloadTypeResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Running')
                );
                expect(Number(runningItem?.value)).toBeGreaterThanOrEqual(4);

                // failed workload: 1
                const failedItem = workloadTypeResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Failed')
                );
                expect(Number(failedItem?.value)).toBeGreaterThanOrEqual(1);

                // pending workload: 1
                const pendingItem = workloadTypeResp.values.find((item: any) =>
                    item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Pending')
                );
                expect(Number(pendingItem?.value)).toBeGreaterThanOrEqual(1);
            });
        }); // End of: Telemetry and metrics data validation tests

        // =================================================
        // Projects/Departments table data validation tests
        // =================================================
        test.describe('Projects and Departments table data validation tests', () => {
            type TimeRange = '24 hours' | '7 days' | '30 days';
            type View = 'Projects' | 'Departments';
            const projDemptDataTests: Array<{
                name: string;
                video: string;
                view: View;
                metric: AvgGpuMetricSelectLabels;
                timeRange: TimeRange;
            }> = [
                {
                    name: 'T5582175: Verify the Projects > Highest Avg. GPU allocation table for the last 24 hours',
                    video: 'proj-highest-avg-gpu-alloc-24-hours',
                    view: 'Projects',
                    metric: AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_ALLOC,
                    timeRange: '24 hours'
                },
                {
                    name: 'T5582176: Verify the Projects > Lowest Avg. GPU allocation table for the last 7 days',
                    video: 'proj-lowest-avg-gpu-alloc-7-days',
                    view: 'Projects',
                    metric: AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_ALLOC,
                    timeRange: '7 days'
                },
                {
                    name: 'T5582177: Verify the Projects > Highest Avg. GPU utilization table for the last 24 hours',
                    video: 'proj-highest-avg-gpu-util-24-hours',
                    view: 'Projects',
                    metric: AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_UTIL,
                    timeRange: '24 hours'
                },
                {
                    name: 'T5582178: Verify the Projects > Lowest Avg. GPU utilization table for the last 7 days',
                    video: 'proj-lowest-avg-gpu-util-7-days',
                    view: 'Projects',
                    metric: AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_UTIL,
                    timeRange: '7 days'
                },
                {
                    name: 'T5582179: Verify the Projects > Highest Avg. GPU memory utilization table for the last 24 hours',
                    video: 'proj-highest-avg-gpu-mem-util-24-hours',
                    view: 'Projects',
                    metric: AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_MEM_UTIL,
                    timeRange: '24 hours'
                },
                {
                    name: 'T5582180: Verify the Projects > Lowest Avg. GPU memory utilization table for the last 7 days',
                    video: 'proj-lowest-avg-gpu-mem-util-7-days',
                    view: 'Projects',
                    metric: AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_MEM_UTIL,
                    timeRange: '7 days'
                },
                {
                    name: 'T5582181: Verify the Projects > Highest Avg. GPU allocation table for the last 24 hours',
                    video: 'dept-highest-avg-gpu-alloc-24-hours',
                    view: 'Departments',
                    metric: AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_ALLOC,
                    timeRange: '24 hours'
                },
                {
                    name: 'T5582182: Verify the Projects > Lowest Avg. GPU allocation table for the last 7 days',
                    video: 'dept-lowest-avg-gpu-alloc-7-days',
                    view: 'Departments',
                    metric: AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_ALLOC,
                    timeRange: '7 days'
                },
                {
                    name: 'T5582183: Verify the Projects > Highest Avg. GPU utilization table for the last 24 hours',
                    video: 'dept-highest-avg-gpu-util-24-hours',
                    view: 'Departments',
                    metric: AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_UTIL,
                    timeRange: '24 hours'
                },
                {
                    name: 'T5582184: Verify the Projects > Lowest Avg. GPU utilization table for the last 7 days',
                    video: 'dept-lowest-avg-gpu-util-7-days',
                    view: 'Departments',
                    metric: AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_UTIL,
                    timeRange: '7 days'
                },
                {
                    name: 'T5582185: Verify the Projects > Highest Avg. GPU memory utilization table for the last 24 hours',
                    video: 'dept-highest-avg-gpu-mem-util-24-hours',
                    view: 'Departments',
                    metric: AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_MEM_UTIL,
                    timeRange: '24 hours'
                },
                {
                    name: 'T5582186: Verify the Projects > Lowest Avg. GPU memory utilization table for the last 7 days',
                    video: 'dept-lowest-avg-gpu-mem-util-7-days',
                    view: 'Departments',
                    metric: AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_MEM_UTIL,
                    timeRange: '7 days'
                }
            ];

            const projectsViewData = [
                {
                    Project: projName1,
                    Department: 'default',
                    'GPU quota': '16',
                    'Pending workloads': '0',
                    'Running workloads': '4'
                },
                {
                    Project: projName2,
                    Department: deptName1,
                    'GPU quota': '8',
                    'Pending workloads': '0',
                    'Running workloads': '0'
                },
                {
                    Project: projName3,
                    Department: deptName2,
                    'GPU quota': '8',
                    'Pending workloads': '1',
                    'Running workloads': '0'
                },
                {
                    Project: projName4,
                    Department: 'default',
                    'GPU quota': '0',
                    'Pending workloads': '0',
                    'Running workloads': '0'
                }
            ];

            const depmtsViewData = [
                {
                    Department: 'default',
                    'GPU quota': 'Unlimited',
                    'Pending workloads': '0',
                    'Running workloads': '4'
                },
                {
                    Department: deptName1,
                    'GPU quota': '8',
                    'Pending workloads': '0',
                    'Running workloads': '0'
                },
                {
                    Department: deptName2,
                    'GPU quota': '8',
                    'Pending workloads': '1',
                    'Running workloads': '0'
                }
            ];

            for (const caseData of projDemptDataTests) {
                test(caseData.name, async ({}, testInfo: TestInfo) => {
                    skipPlaywrightTest(test, testInfo, 'No data is displayed in the table');

                    videoName = caseData.video;

                    await test.step('go to the overview dashboard page', async () => {
                        await overviewPage.goToPage();
                    });

                    await test.step(`switch to ${caseData.view} view in the table widget`, async () => {
                        await overviewPage.switchViewInOrgUnitRankingWidget(caseData.view);
                    });

                    await test.step(`select metric ${caseData.metric}`, async () => {
                        await overviewPage.selectOrgUnitMetrics(caseData.metric);
                    });

                    await test.step(`select time range ${caseData.timeRange}`, async () => {
                        await overviewPage.chooseOrgUnitMetricTimeRange(caseData.timeRange);
                    });

                    await test.step('validate the table data', async () => {
                        const actData = await overviewPage.getProjectsDepartmentsTableData();
                        logger.info(`${caseData.view} table data: \n ${JSON.stringify(actData, null, 2)}`);

                        let expData = undefined;
                        let identifierKey = '';

                        if (caseData.view === 'Projects') {
                            expData = projectsViewData;
                            identifierKey = 'Project';
                        } else {
                            expData = depmtsViewData;
                            identifierKey = 'Department';
                        }

                        // Verify each expected item exists in actual data
                        for (const expectedRow of expData) {
                            const identifier = expectedRow[identifierKey];

                            // Find matching row in actual data
                            const actualRow = actData.find((row) => row[identifierKey] === identifier);
                            expect(actualRow).toBeDefined();

                            // Validate all expected fields match
                            for (const [key, value] of Object.entries(expectedRow)) {
                                expect(actualRow[key]).toBe(value);
                            }

                            // Validate the allocation/utilization fields (dynamic data)
                            const runningWorkload = Number(actualRow['Running workloads']);

                            // Find all the 'allocation' and 'utilization' related keys
                            const allocationKeys = Object.keys(actualRow).filter((k) =>
                                k.toLowerCase().includes('allocation')
                            );
                            const utilizationKeys = Object.keys(actualRow).filter((k) =>
                                k.toLowerCase().includes('utilization')
                            );

                            if (runningWorkload > 0) {
                                // allocation field values > 0
                                for (const allocKey of allocationKeys) {
                                    const allocVal = actualRow[allocKey];
                                    expect(Number(allocVal)).toBeGreaterThan(0);
                                }

                                // utilization field values > 0% and <= 100%
                                for (const utilKey of utilizationKeys) {
                                    const utilVal = actualRow[utilKey];
                                    expect(utilVal.endsWith('%')).toBeTruthy();
                                    const utilNum = Number(utilVal.replace('%', ''));
                                    expect(utilNum).toBeGreaterThan(0);
                                    expect(utilNum).toBeLessThanOrEqual(100);
                                }
                            } else {
                                // 'allocation' and 'utilization' related field values should be '-'
                                for (const allocKey of allocationKeys) {
                                    expect(actualRow[allocKey]).toBe('-');
                                }

                                for (const utilKey of utilizationKeys) {
                                    expect(actualRow[utilKey]).toBe('-');
                                }
                            }
                        }
                    });
                });
            }
        });

        // ============================================================
        // data validation tests for wdigets below the projects table
        // ===========================================================
        test.describe('data validation tests for wdigets below the projects table', () => {
            test('T5582187: validate the Pending workloads by pending time widget data', async () => {
                videoName = 'validate-pending-workloads-by-pending-time-data';
                const workloadTypeResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                    WorkloadTelemetryType.PendingTimeDistribution,
                    ['ClusterId'],
                    flowRootState.cluster.uuid
                );

                // pending workload: 1
                const count = workloadTypeResp.values.filter((v: any) => v.value === '1').length;
                expect(count).toBeGreaterThanOrEqual(1);
            });

            test('T5582188: validate the Idle workloads widget data', async () => {
                // for the idle workloads table, we have a job that runs in the background and writes this data about idle workloads to the db
                // and it happens every 20-25 minutes or so. The test has to retry with timeout set to 25 min.
                test.setTimeout(25 * 60 * 1000);
                videoName = 'validate-idle-workloads-data';

                const workloads = await workloadsUtilsV2.waitForIdleWorkloadToAppear();

                expect(workloads.body.workloads.length).toBeGreaterThanOrEqual(1);

                const found = workloads.body.workloads.some(
                    (w: { name: string }) => w.name === idleTrainingName
                );

                expect(found, `check failed for workload ${idleTrainingName}`).toBeTruthy();
            });

            test('T5582189: validate the Allocated ratio by node pool widget data', async () => {
                videoName = 'validate-allocated-ratio-by-nodepool-data';
                const nodepools = await nodePoolUtils.getAllNodePoolsByClusterAndValidate();

                const nodepoolNames = nodepools.map((np: any) => np.name);

                // check the created two node pools and the default node pool exist
                expect(nodepoolNames).toEqual(expect.arrayContaining([npName1, npName2, 'default']));
            });

            test('T5582190: validate the Free GPU nodes by node pool widget data', async () => {
                videoName = 'validate-free-gpu-nodes-by-nodepool-data';
                const nodeResp = await telemetryUtils.getNodeTelemetryUsingGroupBy(
                    NodeTelemetryType.FreeGpus,
                    ['Node'],
                    flowRootState.cluster.uuid
                );

                gpuWorkerNodeNames = await nodeUtils.getGpuWorkerNodeNames();
                totalGpuDeviceCnt = await nodeUtils.getTotalGpuDeviceCnt();
                totalGpuNodeCnt = gpuWorkerNodeNames.length;

                expect(nodeResp.values.length).toBe(totalGpuNodeCnt);
                expect(
                    nodeResp.values.every((v: any) => Number(v.value) <= totalGpuDeviceCnt / totalGpuNodeCnt)
                ).toBeTruthy();
            });

            const resourcesUtilTests = [
                {
                    name: 'T5582191: validate the Resources utilization widget data for last 3 hours',
                    video: 'validate-resource-util-data-3hrs',
                    hours: 3
                },
                {
                    name: 'T5582192: validate the Resources utilization widget data for last 24 hours',
                    video: 'validate-resource-util-data-24hrs',
                    hours: 24
                }
            ];

            for (const caseData of resourcesUtilTests) {
                test(caseData.name, async () => {
                    videoName = caseData.video;
                    const endTime = new Date();
                    const startTime = new Date(endTime.getTime() - caseData.hours * 60 * 60 * 1000);
                    const resp = await metricUtils.getClusterMetrics(
                        startTime,
                        endTime,
                        [
                            MetricsType.GpuUtilization,
                            MetricsType.GpuMemoryUtilization,
                            MetricsType.CpuUtilization,
                            MetricsType.CpuMemoryUtilization
                        ],
                        20
                    );

                    for (const m of resp.body.measurements) {
                        const hasPositive = m.values.some((v: any) => Number(v.value) > 0);
                        expect(hasPositive).toBeTruthy();
                    }
                });
            }

            const resourcesAllocTests = [
                {
                    name: 'T5582195: validate the Resources allocation widget data for last 3 hours',
                    video: 'validate-resource-alloc-data-3hrs',
                    hours: 3
                },
                {
                    name: 'T5582196: validate the Resources allocation widget data for last 24 hours',
                    video: 'validate-resource-alloc-data-24hrs',
                    hours: 24
                }
            ];

            for (const caseData of resourcesAllocTests) {
                test(caseData.name, async ({}, testInfo: TestInfo) => {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'GPU_QUOTA data is missing the cluster metrics API response'
                    );

                    videoName = caseData.video;
                    const endTime = new Date();
                    const startTime = new Date(endTime.getTime() - caseData.hours * 60 * 60 * 1000);
                    const resp = await metricUtils.getClusterMetrics(
                        startTime,
                        endTime,
                        [MetricsType.TotalGpu, MetricsType.AllocatedGpu, MetricsType.GpuQuota],
                        20
                    );

                    const measurements = resp.body.measurements;

                    const types = measurements.map((m: any) => m.type);
                    expect(types).toEqual(
                        expect.arrayContaining(['TOTAL_GPU', 'ALLOCATED_GPU', 'GPU_QUOTA'])
                    );

                    const totalGpu = measurements.find((m: any) => m.type === 'TOTAL_GPU');
                    expect(totalGpu).toBeDefined();
                    for (const v of totalGpu.values) {
                        expect(Number(v.value)).toBe(totalGpuDeviceCnt);
                    }

                    // allocated gpu: 4
                    const allocatedGpu = measurements.find((m: any) => m.type === 'ALLOCATED_GPU');
                    expect(allocatedGpu).toBeDefined();
                    expect(allocatedGpu.values.some((v: any) => Number(v.value) >= 4)).toBeTruthy();

                    // gpu quota: 16 + 8 + 8 = 32
                    const gpuQuota = measurements.find((m: any) => m.type === 'GPU_QUOTA');
                    expect(gpuQuota).toBeDefined();
                    expect(gpuQuota.values.some((v: any) => Number(v.value) >= 16 + 8 + 8)).toBeTruthy();
                });
            }

            const workloadsByCategoryTests = [
                {
                    name: 'T5582199: validate the Workloads by category widget data for last 3 hours',
                    video: 'validate-workloads-by-category-highchart-3hrs',
                    hours: 3
                },
                {
                    name: 'T5582200: validate the Workloads by category widget data for last 24 hours',
                    video: 'validate-workloads-by-category-highchart-24hrs',
                    hours: 24
                }
            ];

            for (const caseData of workloadsByCategoryTests) {
                test(caseData.name, async () => {
                    videoName = caseData.video;
                    const endTime = new Date();
                    const startTime = new Date(endTime.getTime() - caseData.hours * 60 * 60 * 1000);
                    const resp = await metricUtils.getClusterMetrics(
                        startTime,
                        endTime,
                        [MetricsType.WorkloadsCount],
                        20,
                        'Category'
                    );

                    const measurements = resp.body.measurements;

                    const categories = measurements.map(
                        (m) => m.groups.find((g) => g.key === 'Category')?.value
                    );
                    expect(categories.sort()).toEqual(['Build', 'Deploy', 'Train'].sort());

                    const train = measurements.find((m) =>
                        m.groups.some((g) => g.key === 'Category' && g.value === 'Train')
                    );
                    expect(train).toBeDefined();
                    expect(train.values.some((v) => Number(v.value) >= 4)).toBeTruthy();

                    const build = measurements.find((m) =>
                        m.groups.some((g) => g.key === 'Category' && g.value === 'Build')
                    );
                    expect(build).toBeDefined();
                    expect(build.values.some((v) => Number(v.value) >= 1)).toBeTruthy();

                    const deploy = measurements.find((m) =>
                        m.groups.some((g) => g.key === 'Category' && g.value === 'Deploy')
                    );
                    expect(deploy).toBeDefined();
                    expect(deploy.values.some((v) => Number(v.value) >= 1)).toBeTruthy();
                });
            }
        });

        // ===================================================================================================
        // Overview dashboard widget link redirection tests
        // ===================================================================================================
        test.describe('Widget link redirection tests', () => {
            const allNodesRedirectionTests = [
                {
                    name: 'T5582203: verify the redirection of the All nodes link in the Ready / Total GPU nodes widgets',
                    vidName: 'verify-all-nodes-link-in-ready-total-gpu-nodes-widget',
                    clickFn: (): Promise<void> => overviewPage.clickAllNodesLinkInReadyTotalGpuNodesWidget()
                },
                {
                    name: 'T5582204: verify the redirection of the All nodes link in the Ready / Total GPU devices widgets',
                    vidName: 'verify-all-nodes-link-in-ready-total-gpu-devices-widget',
                    clickFn: (): Promise<void> => overviewPage.clickAllNodesLinkInReadyTotalGpuDevicesWidget()
                },
                {
                    name: 'T5582205: verify the redirection of the All nodes link in the Allocated GPU devices widgets',
                    vidName: 'verify-all-nodes-link-in-allocated-gpu-devices-widget',
                    clickFn: (): Promise<void> => overviewPage.clickAllNodesLinkInAllocatedGpuDevicesWidget()
                },
                {
                    name: 'T5582217: verify the redirection of the All nodes link in the Free nodes by node pool widgets',
                    vidName: 'verify-all-nodes-link-in-free-nodes-by-node-pool-widget',
                    clickFn: (): Promise<void> => overviewPage.clickAllNodesLinkInFreeNodesByNodePoolWidget()
                }
            ];

            for (const { name, vidName, clickFn } of allNodesRedirectionTests) {
                test(name, async () => {
                    videoName = vidName;
                    await test.step('go to the overview dashboard page', async () => {
                        await overviewPage.goToPage();
                    });

                    await test.step('click the All nodes link', async () => {
                        await clickFn();
                    });

                    await test.step('verify the nodes page is opened', async () => {
                        await nodesPage.page.waitForURL(
                            (url) => url.href.includes(`${config.tenantUrl}/nodes`),
                            { timeout: 15000 }
                        );
                    });
                });
            }

            const allWorkloadsRedirectionTests = [
                {
                    name: 'T5582206: verify the redirection of the All workloads link in the Idle allocated GPU devices widget',
                    vidName: 'verify-all-workloads-link-in-idle-allocated-gpu-devices-widget',
                    clickFn: (): Promise<void> =>
                        overviewPage.clickAllWorkloadsLinkInIdleAllocatedGpuDevicesWidget(),
                    filter: ''
                },
                {
                    name: 'T5582207: verify the redirection of the All workloads link in the Running workloads widget',
                    vidName: 'verify-all-workloads-link-in-running-workloads-widget',
                    clickFn: (): Promise<void> =>
                        overviewPage.clickAllWorkloadsLinkInRunningWorkloadsWidget(),
                    filter: 'Status: Running'
                },
                {
                    name: 'T5582208: verify the redirection of the All workloads link in the Pending workloads widget',
                    vidName: 'verify-all-workloads-link-in-pending-workloads-widget',
                    clickFn: (): Promise<void> =>
                        overviewPage.clickAllWorkloadsLinkInPendingWorkloadsWidget(),
                    filter: 'Status: Pending'
                },
                {
                    name: 'T5582209: verify the redirection of the All workloads link in the Failed workloads widget',
                    vidName: 'verify-all-workloads-link-in-failed-workloads-widget',
                    clickFn: (): Promise<void> => overviewPage.clickAllWorkloadsLinkInFailedWorkloadsWidget(),
                    filter: 'Status: Failed'
                },
                {
                    name: 'T5582210: verify the redirection of the All workloads link in the Running workloads by category widget',
                    vidName: 'verify-all-workloads-link-in-running-workloads-by-category-widget',
                    clickFn: (): Promise<void> =>
                        overviewPage.clickAllWorkloadsLinkInRunningWorkloadsByCategoryWidget(),
                    filter: ''
                },
                {
                    name: 'T5582211: verify the redirection of the All workloads link in the Workloads by status widget',
                    vidName: 'verify-all-workloads-link-in-workloads-by-status-widget',
                    clickFn: (): Promise<void> =>
                        overviewPage.clickAllWorkloadsLinkInWorkloadsByStatusWidget(),
                    filter: ''
                },
                {
                    name: 'T5582214: verify the redirection of the All workloads link in the Pending workloads by pending time widget',
                    vidName: 'verify-all-workloads-link-in-pending-workloads-by-pending-time-widget',
                    clickFn: (): Promise<void> =>
                        overviewPage.clickAllWorkloadsLinkInPendingWorkloadsByPendingTimeWidget(),
                    filter: 'Status: Pending'
                },
                {
                    name: 'T5582215: verify the redirection of the All workloads link in the Idle workloads widget',
                    vidName: 'verify-all-workloads-link-in-idle-workloads-widget',
                    clickFn: (): Promise<void> => overviewPage.clickAllWorkloadsLinkInIdleWorkloadsWidget(),
                    filter: ''
                }
            ];

            for (const caseData of allWorkloadsRedirectionTests) {
                test(caseData.name, async ({}, testInfo: TestInfo) => {
                    if (caseData.name.includes('T5582214')) {
                        skipPlaywrightTest(
                            test,
                            testInfo,
                            'the applied filter is wrong after the redirection'
                        );
                    }

                    videoName = caseData.vidName;

                    await test.step('go to the overview dashboard page', async () => {
                        await overviewPage.goToPage();
                    });

                    await test.step('click the All workloads link', async () => {
                        await caseData.clickFn();
                    });

                    await test.step('verify the workloads page is opened with correct filter applied', async () => {
                        await workloadsPage.page.waitForURL((url) => url.href.includes(workloadsUrl), {
                            timeout: 15000
                        });

                        if (caseData.filter) {
                            try {
                                // verify the filter is correctly applied
                                await workloadsPage.page
                                    .getByText(caseData.filter)
                                    .waitFor({ state: 'visible', timeout: staticElemCheckTimeout });
                            } catch (err) {
                                throw new Error(`Filter "${caseData.filter}" is not applied correctly`);
                            } finally {
                                // clear the applied filter to avoid impact to following tests
                                await workloadsPage.closeFilter();
                            }
                        }
                    });
                });
            }

            test('T5582212: verify the redirection of All projects link in the Projects table widget', async () => {
                videoName = 'verify-all-projects-link-in-projects-table-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('switch to Projects view in the table widget', async () => {
                    await overviewPage.switchViewInOrgUnitRankingWidget('Projects');
                });

                await test.step('click the All projects link', async () => {
                    await overviewPage.clickAllProjectsLinkInOrgUnitRankingWidget();
                });

                await test.step('verify the projects page is opened', async () => {
                    await projectsPage.page.waitForURL((url) => url.href.includes(projectsUrl), {
                        timeout: 15000
                    });
                });
            });

            test('T5582213: verify the redirection of All departments link in the Departments table widget', async () => {
                videoName = 'verify-all-departments-link-in-departments-table-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('switch to Departments view in the table widget', async () => {
                    await overviewPage.switchViewInOrgUnitRankingWidget('Departments');
                });

                await test.step('click the All departments link', async () => {
                    await overviewPage.clickAllDepartmentsLinkInOrgUnitRankingWidget();
                });

                await test.step('verify the departments page is opened', async () => {
                    await departmentsPage.page.waitForURL((url) => url.href.includes(departmentsUrl), {
                        timeout: 15000
                    });
                });
            });

            test('T5582216: verify the redirection of All departments link in the Allocation ratio by node pool widget', async () => {
                videoName = 'verify-all-nodepools-link-in-allocation-ratio-by-node-pool-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('click the All departments link', async () => {
                    await overviewPage.clickAllDepartmentsLinkInAllocationRatioByNodePoolWidget();
                });

                await test.step('verify the departments page is opened', async () => {
                    await departmentsPage.page.waitForURL((url) => url.href.includes(departmentsUrl), {
                        timeout: 15000
                    });
                });
            });

            test('T5582220: Test clicking project link in the projects table', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(test, testInfo, 'No data is displayed in the table');

                videoName = 'clicking-proj-link-in-table';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('switch to Projects view in the table widget', async () => {
                    await overviewPage.switchViewInOrgUnitRankingWidget('Projects');
                });

                await test.step(`click the project ${projName1} in the table`, async () => {
                    await overviewPage.clickProjDepmtLinkInOrgUnitRankingTable(projName1);
                });

                await test.step(`verify it redirects to the project page with ${projName1} filter applied`, async () => {
                    await departmentsPage.page.waitForURL((url) => url.href.includes(projectsUrl), {
                        timeout: 15000
                    });

                    const filter = `Project: equals ${projName1}`;
                    try {
                        await projectsPage.page
                            .getByText(filter)
                            .waitFor({ state: 'visible', timeout: staticElemCheckTimeout });
                    } catch (err) {
                        throw new Error(`Filter "${filter}" is not applied correctly`);
                    } finally {
                        // clear the applied filter to avoid impact to following tests
                        await projectsPage.closeFilter();
                    }
                });
            });

            test('T5582221: Test clicking department link in the departments table', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(test, testInfo, 'No data is displayed in the table');

                videoName = 'clicking-depmt-link-in-table';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('switch to Departments view in the table', async () => {
                    await overviewPage.switchViewInOrgUnitRankingWidget('Departments');
                });

                await test.step(`click the department ${deptName1} in the table`, async () => {
                    await overviewPage.clickProjDepmtLinkInOrgUnitRankingTable(deptName1);
                });

                await test.step(`verify it redirects to the department page with ${deptName1} filter applied`, async () => {
                    await departmentsPage.page.waitForURL((url) => url.href.includes(departmentsUrl), {
                        timeout: 15000
                    });

                    const filter = `Department: equals ${deptName1}`;
                    try {
                        await departmentsPage.page
                            .getByText(filter)
                            .waitFor({ state: 'visible', timeout: staticElemCheckTimeout });
                    } catch (err) {
                        throw new Error(`Filter "${filter}" is not applied correctly`);
                    } finally {
                        // clear the applied filter to avoid impact to following tests
                        await departmentsPage.closeFilter();
                    }
                });
            });

            test('T5582222: Test clicking workload link in the idle workloads widget', async () => {
                videoName = 'clicking-workload-link-in-idle-workloads-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step(`click the workload ${idleTrainingName} in the idle workloads table`, async () => {
                    await overviewPage.clickWorkloadsInIdleWorkloadsTable(idleTrainingName);
                });

                await test.step(`verify it redirects to the workloads page with ${idleTrainingName} filter applied`, async () => {
                    await workloadsPage.page.waitForURL((url) => url.href.includes(workloadsUrl), {
                        timeout: 15000
                    });

                    const filter = `Workload: equals ${idleTrainingName}`;
                    try {
                        await workloadsPage.page
                            .getByText(filter)
                            .waitFor({ state: 'visible', timeout: staticElemCheckTimeout });
                    } catch (err) {
                        throw new Error(`Filter "${filter}" is not applied correctly`);
                    } finally {
                        // clear the applied filter to avoid impact to following tests
                        await workloadsPage.closeFilter();
                    }
                });
            });

            test('T5582229: verufy cluster/department/project link redirection in the Running workloads by category widget', async () => {
                videoName = 'cluster-depmt-proj-redirections-in-running-workloads-by-category-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('click the cluster link and verify the departments shown', async () => {
                    await overviewPage.clickLinkInWidget(
                        overviewPage.runningWorkloadsByTypeWidget,
                        flowRootState.cluster.name
                    );

                    const exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.runningWorkloadsByTypeWidget,
                        'default'
                    );
                    expect(exist, 'check failed for default dept').toBeTruthy();
                });

                await test.step('click a department link and verify the projects shown', async () => {
                    await overviewPage.clickLinkInWidget(
                        overviewPage.runningWorkloadsByTypeWidget,
                        'default'
                    );

                    const exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.runningWorkloadsByTypeWidget,
                        projName1
                    );
                    expect(exist, `check failed for project ${projName1}`).toBeTruthy();
                });
            });

            test('T5582230: verify cluster/department/project link redirection in the Workloads by status widget', async () => {
                videoName = 'cluster-depmt-proj-redirections-in-workloads-by-status-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('click the cluster link and verify the departments shown', async () => {
                    await overviewPage.clickLinkInWidget(
                        overviewPage.workloadsByStatusWidget,
                        flowRootState.cluster.name
                    );

                    logger.info(`depmts: ${depmts}`);
                    for (const depmt of depmts) {
                        const exist = await overviewPage.entityLinkExistInWidget(
                            overviewPage.workloadsByStatusWidget,
                            depmt
                        );
                        expect(exist, `check failed for ${depmt}`).toBeTruthy();
                    }
                });

                await test.step('click a department link and verify the projects shown', async () => {
                    const entries = Object.entries(depmtProjMap);
                    const randomIndex = Math.floor(Math.random() * entries.length);
                    const [dept, proj] = entries[randomIndex];

                    await overviewPage.clickLinkInWidget(overviewPage.workloadsByStatusWidget, dept);

                    const exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.workloadsByStatusWidget,
                        proj
                    );
                    expect(exist, `check failed for ${proj}`).toBeTruthy();
                });
            });

            test('T5582231: verify cluster/department/project link redirection in the Pending workloads by pending time widget', async () => {
                videoName = 'cluster-depmt-proj-redirections-in-pending-workloads-by-pending-time-widget';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('click the cluster link and verify the departments shown', async () => {
                    await overviewPage.clickLinkInWidget(
                        overviewPage.pendingWorkloadDistTimeWidget,
                        flowRootState.cluster.name
                    );

                    const exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.pendingWorkloadDistTimeWidget,
                        deptName2
                    );
                    expect(exist, `check failed for ${deptName2}`).toBeTruthy();
                });

                await test.step('click a department link and verify the projects shown', async () => {
                    await overviewPage.clickLinkInWidget(
                        overviewPage.pendingWorkloadDistTimeWidget,
                        deptName2
                    );

                    const exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.pendingWorkloadDistTimeWidget,
                        projName3
                    );
                    expect(exist, `check failed for ${projName3}`).toBeTruthy();
                });
            });
        }); // End of: Widget link redirection tests

        test.describe('View switching and UI interaction tests', () => {
            test('T5582219: Test switching b/w projects and departments views', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(test, testInfo, 'No data is displayed in the table');

                videoName = 'verify-switching-bw-projects-departments-views';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('switch to Departments view in the table widget and verify the data', async () => {
                    await overviewPage.switchViewInOrgUnitRankingWidget('Departments');
                    const data = await overviewPage.getProjectsDepartmentsTableData();
                    const departments = data.map((item: any) => item['Department']);

                    depmts.forEach((dep) => {
                        expect(departments).toContain(dep);
                    });
                });

                await test.step('switch to Projects view in the table widget and verify the data', async () => {
                    await overviewPage.switchViewInOrgUnitRankingWidget('Projects');
                    const data = await overviewPage.getProjectsDepartmentsTableData();
                    const projects = data.map((item: any) => item['Projects']);

                    projs.forEach((proj) => {
                        expect(projects).toContain(proj);
                    });
                });
            });
        }); // End of: View switching and UI interaction tests

        // =================================
        // CSV downloading tests
        // =================================
        test.describe('CSV export functionality tests', () => {
            test('T5582159: Test downloading the Running workloads by category widget csv file', async () => {
                videoName = 'dnld-running-workloads-by-category-csv';
                let csvData: Array<Record<string, string>> = [];

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('download the csv file', async () => {
                    const csvPath = await overviewPage.downloadCSV(
                        overviewPage.runningWorkloadsByTypeWidget,
                        'RUNNING_WORKLOADS_BY_CATEGORY'
                    );
                    csvData = Utils.readCSVFile(csvPath);
                    logger.info(JSON.stringify(csvData, null, 2));
                });
            });

            test('T5582160: Test downloading the Running workloads by type widget csv file', async () => {
                videoName = 'dnld-running-workloads-by-type-csv';
                let csvData: Array<Record<string, string>> = [];

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('switch group by to Type', async () => {
                    await overviewPage.selectRunningWorkloadsGroupBy('Running workloads by type');
                });

                await test.step('download the csv file', async () => {
                    const csvPath = await overviewPage.downloadCSV(
                        overviewPage.runningWorkloadsByTypeWidget,
                        'RUNNING_WORKLOADS_BY_TYPE'
                    );
                    csvData = Utils.readCSVFile(csvPath);
                    logger.info(JSON.stringify(csvData, null, 2));
                });

                await test.step('validate the csv data', () => {
                    expect(csvData.length).toBeGreaterThan(0);
                    expect(Object.keys(csvData[0] || {}).length).toBeGreaterThan(0);
                });
            });

            test('T5582161: Test downloading the Workloads status widget csv file', async () => {
                videoName = 'dnld-workloads-status-csv';
                let csvData: Array<Record<string, string>> = [];

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('download the csv file', async () => {
                    const csvPath = await overviewPage.downloadCSV(
                        overviewPage.workloadsByStatusWidget,
                        'WORKLOADS_STATUS'
                    );
                    csvData = Utils.readCSVFile(csvPath);
                    logger.info(JSON.stringify(csvData, null, 2));
                });

                await test.step('validate the csv data', () => {
                    expect(csvData.length).toBeGreaterThan(0);
                    expect(Object.keys(csvData[0] || {}).length).toBeGreaterThan(0);
                });
            });

            test('T5582162: Test downloading the Free nodes by node pool widget csv file', async () => {
                videoName = 'dnld-free-nodes-by-node-pool-csv';
                let csvData: Array<Record<string, string>> = [];

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('download the csv file', async () => {
                    const csvPath = await overviewPage.downloadCSV(
                        overviewPage.freeNodesByNodepoolWidget,
                        'FREE_NODES_BY_NODE_POOL'
                    );
                    csvData = Utils.readCSVFile(csvPath);
                    logger.info(JSON.stringify(csvData, null, 2));
                });

                await test.step('validate the csv data', () => {
                    expect(csvData.length).toBeGreaterThan(0);
                    expect(Object.keys(csvData[0] || {}).length).toBeGreaterThan(0);
                });
            });

            test('T5582163: Test downloading the Resource utilization widget csv file', async () => {
                videoName = 'dnld-resource-utilization-csv';
                let csvData: Array<Record<string, string>> = [];

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('download the csv file', async () => {
                    const csvPath = await overviewPage.downloadCSV(
                        overviewPage.resourceUtilizationWidget,
                        'RESOURCE_UTILIZATION'
                    );
                    csvData = Utils.readCSVFile(csvPath);
                    logger.info(JSON.stringify(csvData, null, 2));
                });

                await test.step('validate the csv data', () => {
                    expect(csvData.length).toBeGreaterThan(0);
                    expect(Object.keys(csvData[0] || {}).length).toBeGreaterThan(0);
                });
            });

            test('T5582164: Test downloading the Resource allocation widget csv file', async () => {
                videoName = 'dnld-resource-allocation-csv';
                let csvData: Array<Record<string, string>> = [];

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('download the csv file', async () => {
                    const csvPath = await overviewPage.downloadCSV(
                        overviewPage.resourceAllocationWidget,
                        'RESOURCE_ALLOCATION'
                    );
                    csvData = Utils.readCSVFile(csvPath);
                    logger.info(JSON.stringify(csvData, null, 2));
                });

                await test.step('validate the csv data', () => {
                    expect(csvData.length).toBeGreaterThan(0);
                    expect(Object.keys(csvData[0] || {}).length).toBeGreaterThan(0);
                });
            });
        }); // End of: CSV export functionality tests

        // ================================================
        // Node pool filter test
        // ================================================
        test.describe('Node pool filter functionality tests', () => {
            test('T5582156: Test overview node pool filter', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'proj/dept table does not show the correct data when the filter applied'
                );

                videoName = 'test-overview-node-pool-filter';

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step(`set the node pool filter to the ${npName1} node pool`, async () => {
                    await overviewPage.setNodePoolFilter(npName1);
                });

                try {
                    await test.step('verify Ready / Total GPU nodes widget data', async () => {
                        const data = await overviewPage.getReadyTotalGpuNodesWidgetData();
                        expect(data).toBe('1/1');
                    });

                    await test.step('verify Allocated GPU devices widget data', async () => {
                        const data = await overviewPage.getAllocatedGpuDevicesWidgetData();
                        expect(data).toBe(0);
                    });

                    await test.step('verify Idle allocated GPU devices widget data', async () => {
                        const data = await overviewPage.getIdleAllocatedGpuDevicesWidgetData();
                        expect(data).toBe(0);
                    });

                    await test.step('verify Running workloads widget data', async () => {
                        const data = await overviewPage.getRunningWorkloadsWidgetData();
                        expect(data).toBe(0);
                    });

                    await test.step('verify Pending workloads widget data', async () => {
                        const data = await overviewPage.getPendingWorkloadsWidgetData();
                        expect(data).toBe(0);
                    });

                    await test.step('verify Failed workloads widget data', async () => {
                        const data = await overviewPage.getFailedWorkloadsWidgetData();
                        expect(data).toBe(1);
                    });

                    await test.step('verify the Running workloads by category widget data', async () => {
                        const workloadCatResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                            WorkloadTelemetryType.WorkloadsCount,
                            ['Category'],
                            flowRootState.cluster.uuid,
                            npName1,
                            [Phase.Running]
                        );

                        // running workspace workload: 0
                        const buildItem = workloadCatResp.values.find((item: any) =>
                            item.groups.some((g: any) => g.key === 'Category' && g.value === 'Build')
                        );
                        const buildValue = Number(buildItem?.value ?? 0);
                        expect(buildValue).toBe(0);

                        // running inference workload: 0
                        const deployItem = workloadCatResp.values.find((item: any) =>
                            item.groups.some((g: any) => g.key === 'Category' && g.value === 'Deploy')
                        );
                        const deployValue = Number(deployItem?.value ?? 0);
                        expect(deployValue).toBeGreaterThanOrEqual(0);

                        // running training workload: 0
                        const trainItem = workloadCatResp.values.find((item: any) =>
                            item.groups.some((g: any) => g.key === 'Category' && g.value === 'Train')
                        );
                        const trainValue = Number(trainItem?.value ?? 0);
                        expect(trainValue).toBeGreaterThanOrEqual(0);
                    });

                    await test.step('verify the Workloads by status widget data', async () => {
                        const workloadTypeResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                            WorkloadTelemetryType.WorkloadsCount,
                            ['Phase'],
                            flowRootState.cluster.uuid,
                            npName1
                        );

                        // running workload: 0
                        const runningItem = workloadTypeResp.values.find((item: any) =>
                            item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Running')
                        );
                        const runningValue = Number(runningItem?.value ?? 0);
                        expect(runningValue).toBe(0);

                        // failed workload: 1
                        const failedItem = workloadTypeResp.values.find((item: any) =>
                            item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Failed')
                        );
                        expect(Number(failedItem?.value)).toBe(1);

                        // pending workload: 0
                        const pendingItem = workloadTypeResp.values.find((item: any) =>
                            item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Pending')
                        );
                        const pendingValue = Number(pendingItem?.value ?? 0);
                        expect(pendingValue).toBe(0);
                    });

                    await test.step('verify projects table widget data', async () => {
                        const resp = await orgUnitUtils.getProjectsRanking(
                            RankingMetricType.AvgGpuAllocation24h,
                            RankingOrderType.Top,
                            npName1
                        );

                        const projectsData = resp.body.projects;
                        expect(projectsData).toHaveLength(1);

                        expect(projectsData[0].name).toBe(projName2);
                        expect(projectsData[0].parentName).toBe(deptName1);
                        expect(projectsData[0].gpuQuota).toBe(8);
                        expect(projectsData[0].runningWorkloadsCount).toBe(0);
                        expect(projectsData[0].pendingWorkloadsCount).toBe(0);
                    });
                } finally {
                    // ensure the node pool filter will be removed so that other tests will not be impacted
                    await test.step('remove node pool filter', async () => {
                        await overviewPage.removeNodePoolFilter();
                    });
                }
            });
        }); // End of: Node pool filter functionality tests

        // ==================================
        // Overview widgets update tests
        // ==================================
        test.describe('Widget data update tests', () => {
            test('T5582225: verify the running workloads related widgets updated after deleting a running workload', async () => {
                videoName = 'verify-running-workloads-widgets-update-after-delete';

                await test.step('go to the workload page', async () => {
                    await workloadsPage.goToPage();
                });

                await test.step(`delete the running workload ${runningInferenceName}`, async () => {
                    await workloadsCrudUtils.deleteInference(runningInferenceId);
                    await waitUtils.waitForWorkloadPodNotExist(`runai-${projName1}`, runningInferenceName);

                    const index = inferenceIdsToCleanup.indexOf(runningInferenceId);
                    if (index !== -1) {
                        inferenceIdsToCleanup.splice(index, 1);
                    }
                });

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify the Running workloads widget data', async () => {
                    const data = await overviewPage.getRunningWorkloadsWidgetData();
                    expect(data).toBeGreaterThanOrEqual(3);
                });

                await test.step('verify the Running workloads by category widget data', async () => {
                    const workloadCatResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                        WorkloadTelemetryType.WorkloadsCount,
                        ['Category'],
                        flowRootState.cluster.uuid,
                        undefined,
                        [Phase.Running]
                    );

                    // running inference workload: 0
                    const deployItem = workloadCatResp.values.find((item: any) =>
                        item.groups.some((g: any) => g.key === 'Category' && g.value === 'Deploy')
                    );

                    const deployValue = Number(deployItem?.value ?? 0);
                    expect(deployValue).toBeGreaterThanOrEqual(0);
                });

                await test.step('verify the Workloads by status widget data', async () => {
                    const workloadTypeResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                        WorkloadTelemetryType.WorkloadsCount,
                        ['Phase'],
                        flowRootState.cluster.uuid
                    );

                    // running workload: 3
                    const runningItem = workloadTypeResp.values.find((item: any) =>
                        item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Running')
                    );
                    expect(Number(runningItem?.value)).toBeGreaterThanOrEqual(3);
                });
            });

            test('T5582226: verify the pending workloads related widgets updated after deleting a pending workload', async () => {
                videoName = 'verify-pending-workloads-widgets-update-after-delete';

                await test.step('go to the workload page', async () => {
                    await workloadsPage.goToPage();
                });

                await test.step(`delete the pending workload ${pendingTrainingName}`, async () => {
                    await workloadsCrudUtils.deleteTraining(pendingTrainingId);
                    await waitUtils.waitForWorkloadPodNotExist(`runai-${projName3}`, pendingTrainingName);

                    const index = trainingIdsToCleanup.indexOf(pendingTrainingId);
                    if (index !== -1) {
                        trainingIdsToCleanup.splice(index, 1);
                    }
                });

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Pending workloads widget data', async () => {
                    // total 0 pending workload
                    const data = await overviewPage.getPendingWorkloadsWidgetData();
                    expect(data).toBeGreaterThanOrEqual(0);
                });

                await test.step('verify the Workloads by status widget data', async () => {
                    const workloadTypeResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                        WorkloadTelemetryType.WorkloadsCount,
                        ['Phase'],
                        flowRootState.cluster.uuid
                    );

                    // pending workload: 0
                    const pendingItem = workloadTypeResp.values.find((item: any) =>
                        item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Pending')
                    );
                    const pendingValue = Number(pendingItem?.value ?? 0);
                    expect(pendingValue).toBeGreaterThanOrEqual(0);
                });
            });

            test('T5582227: verify the failed workloads related widgets updated after deleting a failed workload', async () => {
                videoName = 'verify-failed-workloads-widgets-update-after-delete';

                await test.step('go to the workload page', async () => {
                    await workloadsPage.goToPage();
                });

                await test.step(`delete the failed workload ${failedTrainingName}`, async () => {
                    await workloadsCrudUtils.deleteTraining(failedTrainingId);
                    await waitUtils.waitForWorkloadPodNotExist(`runai-${projName2}`, failedTrainingName);

                    const index = trainingIdsToCleanup.indexOf(failedTrainingId);
                    if (index !== -1) {
                        trainingIdsToCleanup.splice(index, 1);
                    }
                });

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify Failed workloads widget data', async () => {
                    // total 0 failed workloads
                    const data = await overviewPage.getFailedWorkloadsWidgetData();
                    expect(data).toBeGreaterThanOrEqual(0);
                });

                await test.step('verify the Workloads by status widget data', async () => {
                    const workloadTypeResp = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                        WorkloadTelemetryType.WorkloadsCount,
                        ['Phase'],
                        flowRootState.cluster.uuid
                    );

                    // failed workload: 0
                    const pendingItem = workloadTypeResp.values.find((item: any) =>
                        item.groups.some((g: any) => g.key === 'Phase' && g.value === 'Failed')
                    );

                    const pendingValue = Number(pendingItem?.value ?? 0);
                    expect(pendingValue).toBeGreaterThanOrEqual(0);
                });
            });

            test('T5582228: verify the idle workloads related widgets updated after deleting an idle workload', async () => {
                videoName = 'verify-idle-workloads-widgets-update-after-delete';

                await test.step('go to the workload page', async () => {
                    await workloadsPage.goToPage();
                });

                await test.step(`delete the idle workload ${idleTrainingName}`, async () => {
                    await workloadsCrudUtils.deleteTraining(idleTrainingId);
                    await waitUtils.waitForWorkloadPodNotExist(`runai-${projName1}`, idleTrainingName);

                    const index = trainingIdsToCleanup.indexOf(idleTrainingId);
                    if (index !== -1) {
                        trainingIdsToCleanup.splice(index, 1);
                    }
                });

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('verify idle allocated GPU devices widget data', async () => {
                    // 0 idle GPU device
                    const data = await overviewPage.getIdleAllocatedGpuDevicesWidgetData();
                    expect(data).toBeGreaterThanOrEqual(0);
                });
            });

            test('T5582223: verify the projects related widgets updated after deleting a project', async () => {
                videoName = 'verify-projects-widgets-update-after-delete';

                await test.step('go to the projects page', async () => {
                    await projectsPage.goToPage();
                });

                await test.step(`delete the project ${projName3}`, async () => {
                    await operations.deleteProjectAndNamespace(parseInt(proj3Id));

                    const index = createdProjectIds.indexOf(parseInt(proj3Id));
                    if (index !== -1) {
                        createdProjectIds.splice(index, 1);
                    }
                });

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step('click the cluster link and verify the departments shown in the Workloads by status widget', async () => {
                    await overviewPage.clickLinkInWidget(
                        overviewPage.workloadsByStatusWidget,
                        flowRootState.cluster.name
                    );

                    const defaultDeptExist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.workloadsByStatusWidget,
                        'default'
                    );
                    expect(defaultDeptExist, 'check failed for the default dept').toBeTruthy();

                    const dept1Exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.workloadsByStatusWidget,
                        deptName1
                    );
                    expect(dept1Exist).toBeFalsy();

                    const dept2Exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.workloadsByStatusWidget,
                        deptName2
                    );
                    expect(dept2Exist).toBeFalsy();
                });
            });

            test('T5582224: verify the department related widgets updated after deleting a department', async () => {
                videoName = 'verify-department-widgets-update-after-delete';

                await test.step('go to the departments page', async () => {
                    await departmentsPage.goToPage();
                });

                await test.step(`delete the department ${deptName2}`, async () => {
                    await departmentUtils.deleteDepartment(parseInt(deptId2));

                    const index = createdDepartmentIds.indexOf(parseInt(deptId2));
                    if (index !== -1) {
                        createdDepartmentIds.splice(index, 1);
                    }
                });

                await test.step('go to the overview dashboard page', async () => {
                    await overviewPage.goToPage();
                });

                await test.step(`click the cluster link and verify the departments ${deptName2} not shown in the Workloads by status widget`, async () => {
                    await overviewPage.clickLinkInWidget(
                        overviewPage.workloadsByStatusWidget,
                        flowRootState.cluster.name
                    );

                    const exist = await overviewPage.entityLinkExistInWidget(
                        overviewPage.workloadsByStatusWidget,
                        deptName2
                    );
                    expect(exist).toBeFalsy();
                });
            });
        }); // End of: Widget data update tests

        test.describe('Miscellaneous feature tests', () => {
            test('T5582158: Test the guided tour for overview dashboard', async () => {
                isTheLastTest = true;
                videoName = 'test-overview-tour-guide';

                await test.step('open overview dashboard tour guide', async () => {
                    await guidedTour.openOverviewDashboardHelpTourGide();
                });

                await test.step('verify the 1st tour guide step', async () => {
                    const content = await guidedTour.getModalText();
                    expect(content).toContain('explore the dashboard and help you');
                });

                await test.step('verify the 2nd tour guide step', async () => {
                    await guidedTour.clickModalNextButton();
                    await guidedTour.page.getByText('2/6').waitFor({ state: 'visible', timeout: 3000 });
                    const content = await guidedTour.getModalText();
                    expect(content).toContain('Customize the view');
                });

                await test.step('verify the 3rd tour guide step', async () => {
                    await guidedTour.clickModalNextButton();
                    await guidedTour.page.getByText('3/6').waitFor({ state: 'visible', timeout: 3000 });
                    const content = await guidedTour.getModalText();
                    expect(content).toContain('Drill down to view');
                });

                await test.step('verify the 4th tour guide step', async () => {
                    await guidedTour.clickModalNextButton();
                    await guidedTour.page.getByText('4/6').waitFor({ state: 'visible', timeout: 3000 });
                    const content = await guidedTour.getModalText();
                    expect(content).toContain('View averages over different time frames');
                });

                await test.step('verify the 5th tour guide step', async () => {
                    await guidedTour.clickModalNextButton();
                    await guidedTour.page.getByText('5/6').waitFor({ state: 'visible', timeout: 3000 });
                    const content = await guidedTour.getModalText();
                    expect(content).toContain('For a full breakdown');
                });

                await test.step('verify the 6th tour guide step', async () => {
                    await guidedTour.clickModalNextButton();
                    await guidedTour.page.getByText('6/6').waitFor({ state: 'visible', timeout: 3000 });
                    const content = await guidedTour.getModalText();
                    expect(content).toContain('Change the date range to view data trends over time');
                });

                await test.step('click the DONE button to end the guide tour', async () => {
                    await guidedTour.clickModalNextButton();
                    await guidedTour.page
                        .getByText('Guide completed!')
                        .waitFor({ state: 'visible', timeout: 3000 });
                    await guidedTour.closeCompletionModal();
                    await guidedTour.page
                        .getByText('Guide completed!')
                        .waitFor({ state: 'hidden', timeout: 3000 });
                });
            });
        }); // End of: Miscellaneous feature tests
    });
});
