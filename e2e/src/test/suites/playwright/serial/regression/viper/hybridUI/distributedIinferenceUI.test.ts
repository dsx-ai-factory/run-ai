/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    computeResourceUtils,
    config,
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    runaiHealthChecker,
    tags,
    User,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { expect } from 'chai';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test } from '@playwright/test';
import {
    EEnvironmentVariableSource,
    EVolumeAccessMode,
    IServingPortAccessOptions,
    IServingPortOptions,
    IVolumeOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import {
    Annotation,
    DistributedInference,
    DistributedInferenceStartupPolicy,
    ImagePullPolicy,
    Label
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    EImagePullPolicyOptions,
    IComputeOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import {
    ComputeAssetSpec,
    PVCAsset,
    PVCAssetSpec,
    PVCCreationRequest,
    PvcVolumeMode,
    Scope
} from '../../../../../../../modules/api/apis/assetsApi/generated/api';
import { InferenceFlatUIUtil } from '../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import { ComputeAsset } from '../../../../../../../modules/api/apis/assetsApi/generated/api';
import { InferencePage } from '../../../../../../../modules/ui/runaiV2/pages/InferencePage';

test.describe(
    `Inferences distributed UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper} ${tags.runOn.nightly}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'distributed-inference-ui';
        let traceName: string;

        let project: Project;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let inference: InferencePage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let inferenceName: string;
        let inferenceWorkload: DistributedInference;

        const currentUser: User = flowRootState.users.system_administrator;

        function findItem(
            items: Array<Annotation | Label> | null,
            itemName: string,
            itemValue: string
        ): boolean {
            const item = items?.find((item) => item.name === itemName && item.value === itemValue);
            return !!item;
        }

        async function createCompute(spec: ComputeAssetSpec): Promise<ComputeAsset> {
            try {
                const computeResourceName = objectGenerator.generateShortNameByPrefix('cr', 5);
                const computeCreationRequest = {
                    meta: {
                        name: computeResourceName,
                        scope: Scope.Cluster,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec
                };

                const computeAsset = await computeResourceUtils.create(computeCreationRequest);
                logger.debug(`computeAsset created: ${JSON.stringify(computeAsset)}`);
                return computeAsset;
            } catch (error) {
                logger.error(`failed to create compute resource: ${error}`);
                throw error;
            }
        }

        async function createPvcAsset(specOverrides?: Partial<PVCAssetSpec>): Promise<PVCAsset> {
            try {
                const name = objectGenerator.generateNameByPrefix('inf-pvc-storage');

                const defaultSpec: PVCAssetSpec = {
                    path: '/home',
                    claimName: objectGenerator.generateNameByPrefix('inf-pvc-storage-claim'),
                    existingPvc: false,
                    readOnly: false,
                    ephemeral: false,
                    claimInfo: {
                        size: '100M',
                        accessModes: {
                            readWriteOnce: true
                        },
                        volumeMode: PvcVolumeMode.Filesystem
                    }
                };

                const pvc: PVCCreationRequest = {
                    meta: {
                        name: name,
                        scope: Scope.Project,
                        projectId: parseInt(project.id)
                    },
                    spec: { ...defaultSpec, ...specOverrides }
                };

                const pvcAsset = await dataSourcesUtils.createPvcAsset(pvc);
                logger.info(`created pvc asset ${pvcAsset.meta.name}`);

                await dataSourcesUtils.waitForPvcAssetToBeReady(
                    pvcAsset.meta.name,
                    pvcAsset.meta.id,
                    parseInt(project.id),
                    'Inference'
                );
                logger.info(`created pvc asset ${pvcAsset.meta.name} is ready`);
                return pvcAsset;
            } catch (error) {
                logger.error(`failed to create pvc asset: ${error}`);
                throw error;
            }
        }

        async function initBrowser(): Promise<BrowserEngine> {
            // Initialize browser once for all tests with video recording AND tracing support
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            // inference = new InferencePage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            inference = new InferencePage(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);

            await auth.signIn(currentUser.email, currentUser.password);

            return browserEngine;
        }

        test.beforeAll('before all create distributed inference workload', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('p-inf');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });

            logger.info(`created project ${projectName}`);

            browserEngine = await initBrowser();

            logger.info('Browser initialized and user signed in');
        });

        test.afterAll('after all create distributed inference workload', async ({}, testInfo) => {
            // Close browser once after all tests and save video
            if (browserEngine && browserEngine.browser) {
                try {
                    await browserEngine.stopAndArchive(testInfo, suiteName);
                    logger.info('Browser closed and video saved');
                } catch (error) {
                    logger.error(`Error closing browser: ${error}`);
                }
            }

            if (!(await runaiHealthChecker.isSelfHostedCluster())) {
                // This condition should be removed when fixing problem with saas envs with cluster version.
                return;
            }

            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                return;
            }

            try {
                await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id);
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                logger.info(
                    `deleted workloads and project ${project.name} and namespace ${project.status.namespace}`
                );
            } catch (error) {
                logger.error(
                    `failed to delete workloads and project ${project.name} and namespace ${project.status.namespace}. Error: ${JSON.stringify(error)}`
                );
            }
        });

        test.beforeEach('before each create distributed inference workload', async function () {
            if (!browserEngine || !browserEngine.browser) {
                browserEngine = await initBrowser();
            }

            inferenceName = objectGenerator.generateNameByPrefix('inference');

            // Generate trace name from test title
            traceName = objectGenerator.generateNameByPrefix(suiteName);

            // Start tracing for this test (context and page stay alive)
            await browserEngine.startTracing(traceName);

            // Navigate to workloads page for each test
            await workload.navigateToWorkloads();
        });

        test.afterEach('after each create distributed inference workload', async ({}, testInfo) => {
            logger.debug(`waiting for workload ${inferenceName} to be deleted`);

            // delete inference
            if (inferenceWorkload) {
                await workloadsUtilsV2.deleteWorkload(
                    inferenceWorkload.workloadId,
                    WorkloadType.INFERENCE,
                    flowRootState.users.system_administrator.accessToken
                );
            }
            inferenceWorkload = undefined;

            logger.debug(`inference ${inferenceName} deleted!`);

            // Stop and archive tracing for this test (browser, context, and page remain alive)
            await browserEngine.stopAndArchiveTracing(testInfo, traceName);
        });

        supportedFrom(SupportedVersions.V2_23)(function () {
            test('should create distributed inference with minimal config and validate default workload settings', async function () {
                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        servingPort: {
                            port: 4000
                        }
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // ----- leader -----
                // environment
                expect(inferenceWorkload.spec.servingPort.port).to.equal(
                    leader.environmentSection.servingPort.port
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');

                // ----- workload setup -----
                expect(inferenceWorkload.spec.priorityClass).to.equal('very-high');
                expect(inferenceWorkload.spec.startupPolicy).to.equal(
                    DistributedInferenceStartupPolicy.LeaderCreated
                );

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            test('should create distributed inference with full leader config and custom workload setup', async function () {
                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const computeUpdates: IComputeOptions = {
                    gpu: {
                        devicesInput: '3',
                        memoryRequest: '30'
                    },
                    cpu: {
                        compute: {
                            request: '4'
                        },
                        memory: {
                            request: '5'
                        }
                    },
                    extendedResource: {
                        increaseSharedMemory: true,
                        resources: [
                            { resource: 'puki/shuki', quantity: '300' },
                            { resource: 'puki/tuki', quantity: '301' }
                        ]
                    }
                };

                const volumeOptions: IVolumeOptions = {
                    accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                    sizeUnit: 'MB',
                    path: '/home'
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""',
                            fullEnvironmentVariables: [
                                { name: 'MY_VAR_1', value: 'value1' },
                                {
                                    name: 'MY_VAR_2',
                                    source: EEnvironmentVariableSource.Custom,
                                    value: 'value2'
                                }
                            ]
                        },
                        servingPort: {
                            port: 4000,
                            access: {
                                option: IServingPortAccessOptions.Users
                            }
                        }
                    },
                    computeSection: computeUpdates,
                    dataSourceOptions: {
                        volume: volumeOptions
                    },
                    generalSectionOptions: {
                        annotations: [
                            { name: 'ann1', value: 'ann-val1' },
                            { name: 'ann2', value: 'ann-val2' }
                        ],
                        labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                    }
                };

                const workloadSetup = {
                    generalSectionOptions: {
                        priority: 'low',
                        startupPolicy: 'Once the leader pod is ready'
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader,
                    workloadSetup
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // ----- leader -----
                // environment
                expect(inferenceWorkload.spec.leader.image).to.equal(leader.environmentSection.image);
                expect(inferenceWorkload.spec.leader.imagePullPolicy).to.equal('Always');
                expect(inferenceWorkload.spec.leader.command).to.equal(
                    leader.environmentSection.runtimeSettings.cmd
                );
                expect(inferenceWorkload.spec.leader.args).to.equal(
                    leader.environmentSection.runtimeSettings.args
                );
                for (const specEnvVar of inferenceWorkload.spec.leader.environmentVariables) {
                    const matchingEnvVar =
                        leader.environmentSection.runtimeSettings.fullEnvironmentVariables.find(
                            (uiEnvVar) => uiEnvVar.name === specEnvVar.name
                        );
                    expect(matchingEnvVar).to.not.be.undefined;
                    expect(matchingEnvVar.value).to.equal(specEnvVar.value);
                }
                expect(inferenceWorkload.spec.servingPort.port).to.equal(
                    leader.environmentSection.servingPort.port
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal('authorizedUsers');
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;

                // compute
                expect(inferenceWorkload.spec.leader.compute.gpuDevicesRequest).to.equal(3);
                expect(inferenceWorkload.spec.leader.compute.gpuRequestType).to.equal('portion');
                expect(inferenceWorkload.spec.leader.compute.gpuPortionRequest).to.equal(0.3);
                expect(inferenceWorkload.spec.leader.compute.cpuCoreRequest).to.equal(0.004);
                expect(inferenceWorkload.spec.leader.compute.cpuMemoryRequest).to.equal('5M');
                const resources = computeUpdates.extendedResource.resources.map((resource) => {
                    return {
                        ...resource,
                        exclude: undefined
                    };
                });
                expect(inferenceWorkload.spec.leader.compute.extendedResources).to.deep.equal(resources);

                // datasources
                const pvc = inferenceWorkload.spec.leader.storage.pvc[0];
                expect(pvc).to.be.exist;
                expect(pvc.path).to.equal('/home');
                expect(pvc.claimInfo.size).to.equal('1M');
                expect(pvc.claimInfo.storageClass).to.be.exist;
                expect(pvc.claimInfo.accessModes.readWriteOnce).to.equal(true);
                expect(pvc.claimInfo.accessModes.readOnlyMany).to.be.undefined;
                expect(pvc.claimInfo.accessModes.readWriteMany).to.be.undefined;

                // general section
                expect(findItem(inferenceWorkload.spec.leader.annotations, 'ann1', 'ann-val1')).to.be.true;
                expect(findItem(inferenceWorkload.spec.leader.annotations, 'ann2', 'ann-val2')).to.be.true;
                expect(findItem(inferenceWorkload.spec.leader.labels, 'lbl1', 'lbl-val1')).to.be.true;

                // ----- worker -----
                try {
                    expect(inferenceWorkload.spec.worker).to.deep.equal(inferenceWorkload.spec.leader);
                } catch (error) {
                    logger.error(`error comparing worker and leader: ${error}`);
                    logger.debug(`leader: ${JSON.stringify(inferenceWorkload.spec.leader)}`);
                    logger.debug(`worker: ${JSON.stringify(inferenceWorkload.spec.worker)}`);
                    throw error;
                }

                // ----- workload setup -----
                expect(inferenceWorkload.spec.priorityClass).to.equal(
                    workloadSetup.generalSectionOptions.priority
                );
                expect(inferenceWorkload.spec.startupPolicy).to.equal(
                    DistributedInferenceStartupPolicy.LeaderReady
                );

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            test('should create distributed inference with different leader and worker configurations', async function () {
                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""',
                            fullEnvironmentVariables: [
                                { name: 'MY_VAR_1', value: 'value1' },
                                {
                                    name: 'MY_VAR_2',
                                    source: EEnvironmentVariableSource.Custom,
                                    value: 'value2'
                                }
                            ]
                        },
                        servingPort: {
                            port: 4000,
                            access: {
                                option: IServingPortAccessOptions.Users
                            }
                        }
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '3',
                            memoryRequest: '30'
                        },
                        cpu: {
                            compute: {
                                request: '4'
                            },
                            memory: {
                                request: '5'
                            }
                        },
                        extendedResource: {
                            increaseSharedMemory: true,
                            resources: [
                                { resource: 'puki/shuki', quantity: '300' },
                                { resource: 'puki/tuki', quantity: '301' }
                            ]
                        }
                    },
                    dataSourceOptions: {
                        volume: {
                            accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                            sizeUnit: 'MB',
                            path: '/home'
                        }
                    },
                    generalSectionOptions: {
                        annotations: [
                            { name: 'ann1', value: 'ann-val1' },
                            { name: 'ann2', value: 'ann-val2' }
                        ],
                        labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                    }
                };

                const worker = {
                    environmentSection: {
                        imagePullPolicy: EImagePullPolicyOptions.ImageOnly,
                        runtimeSettings: {
                            cmd: 'start-notebook2.sh',
                            args: '--NotebookApp'
                        }
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '5',
                            memoryRequest: '50'
                        },
                        cpu: {
                            compute: {
                                request: '6'
                            },
                            memory: {
                                request: '7'
                            }
                        }
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader,
                    worker
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // ----- leader -----
                // environment
                expect(inferenceWorkload.spec.leader.image).to.equal(leader.environmentSection.image);
                expect(inferenceWorkload.spec.leader.imagePullPolicy).to.equal(ImagePullPolicy.Always);
                expect(inferenceWorkload.spec.leader.command).to.equal(
                    leader.environmentSection.runtimeSettings.cmd
                );
                expect(inferenceWorkload.spec.leader.args).to.equal(
                    leader.environmentSection.runtimeSettings.args
                );
                for (const specEnvVar of inferenceWorkload.spec.leader.environmentVariables) {
                    const matchingEnvVar =
                        leader.environmentSection.runtimeSettings.fullEnvironmentVariables.find(
                            (uiEnvVar) => uiEnvVar.name === specEnvVar.name
                        );
                    expect(matchingEnvVar).to.not.be.undefined;
                    expect(matchingEnvVar.value).to.equal(specEnvVar.value);
                }
                expect(inferenceWorkload.spec.servingPort.port).to.equal(
                    leader.environmentSection.servingPort.port
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal('authorizedUsers');
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;

                // compute
                expect(inferenceWorkload.spec.leader.compute.gpuDevicesRequest).to.equal(3);
                expect(inferenceWorkload.spec.leader.compute.gpuRequestType).to.equal('portion');
                expect(inferenceWorkload.spec.leader.compute.gpuPortionRequest).to.equal(0.3);
                expect(inferenceWorkload.spec.leader.compute.cpuCoreRequest).to.equal(0.004);
                expect(inferenceWorkload.spec.leader.compute.cpuMemoryRequest).to.equal('5M');
                const resources = leader.computeSection.extendedResource.resources.map((resource) => {
                    return {
                        ...resource,
                        exclude: undefined
                    };
                });
                expect(inferenceWorkload.spec.leader.compute.extendedResources).to.deep.equal(resources);

                // datasources
                const pvc = inferenceWorkload.spec.leader.storage.pvc[0];
                expect(pvc).to.be.exist;
                expect(pvc.path).to.equal('/home');
                expect(pvc.claimInfo.size).to.equal('1M');
                expect(pvc.claimInfo.storageClass).to.be.exist;
                expect(pvc.claimInfo.accessModes.readWriteOnce).to.equal(true);
                expect(pvc.claimInfo.accessModes.readOnlyMany).to.be.undefined;
                expect(pvc.claimInfo.accessModes.readWriteMany).to.be.undefined;

                // general section
                expect(findItem(inferenceWorkload.spec.leader.annotations, 'ann1', 'ann-val1')).to.be.true;
                expect(findItem(inferenceWorkload.spec.leader.annotations, 'ann2', 'ann-val2')).to.be.true;
                expect(findItem(inferenceWorkload.spec.leader.labels, 'lbl1', 'lbl-val1')).to.be.true;

                // ----- worker -----
                // environment
                expect(inferenceWorkload.spec.worker.imagePullPolicy).to.equal(ImagePullPolicy.IfNotPresent);
                expect(inferenceWorkload.spec.worker.command).to.equal(
                    worker.environmentSection.runtimeSettings.cmd
                );
                expect(inferenceWorkload.spec.worker.args).to.equal(
                    worker.environmentSection.runtimeSettings.args
                );

                // compute
                expect(inferenceWorkload.spec.worker.compute.gpuDevicesRequest).to.equal(5);
                expect(inferenceWorkload.spec.worker.compute.gpuRequestType).to.equal('portion');
                expect(inferenceWorkload.spec.worker.compute.gpuPortionRequest).to.equal(0.5);
                expect(inferenceWorkload.spec.worker.compute.cpuCoreRequest).to.equal(0.006);
                expect(inferenceWorkload.spec.worker.compute.cpuMemoryRequest).to.equal('7M');
                expect(inferenceWorkload.spec.worker.compute.extendedResources).to.deep.equal(resources);

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            test('should create distributed inference with leader using existing compute and PVC assets', async function () {
                const computeAsset = await createCompute({
                    cpuCoreRequest: 0.2,
                    cpuMemoryRequest: '200M',
                    gpuDevicesRequest: 2
                });

                const pvcAsset = await createPvcAsset();

                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        servingPort: {
                            port: 4000
                        }
                    },
                    computeSection: {
                        assetName: computeAsset.meta.name,
                        waitForUnifiedRequest: false
                    },
                    dataSourceOptions: {
                        assetNames: [pvcAsset.meta.name],
                        waitForUnifiedRequest: false
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // compute section
                expect(inferenceWorkload.spec.leader.compute.cpuCoreRequest).to.equal(
                    computeAsset.spec.cpuCoreRequest
                );
                expect(inferenceWorkload.spec.leader.compute.cpuMemoryRequest).to.equal(
                    computeAsset.spec.cpuMemoryRequest
                );
                expect(inferenceWorkload.spec.leader.compute.gpuDevicesRequest).to.equal(
                    computeAsset.spec.gpuDevicesRequest
                );

                // storage section
                expect(inferenceWorkload.spec.leader.storage.pvc[0]).to.be.exist;
                expect(inferenceWorkload.spec.leader.storage.pvc[0].path).to.equal(pvcAsset.spec.path);
                expect(inferenceWorkload.spec.leader.storage.pvc[0].existingPvc).to.equal(true);
                expect(inferenceWorkload.spec.leader.storage.pvc[0].claimName).to.equal(
                    pvcAsset.spec.claimName
                );
                expect(inferenceWorkload.spec.leader.storage.pvc[0].readOnly).to.equal(
                    pvcAsset.spec.readOnly
                );

                // cleanup
                await Promise.all([
                    computeResourceUtils.delete(computeAsset.meta.id),
                    dataSourcesUtils.deletePvcAsset(pvcAsset.meta.id)
                ]);

                await Promise.all([
                    computeResourceUtils.waitForComputeResourceToBeDeleted(computeAsset.meta.name),
                    dataSourcesUtils.waitForDataSourceToBeDeleted(pvcAsset.meta.name)
                ]);

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            test('should create distributed inference with worker using existing compute and PVC assets', async function () {
                const computeAsset = await createCompute({
                    cpuCoreRequest: 0.2,
                    cpuMemoryRequest: '200M',
                    gpuDevicesRequest: 2
                });

                const pvcAsset = await createPvcAsset();

                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        servingPort: {
                            port: 4000
                        }
                    }
                };

                const worker = {
                    computeSection: {
                        assetName: computeAsset.meta.name,
                        waitForUnifiedRequest: false
                    },
                    dataSourceOptions: {
                        assetNames: [pvcAsset.meta.name],
                        waitForUnifiedRequest: false
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader,
                    worker
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // compute section
                expect(inferenceWorkload.spec.worker.compute.cpuCoreRequest).to.equal(
                    computeAsset.spec.cpuCoreRequest
                );
                expect(inferenceWorkload.spec.worker.compute.cpuMemoryRequest).to.equal(
                    computeAsset.spec.cpuMemoryRequest
                );
                expect(inferenceWorkload.spec.worker.compute.gpuDevicesRequest).to.equal(
                    computeAsset.spec.gpuDevicesRequest
                );

                // storage section
                expect(inferenceWorkload.spec.worker.storage.pvc[0]).to.be.exist;
                expect(inferenceWorkload.spec.worker.storage.pvc[0].path).to.equal(pvcAsset.spec.path);
                expect(inferenceWorkload.spec.worker.storage.pvc[0].existingPvc).to.equal(true);
                expect(inferenceWorkload.spec.worker.storage.pvc[0].claimName).to.equal(
                    pvcAsset.spec.claimName
                );
                expect(inferenceWorkload.spec.worker.storage.pvc[0].readOnly).to.equal(
                    pvcAsset.spec.readOnly
                );

                // cleanup
                await Promise.all([
                    computeResourceUtils.delete(computeAsset.meta.id),
                    dataSourcesUtils.deletePvcAsset(pvcAsset.meta.id)
                ]);
                await Promise.all([
                    computeResourceUtils.waitForComputeResourceToBeDeleted(computeAsset.meta.name),
                    dataSourcesUtils.waitForDataSourceToBeDeleted(pvcAsset.meta.name)
                ]);

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            test('should create distributed inference with leader using new environment, compute resource and data sources assets', async function () {
                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const pvc = {
                    name: objectGenerator.generateNameByPrefix('inf-pvc-storage'),
                    storageClass: 'none',
                    existingPvc: false,
                    readOnly: false,
                    ephemeral: false,
                    containerPath: '/home/data'
                };

                const servingPort: IServingPortOptions = {
                    port: 3040
                };
                const leader = {
                    environmentSection: {
                        newEnvironmentResource: {
                            assetName: objectGenerator.generateNameByPrefix('inf-environment'),
                            projectName: project.name,
                            image: config.quickStartImage.inferenceHelloWorld,
                            servingPort: {
                                container: servingPort.port
                            },
                            runtimeSettings: {
                                cmd: 'start-notebook.sh',
                                args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                            }
                        }
                    },
                    computeSection: {
                        newComputeResource: {
                            assetName: objectGenerator.generateNameByPrefix('inf-compute'),
                            projectName: project.name,
                            gpu: {
                                devicesInput: 3
                            },
                            isProjectScope: false
                        }
                    },
                    dataSourceOptions: {
                        pvc
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // environment section
                expect(inferenceWorkload.spec.leader.image).to.equal(
                    leader.environmentSection.newEnvironmentResource.image
                );
                expect(inferenceWorkload.spec.servingPort.port).to.equal(
                    leader.environmentSection.newEnvironmentResource.servingPort.container
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
                expect(inferenceWorkload.spec.leader.command).to.equal(
                    leader.environmentSection.newEnvironmentResource.runtimeSettings.cmd
                );
                expect(inferenceWorkload.spec.leader.args).to.equal(
                    leader.environmentSection.newEnvironmentResource.runtimeSettings.args
                );

                // compute section
                expect(inferenceWorkload.spec.leader.compute.gpuDevicesRequest).to.be.equal(1);
                expect(inferenceWorkload.spec.leader.compute.gpuPortionRequest).to.be.equal(0.03);
                expect(inferenceWorkload.spec.leader.compute.gpuRequestType).to.be.equal('portion');

                // storage section
                expect(inferenceWorkload.spec.leader.storage.pvc).to.be.exist;
                expect(inferenceWorkload.spec.leader.storage.pvc[0].path).to.equal(pvc.containerPath);
                expect(inferenceWorkload.spec.leader.storage.pvc[0].existingPvc).to.equal(true);
                expect(inferenceWorkload.spec.leader.storage.pvc[0].readOnly).to.equal(pvc.readOnly);

                // cleanup
                const computeAssetId = await computeResourceUtils.getId(
                    leader.computeSection.newComputeResource.assetName
                );
                const pvcAsset = await dataSourcesUtils.getPvcByName(pvc.name);
                await Promise.all([
                    computeResourceUtils.delete(computeAssetId),
                    dataSourcesUtils.deletePvcAsset(pvcAsset.meta.id)
                ]);

                await Promise.all([
                    computeResourceUtils.waitForComputeResourceToBeDeleted(
                        leader.computeSection.newComputeResource.assetName
                    ),
                    dataSourcesUtils.waitForDataSourceToBeDeleted(pvc.name)
                ]);

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            test('should create distributed inference with worker using new environment, compute resource and data sources assets', async function () {
                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const pvc = {
                    name: objectGenerator.generateNameByPrefix('inf-pvc-storage'),
                    storageClass: 'none',
                    existingPvc: false,
                    readOnly: false,
                    ephemeral: false,
                    containerPath: '/home/data'
                };

                const leader = {
                    environmentSection: {
                        image: 'ubuntu',
                        servingPort: {
                            port: 4000
                        }
                    }
                };

                const worker = {
                    environmentSection: {
                        newEnvironmentResource: {
                            assetName: objectGenerator.generateNameByPrefix('inf-environment'),
                            projectName: project.name,
                            image: config.quickStartImage.inferenceHelloWorld,
                            runtimeSettings: {
                                cmd: 'start-notebook.sh',
                                args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                            },
                            servingPort: { // in environment creations we must provide the serving port
                                container: 2020
                            },
                        }
                    },
                    computeSection: {
                        newComputeResource: {
                            assetName: objectGenerator.generateNameByPrefix('inf-compute'),
                            projectName: project.name,
                            gpu: {
                                devicesInput: 3
                            },
                            isProjectScope: false
                        }
                    },
                    dataSourceOptions: {
                        pvc
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader,
                    worker
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // leader sections
                // environment section
                expect(inferenceWorkload.spec.leader.image).to.equal(leader.environmentSection.image);
                expect(inferenceWorkload.spec.leader.command).to.be.undefined;
                expect(inferenceWorkload.spec.leader.args).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.port).to.equal(
                    leader.environmentSection.servingPort.port
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');

                // worker sections
                // environment section
                expect(inferenceWorkload.spec.worker.image).to.equal(worker.environmentSection.newEnvironmentResource.image);
                expect(inferenceWorkload.spec.worker.command).to.equal(
                    worker.environmentSection.newEnvironmentResource.runtimeSettings.cmd
                );
                expect(inferenceWorkload.spec.worker.args).to.equal(
                    worker.environmentSection.newEnvironmentResource.runtimeSettings.args
                );

                // compute section
                expect(inferenceWorkload.spec.worker.compute.gpuDevicesRequest).to.be.equal(1);
                expect(inferenceWorkload.spec.worker.compute.gpuPortionRequest).to.be.equal(0.03);
                expect(inferenceWorkload.spec.worker.compute.gpuRequestType).to.be.equal('portion');

                // storage section
                expect(inferenceWorkload.spec.worker.storage.pvc).to.be.exist;
                expect(inferenceWorkload.spec.worker.storage.pvc[0].path).to.equal(pvc.containerPath);
                expect(inferenceWorkload.spec.worker.storage.pvc[0].existingPvc).to.equal(true);
                expect(inferenceWorkload.spec.worker.storage.pvc[0].readOnly).to.equal(pvc.readOnly);

                // cleanup
                const computeAssetId = await computeResourceUtils.getId(
                    worker.computeSection.newComputeResource.assetName
                );
                const pvcAsset = await dataSourcesUtils.getPvcByName(pvc.name);
                await Promise.all([
                    computeResourceUtils.delete(computeAssetId),
                    dataSourcesUtils.deletePvcAsset(pvcAsset.meta.id)
                ]);

                await Promise.all([
                    computeResourceUtils.waitForComputeResourceToBeDeleted(
                        worker.computeSection.newComputeResource.assetName
                    ),
                    dataSourcesUtils.waitForDataSourceToBeDeleted(pvc.name)
                ]);

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            test('should copy distributed inference workload', async function () {
                const copiedInferenceName = objectGenerator.generateShortNameByPrefix('copied-dist-inf');

                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const volumeOptions: IVolumeOptions = {
                    accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                    sizeUnit: 'MB',
                    path: '/home'
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}'
                        },
                        servingPort: {
                            port: 4000
                        }
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '2'
                        }
                    },
                    dataSourceOptions: {
                        volume: volumeOptions
                    },
                    generalSectionOptions: {
                        annotations: [{ name: 'ann1', value: 'ann-val1' }],
                        labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                    }
                };

                const workloadSetup = {
                    generalSectionOptions: {
                        priority: 'low',
                        startupPolicy: 'Once the leader pod is ready'
                    }
                };

                // Create the original distributed inference workload
                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader,
                    workloadSetup
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`originalDistributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // Copy the workload
                await workload.toggleWorkload(inferenceName);
                await workload.clickOnCopyAndEditWorkload();

                await inference.fillInferenceName(copiedInferenceName);

                await inference.clickOnElementByText('.runai-folding-sub-section-header', 'Image pull');
                const imagePullPolicy = await inference.page
                    .locator('.q-radio[aria-checked="true"]')
                    .filter({ has: inference.page.locator('[aid*="pull the image"], [aid*="Never pull"]') })
                    .locator('span.ellipsis-label')
                    .textContent();
                expect(imagePullPolicy.trim()).to.equal('Always pull the image from the registry');

                // Navigate through the wizard steps to submit
                await inference.clickOnCreateInferenceBtn(); // Next to worker step
                await inference.page.waitForTimeout(1000); // wait for 1 second to ensure the section is opened
                await inference.clickOnCreateInferenceBtn(); // Next to workload setup step
                await inference.page.waitForTimeout(1000); // wait for 1 second to ensure the section is opened
                await inference.clickOnCreateInferenceBtn(); // Submit

                await waitUtils.waitForElementToExist(workload, `td:has-text("${copiedInferenceName}")`);

                const copiedInferenceWorkload: DistributedInference =
                    await workloadsCrudUtils.getDistributedInferenceByName(
                        copiedInferenceName,
                        project.id.toString()
                    );
                expect(copiedInferenceWorkload).to.not.be.undefined;
                logger.debug(`copiedDistributedWorkload: ${JSON.stringify(copiedInferenceWorkload)}`);

                // Verify copied workload has same configuration as original
                // Leader configuration
                expect(copiedInferenceWorkload.spec.leader.image).to.equal(
                    inferenceWorkload.spec.leader.image
                );
                expect(copiedInferenceWorkload.spec.leader.imagePullPolicy).to.equal(
                    inferenceWorkload.spec.leader.imagePullPolicy
                );
                expect(copiedInferenceWorkload.spec.leader.command).to.equal(
                    inferenceWorkload.spec.leader.command
                );
                expect(copiedInferenceWorkload.spec.leader.args).to.equal(inferenceWorkload.spec.leader.args);

                // Compute
                expect(copiedInferenceWorkload.spec.leader.compute.gpuDevicesRequest).to.equal(
                    inferenceWorkload.spec.leader.compute.gpuDevicesRequest
                );

                // Storage
                expect(copiedInferenceWorkload.spec.leader.storage.pvc[0].path).to.equal(
                    inferenceWorkload.spec.leader.storage.pvc[0].path
                );
                expect(copiedInferenceWorkload.spec.leader.storage.pvc[0].claimInfo).to.deep.equal(
                    inferenceWorkload.spec.leader.storage.pvc[0].claimInfo
                );

                // Annotations and labels
                expect(findItem(copiedInferenceWorkload.spec.leader.annotations, 'ann1', 'ann-val1')).to.be
                    .true;
                expect(findItem(copiedInferenceWorkload.spec.leader.labels, 'lbl1', 'lbl-val1')).to.be.true;

                // Worker should match leader (since no worker config was specified)
                expect(copiedInferenceWorkload.spec.worker.image).to.equal(
                    copiedInferenceWorkload.spec.leader.image
                );

                // Workload setup
                expect(copiedInferenceWorkload.spec.priorityClass).to.equal(
                    inferenceWorkload.spec.priorityClass
                );
                expect(copiedInferenceWorkload.spec.startupPolicy).to.equal(
                    inferenceWorkload.spec.startupPolicy
                );

                // Serving port
                expect(copiedInferenceWorkload.spec.servingPort.port).to.equal(
                    inferenceWorkload.spec.servingPort.port
                );

                logger.debug(`distributedWorkload ${copiedInferenceName} copy test ended successfully!`);
            });
        });

        supportedFrom(SupportedVersions.V2_24)(function () {
            // this test is similar to the one above, but with the new preemptibility field that supports only in V2_24+
            test('should create distributed inference with minimal config and validate default workload settings with preemptibility', async function () {
                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        servingPort: {
                            port: 4000
                        }
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // ----- leader -----
                // environment
                expect(inferenceWorkload.spec.servingPort.port).to.equal(
                    leader.environmentSection.servingPort.port
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');

                // ----- workload setup -----
                expect(inferenceWorkload.spec.preemptibility).to.equal('non-preemptible');
                expect(inferenceWorkload.spec.priorityClass).to.equal('very-high');
                expect(inferenceWorkload.spec.startupPolicy).to.equal(
                    DistributedInferenceStartupPolicy.LeaderCreated
                );

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });

            // this test is similar to the one above, but with the new preemptibility field that supports only in V2_24+
            test('should create distributed inference with full leader config and custom workload setup with preemptibility', async function () {
                const meta = {
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName
                };

                const computeUpdates: IComputeOptions = {
                    gpu: {
                        devicesInput: '3',
                        memoryRequest: '30'
                    },
                    cpu: {
                        compute: {
                            request: '4'
                        },
                        memory: {
                            request: '5'
                        }
                    },
                    extendedResource: {
                        increaseSharedMemory: true,
                        resources: [
                            { resource: 'puki/shuki', quantity: '300' },
                            { resource: 'puki/tuki', quantity: '301' }
                        ]
                    }
                };

                const volumeOptions: IVolumeOptions = {
                    accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                    sizeUnit: 'MB',
                    path: '/home'
                };

                const leader = {
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""',
                            fullEnvironmentVariables: [
                                { name: 'MY_VAR_1', value: 'value1' },
                                {
                                    name: 'MY_VAR_2',
                                    source: EEnvironmentVariableSource.Custom,
                                    value: 'value2'
                                }
                            ]
                        },
                        servingPort: {
                            port: 4000,
                            access: {
                                option: IServingPortAccessOptions.Users
                            }
                        }
                    },
                    computeSection: computeUpdates,
                    dataSourceOptions: {
                        volume: volumeOptions
                    },
                    generalSectionOptions: {
                        annotations: [
                            { name: 'ann1', value: 'ann-val1' },
                            { name: 'ann2', value: 'ann-val2' }
                        ],
                        labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                    }
                };

                const workloadSetup = {
                    generalSectionOptions: {
                        priority: 'low',
                        preemptibility: 'Preemptible',
                        startupPolicy: 'Once the leader pod is ready'
                    }
                };

                await inferenceFlatUIUtil.createDistributedWorkload({
                    ...meta,
                    leader,
                    workloadSetup
                });

                inferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`distributedWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // ----- leader -----
                // environment
                expect(inferenceWorkload.spec.leader.image).to.equal(leader.environmentSection.image);
                expect(inferenceWorkload.spec.leader.imagePullPolicy).to.equal('Always');
                expect(inferenceWorkload.spec.leader.command).to.equal(
                    leader.environmentSection.runtimeSettings.cmd
                );
                expect(inferenceWorkload.spec.leader.args).to.equal(
                    leader.environmentSection.runtimeSettings.args
                );
                for (const specEnvVar of inferenceWorkload.spec.leader.environmentVariables) {
                    const matchingEnvVar =
                        leader.environmentSection.runtimeSettings.fullEnvironmentVariables.find(
                            (uiEnvVar) => uiEnvVar.name === specEnvVar.name
                        );
                    expect(matchingEnvVar).to.not.be.undefined;
                    expect(matchingEnvVar.value).to.equal(specEnvVar.value);
                }
                expect(inferenceWorkload.spec.servingPort.port).to.equal(
                    leader.environmentSection.servingPort.port
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal('authorizedUsers');
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;

                // compute
                expect(inferenceWorkload.spec.leader.compute.gpuDevicesRequest).to.equal(3);
                expect(inferenceWorkload.spec.leader.compute.gpuRequestType).to.equal('portion');
                expect(inferenceWorkload.spec.leader.compute.gpuPortionRequest).to.equal(0.3);
                expect(inferenceWorkload.spec.leader.compute.cpuCoreRequest).to.equal(0.004);
                expect(inferenceWorkload.spec.leader.compute.cpuMemoryRequest).to.equal('5M');
                const resources = computeUpdates.extendedResource.resources.map((resource) => {
                    return {
                        ...resource,
                        exclude: undefined
                    };
                });
                expect(inferenceWorkload.spec.leader.compute.extendedResources).to.deep.equal(resources);

                // datasources
                const pvc = inferenceWorkload.spec.leader.storage.pvc[0];
                expect(pvc).to.be.exist;
                expect(pvc.path).to.equal('/home');
                expect(pvc.claimInfo.size).to.equal('1M');
                expect(pvc.claimInfo.storageClass).to.be.exist;
                expect(pvc.claimInfo.accessModes.readWriteOnce).to.equal(true);
                expect(pvc.claimInfo.accessModes.readOnlyMany).to.be.undefined;
                expect(pvc.claimInfo.accessModes.readWriteMany).to.be.undefined;

                // general section
                expect(findItem(inferenceWorkload.spec.leader.annotations, 'ann1', 'ann-val1')).to.be.true;
                expect(findItem(inferenceWorkload.spec.leader.annotations, 'ann2', 'ann-val2')).to.be.true;
                expect(findItem(inferenceWorkload.spec.leader.labels, 'lbl1', 'lbl-val1')).to.be.true;

                // ----- worker -----
                try {
                    expect(inferenceWorkload.spec.worker).to.deep.equal(inferenceWorkload.spec.leader);
                } catch (error) {
                    logger.error(`error comparing worker and leader: ${error}`);
                    logger.debug(`leader: ${JSON.stringify(inferenceWorkload.spec.leader)}`);
                    logger.debug(`worker: ${JSON.stringify(inferenceWorkload.spec.worker)}`);
                    throw error;
                }

                // ----- workload setup -----
                expect(inferenceWorkload.spec.preemptibility).to.equal(
                    workloadSetup.generalSectionOptions.preemptibility.toLowerCase()
                );
                expect(inferenceWorkload.spec.priorityClass).to.equal(
                    workloadSetup.generalSectionOptions.priority
                );
                expect(inferenceWorkload.spec.startupPolicy).to.equal(
                    DistributedInferenceStartupPolicy.LeaderReady
                );

                logger.debug(`distributedWorkload ${inferenceName} test ended successfully!`);
            });
        });
    }
);
