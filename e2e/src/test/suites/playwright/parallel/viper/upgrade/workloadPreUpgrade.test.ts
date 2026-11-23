/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import { test } from '@playwright/test';
import {
    apiUtils,
    authUtils,
    BaseWorkload,
    config,
    crds,
    externalWorkloadUtils,
    flowRootState,
    logger,
    objectGenerator,
    PodStatus,
    settings,
    tags,
    Utils,
    usersManager,
    versionUtils,
    waitUtils,
    workloadsManagerUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { createProject } from '../../../../../../modules/helpers/models/projects';
import { PyTorchJob } from '../../../../../../modules/helpers/models/externalWorkload';
import { JobUtils } from '../../../../../../modules/helpers/utils/jobUtils';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    shouldRunNativeWorkloadCrdTests,
    shouldRunTests,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    isKubeflowInstalled,
    isLwsInstalled
} from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { SoftAssert } from '../../../../../../modules/helpers/SoftAssert';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadUtils';
import {
    captureV2WorkloadSeed,
    completedWlTestName,
    failedWlTestName,
    initializingWlTestName,
    pendingWlTestName,
    podsId,
    podsYaml,
    runningWlTestName,
    suspendedWlTestName,
    v2WorkloadsKey,
    V2WorkloadSeed,
    verifyV2WorkloadSeedHealthy,
    workloadCategory,
    workloadGen
} from './workloadUpgradeCommon';
import { resetNodePoolCrdDetection } from '../../../../../../modules/helpers/utils/NodePoolCrdFields';

// Inlined to avoid pulling another playwright test file's `test.describe()` registrations.
interface frameworkParams {
    type: cliV2WorkloadType;
    wlType: WorkloadType;
}

const paramsAllFrameworks: frameworkParams[] = [
    { type: cliV2WorkloadType.tensorFlow, wlType: WorkloadType.Tensor },
    { type: cliV2WorkloadType.xgBoost, wlType: WorkloadType.Xgboost },
    { type: cliV2WorkloadType.mpi, wlType: WorkloadType.Mpi },
    { type: cliV2WorkloadType.pyTorch, wlType: WorkloadType.Pytorch }
];

const UPGRADE_ARTIFACT_PATH = '/tmp/upgradeArtifact.txt';

shouldRunTests(settings.isUpgradesTests)(() => {
    test.describe(`generate pre upgrade workload data ${tags.runOn.workloadPreUpgrade} ${tags.teams.viper}`, () => {
        const imagePath = config.quickStartImage.ubuntu;

        const jobKindTraining = 'training';
        const jobKindInteractive = 'interactive';
        const jobKindPytorch = 'pytorch';
        const jobKindTensor = 'tensor';
        const JobKindMpi = 'mpi';
        const jobKindXgBoost = 'xgboost';

        const submitFlag = {
            [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu
        } as Partial<typeof SubmitFlagsMerged>;

        const projectOptions = { maxAllowedGpus: -1, deservedGpus: 3 };

        let projectName: string;
        let namespace: string;
        let project: Project;
        const upgradeArtifact: Record<string, any> = {};
        let workloads: { [kind: string]: BaseWorkload } = {};

        async function setWorkloadPodsData(workload: BaseWorkload): Promise<BaseWorkload> {
            workload['id'] = await workloadsUtilsV2.getWorkloadIdByName(workload.name, project.id);
            const workloadPods = (await workloadsUtilsV2.getWorkloadPods(workload['id'])).body.pods;
            workload[podsId] = [];
            workload[podsYaml] = [];
            for (const pod in workloadPods) {
                workload[podsId].push(workloadPods[pod].id);
                const podYaml = await kubeCli.getPodYaml(workloadPods[pod].name, namespace);
                workload[podsYaml].push(podYaml);
            }
            return workload;
        }

        async function setWorkloadsData(
            workloadsToProcess: { [p: string]: BaseWorkload },
            testName: string
        ): Promise<void> {
            for (const workload in workloadsToProcess) {
                workloadsToProcess[workload] = await setWorkloadPodsData(workloadsToProcess[workload]);
                workloadsToProcess[workload]['crd'] = workloadsToProcess[workload].getCrd();
                workloadsToProcess[workload]['yaml'] = await kubeCli.getWorkload(
                    workloadsToProcess[workload].name,
                    namespace,
                    workloadsToProcess[workload]['crd']
                );
                workloadsToProcess[workload][workloadGen] =
                    workloadsToProcess[workload]['yaml'].metadata.generation;
                upgradeArtifact[workloadCategory][testName].push(workloadsToProcess[workload]);
            }
        }

        test.beforeAll(async () => {
            const adminTokens = await authUtils.authenticateKeycloakAndGetTokenOld(
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword
            );
            flowRootState.cluster = await apiUtils.getClusterByPrefix(
                config.clusterNamePrefix,
                adminTokens.idToken
            );
            await usersManager.initLocalUsers(adminTokens.accessToken);
            await kubeCli.initKubeContext(config.kubeconfigClusterName, false);
            kubeCli.useAdminContext();
            await waitUtils.waitForCondition(
                async () => (await kubeCli.getNumOfGpus()) > 0,
                'wait for gpus to be available in the cluster',
                true,
                300
            );
            flowRootState.nodesResources = await kubeCli.getNodesResources();
            flowRootState.numOfGpus = flowRootState.nodesResources.reduce((sum, node) => sum + node.gpu, 0);
            crds.initCrds(flowRootState.cluster.version);
            resetNodePoolCrdDetection();

            runaiCliV2.downloadAndInstall();

            project = await createProject('upgrade', projectOptions);
            projectName = project.name;
            namespace = Utils.namespaceFromProjectName(projectName);
            project.namespace = namespace;
            upgradeArtifact['project'] = project;
            upgradeArtifact[workloadCategory] = {};
            runaiCliV2.projectSet(projectName);
            logger.info(`pre-upgrade project created: ${projectName} (${namespace})`);
        });

        test.afterAll(() => {
            fs.writeFileSync(UPGRADE_ARTIFACT_PATH, JSON.stringify(upgradeArtifact));
            logger.info(`upgrade artifact written to ${UPGRADE_ARTIFACT_PATH}`);
        });

        shouldRunNativeWorkloadCrdTests(() =>
        test(`1) generate 1 ${runningWlTestName}`, async () => {
            workloads = await JobUtils.createAllTypeJobs(
                projectOptions,
                namespace,
                imagePath,
                `${jobKindInteractive}-run`,
                `${jobKindTraining}-run`,
                `${JobKindMpi}-run`,
                `${jobKindPytorch}-run`,
                `${jobKindTensor}-run`,
                `${jobKindXgBoost}-run`,
                'sleep infinity'
            );
            upgradeArtifact[workloadCategory][runningWlTestName] = [];
            await setWorkloadsData(workloads, runningWlTestName);
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test(`2) generate ${suspendedWlTestName}`, async () => {
            workloads = await JobUtils.createAllTypeJobs(
                projectOptions,
                namespace,
                imagePath,
                `${jobKindInteractive}-susp`,
                `${jobKindTraining}-susp`,
                '',
                '',
                '',
                '',
                'sleep infinity'
            );
            upgradeArtifact[workloadCategory][suspendedWlTestName] = [];
            for (const workload in workloads) {
                workloads[workload].active = false;
                await kubeCli.updateWorkload(workloads[workload]);
            }
            await setWorkloadsData(workloads, suspendedWlTestName);
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test(`3) generate ${completedWlTestName}`, async () => {
            const wlSuffix = objectGenerator.generateNameByPrefix('-comp');
            workloads = await JobUtils.createAllTypeJobs(
                projectOptions,
                namespace,
                imagePath,
                `${jobKindInteractive}`,
                `${jobKindTraining}`,
                '',
                `${jobKindPytorch}`,
                `${jobKindTensor}`,
                `${jobKindXgBoost}`,
                '',
                null,
                '0',
                wlSuffix
            );
            await waitUtils.waitForBackendWorkloadPhase(
                `${jobKindInteractive}${wlSuffix}`,
                project.id,
                Phase.Completed,
                180
            );
            await waitUtils.waitForBackendWorkloadPhase(
                `${jobKindTraining}${wlSuffix}`,
                project.id,
                Phase.Completed,
                180
            );
            await waitUtils.waitForBackendWorkloadPhase(
                `${jobKindPytorch}${wlSuffix}`,
                project.id,
                Phase.Completed,
                180
            );
            await waitUtils.waitForBackendWorkloadPhase(
                `${jobKindTensor}${wlSuffix}`,
                project.id,
                Phase.Completed,
                180
            );
            await waitUtils.waitForBackendWorkloadPhase(
                `${jobKindXgBoost}${wlSuffix}`,
                project.id,
                Phase.Completed,
                180
            );
            upgradeArtifact[workloadCategory][completedWlTestName] = [];

            await setWorkloadsData(workloads, completedWlTestName);
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test(`4) generate ${initializingWlTestName}`, async () => {
            workloads = await JobUtils.createAllTypeJobs(
                projectOptions,
                namespace,
                'invalid-image',
                `${jobKindInteractive}-init`,
                `${jobKindTraining}-init`,
                `${JobKindMpi}-init`,
                `${jobKindPytorch}-init`,
                `${jobKindTensor}-init`,
                `${jobKindXgBoost}-init`,
                '',
                PodStatus.PENDING
            );
            upgradeArtifact[workloadCategory][initializingWlTestName] = [];

            await setWorkloadsData(workloads, initializingWlTestName);
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test(`5) generate ${pendingWlTestName}`, async () => {
            workloads = await JobUtils.createAllTypeJobs(
                projectOptions,
                namespace,
                imagePath,
                `${jobKindInteractive}-pend`,
                `${jobKindTraining}-pend`,
                `${JobKindMpi}-run`,
                `${jobKindPytorch}-pend`,
                `${jobKindTensor}-pend`,
                `${jobKindXgBoost}-pend`,
                '',
                PodStatus.PENDING,
                '100'
            );
            upgradeArtifact[workloadCategory][pendingWlTestName] = [];

            await setWorkloadsData(workloads, pendingWlTestName);
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test(`6) generate ${failedWlTestName}2`, async () => {
            test.setTimeout(15 * 60 * 1000);
            upgradeArtifact[workloadCategory][failedWlTestName] = [];

            const failPod = 'exit 1';
            const failCommand = '-- cat nonExistCommand';
            submitFlag[SubmitFlagsMerged.backoffLimit] = 1;
            submitFlag[SubmitFlagsMerged.command] = failCommand;

            workloads = await JobUtils.createAllTypeJobs(
                projectOptions,
                namespace,
                imagePath,
                `${jobKindInteractive}-fail`,
                `${jobKindTraining}-fail`,
                '',
                '',
                '',
                '',
                failPod,
                PodStatus.FAILED,
                '0'
            );

            for (const params of paramsAllFrameworks) {
                const jobName = objectGenerator.generateNameByPrefix(`${params.type}-fail`);
                runaiCliV2.submit(submitFlag, jobName, params.type);
                await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, Phase.Failed);
                const workload = await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                    jobName,
                    project.name
                );
                const workloadCrd = workloadUtils.getDistributedWorkload(
                    params.wlType,
                    project.name,
                    jobName,
                    imagePath
                );
                workload['crd'] = workloadCrd.getCrd();
                workload['yaml'] = await kubeCli.getWorkload(workload.name, namespace, workload['crd']);
                workload[workloadGen] = workload['yaml'].metadata.generation;
                upgradeArtifact[workloadCategory][failedWlTestName].push(workload);
            }

            await setWorkloadsData(workloads, failedWlTestName);
            })
        );

        // ---------------- v2 workloads pre-upgrade ----------------
        // Two submission entry points, both kept in Running through the upgrade:
        //   Path A: workloads-v2 API (raw manifest -> WorkloadRunner CR)
        //   Path B: kubectl apply of a framework CRD (PyTorchJob -> EWI wrapper)
        // Same upgradeArtifact, separate top-level key so v1 keys are untouched.
        test('7) generate v2 Path A: LeaderWorkerSet via workloads-v2 API', async () => {
            if (!versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_24)) {
                logger.warn(
                    `cluster version ${flowRootState.cluster.version} < ${SupportedVersions.V2_24}, skipping LWS via workloads-v2 API`
                );
                test.skip();
            }
            if (!(await isLwsInstalled())) {
                logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                test.skip();
            }

            if (!upgradeArtifact[v2WorkloadsKey]) {
                upgradeArtifact[v2WorkloadsKey] = [] as V2WorkloadSeed[];
            }

            const workloadName = objectGenerator.generateNameByPrefix('lws-wr');
            const lwsManifest = {
                apiVersion: 'leaderworkerset.x-k8s.io/v1',
                kind: 'LeaderWorkerSet',
                metadata: { name: workloadName },
                spec: {
                    replicas: 1,
                    leaderWorkerTemplate: {
                        size: 2,
                        leaderTemplate: {
                            spec: {
                                containers: [
                                    {
                                        name: 'leader',
                                        image: config.quickStartImage.ubuntu,
                                        command: ['sleep', 'infinity']
                                    }
                                ]
                            }
                        },
                        workerTemplate: {
                            spec: {
                                containers: [
                                    {
                                        name: 'worker',
                                        image: config.quickStartImage.ubuntu,
                                        command: ['sleep', 'infinity']
                                    }
                                ]
                            }
                        }
                    }
                }
            };

            const create = await workloadsManagerUtils.createWorkloadV2(
                workloadName,
                String(project.id),
                lwsManifest
            );
            const workloadId = create.body.metadata.id;
            await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running, 240);

            const seed = await captureV2WorkloadSeed({
                kind: 'LeaderWorkerSet',
                submissionPath: 'workloads-v2-api',
                id: workloadId,
                name: workloadName,
                namespace,
                projectId: String(project.id),
                crdKey: 'workloadRunnerCrd',
                innerManifest: lwsManifest
            });
            const softAssert = new SoftAssert();
            await verifyV2WorkloadSeedHealthy(seed, softAssert);
            softAssert.throwErrorIfFailed();
            (upgradeArtifact[v2WorkloadsKey] as V2WorkloadSeed[]).push(seed);
        });

        test('8) generate v2 Path B: PyTorchJob via kubectl apply (EWI)', async () => {
            // No version gate: EWI external-workload support predates every
            // version we'd ever run an upgrade test against.
            if (!(await isKubeflowInstalled())) {
                logger.warn('Kubeflow training-operator is not installed on the cluster, skipping test');
                test.skip();
            }

            if (!upgradeArtifact[v2WorkloadsKey]) {
                upgradeArtifact[v2WorkloadsKey] = [] as V2WorkloadSeed[];
            }

            const workloadName = objectGenerator.generateNameByPrefix('pytorch-ewi');
            const cpuOnly = {
                requests: { cpu: '100m', memory: '100Mi' },
                limits: { cpu: '100m', memory: '100Mi' }
            };
            const pytorchJob = new PyTorchJob(
                workloadName,
                namespace,
                config.quickStartImage.ubuntu,
                'sleep infinity',
                1, // masterReplicas
                1, // workerReplicas
                'OnFailure',
                cpuOnly,
                cpuOnly
            );
            const workloadId = await externalWorkloadUtils.submitWorkloadAndGetWorkloadId(
                pytorchJob,
                project.id
            );
            await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running, 240);

            const seed = await captureV2WorkloadSeed({
                kind: 'PyTorchJob',
                submissionPath: 'kubectl-apply-ewi',
                id: workloadId,
                name: workloadName,
                namespace,
                projectId: String(project.id),
                crdKey: 'externalWorkloadCrd',
                externalWorkloadWrapperName: `ew-pytorchjob-${workloadName}`,
                innerManifest: pytorchJob.getBody()
            });
            const softAssert = new SoftAssert();
            await verifyV2WorkloadSeedHealthy(seed, softAssert);
            softAssert.throwErrorIfFailed();
            (upgradeArtifact[v2WorkloadsKey] as V2WorkloadSeed[]).push(seed);
        });
    });
});
