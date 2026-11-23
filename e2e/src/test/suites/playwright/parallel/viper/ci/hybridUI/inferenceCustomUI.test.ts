/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { InferencePage } from '../../../../../../../modules/ui/runaiV2/pages/InferencePage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    computeResourceUtils,
    config,
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
    EVolumeAccessMode,
    IInferenceType,
    IServingPortAccessOptions,
    IVolumeOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import {
    Annotation,
    Inference,
    Label
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { IUpdateWorkload } from '../../../../../../../modules/helpers/models/inferenceFlatUI.models';
import {
    EImagePullPolicyOptions,
    IAutoScalingOptions,
    IComputeOptions,
    IDataSourceOptions,
    IEnvironmentOptions,
    IGeneralSectionOptions,
    IServingPortOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { InferenceFlatUIUtil } from '../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(
    `Inferences Custom UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper} ${tags.runOn.parallelCI} ${tags.runOn.nightly}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'inference-custom-ui';
        let traceName: string;

        let project: Project;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let inference: InferencePage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let inferenceName: string;
        let inferenceWorkload: Inference;

        const currentUser: User = flowRootState.users.system_administrator;

        function findItem(
            items: Array<Annotation | Label> | null,
            itemName: string,
            itemValue: string
        ): boolean {
            const item = items?.find((item) => item.name === itemName && item.value === itemValue);
            return !!item;
        }

        async function initBrowser(): Promise<BrowserEngine> {
            // Initialize browser once for all tests with video recording AND tracing support
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            inference = new InferencePage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);

            await auth.signIn(currentUser.email, currentUser.password);

            return browserEngine;
        }

        test.beforeAll('before all create inference workload based on custom UI', async function () {
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

        test.afterAll('after all create inference workload based on custom UI', async ({}, testInfo) => {
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

        test.beforeEach('before each create inference workload based on custom UI', async function () {
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

        test.afterEach('after each create inference workload based on custom UI', async ({}, testInfo) => {
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

        supportedFrom(SupportedVersions.V2_19)(function () {
            test('should create new flat inference from scratch with required fields only', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'inference any-workload submission path does not support exposedUrls yet'
                );
                const servingPort: IServingPortOptions = {
                    port: 4000
                };

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                        tool: 'Jupyter',
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                        },
                        servingPort
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // auto scaling
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceWorkload.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;
            });

            test('should create new flat inference from scratch with compute', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'inference any-workload submission path does not support exposedUrls yet'
                );
                const servingPort: IServingPortOptions = {
                    protocol: 'gRPC',
                    port: 4000,
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 2,
                    maxReplicas: 4,
                    thresholdValue: 50
                };

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.Never,
                        tool: 'Jupyter',
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                        },
                        servingPort
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '3'
                        },
                        autoScaling
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // auto scaling
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.equal(2);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceWorkload.spec.autoscaling.metric).to.equal('throughput');
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.equal(
                    autoScaling.thresholdValue
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal(
                    'authorizedUsersOrGroups'
                );
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;
            });

            test('should create new flat inference from scratch with compute with custom auto-scaling period', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'inference any-workload submission path does not support exposedUrls yet'
                );
                const servingPort: IServingPortOptions = {
                    protocol: 'gRPC',
                    port: 4000,
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 0,
                    maxReplicas: 4,
                    thresholdValue: 50,
                    autoScalingPeriod: '1m'
                };

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.Never,
                        tool: 'Jupyter',
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                        },
                        servingPort
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '3'
                        },
                        autoScaling
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // auto scaling
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.equal(0);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceWorkload.spec.autoscaling.metric).to.equal('throughput');
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.equal(
                    autoScaling.thresholdValue
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(60); // 1m

                // serving port
                expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal(
                    'authorizedUsersOrGroups'
                );
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;
            });

            test('should create new flat inference using an existing environment and compute assets', async function () {
                const servingPort: IServingPortOptions = {
                    protocol: 'HTTP',
                    port: 4000
                };

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 2,
                    maxReplicas: 2
                };

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        assetName: 'llm-server',
                        servingPort,
                        runtimeSettings: {
                            environmentVariablesValues: ['3000', '4000']
                        }
                    },
                    computeSection: {
                        assetName: 'cpu-only',
                        autoScaling
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // auto scaling
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.equal(autoScaling.minReplicas);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceWorkload.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;
            });

            test('should create new flat inference with new nfs from drawer', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'inference any-workload submission path does not support storage fields (new PVC / NFS / hostPath / configMap volume) yet'
                );
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                const dataSourceOptions: IDataSourceOptions = {
                    nfs: dataSourceName
                };

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        assetName: 'llm-server',
                        servingPort: {
                            protocol: 'HTTP',
                            port: 4000
                        },
                        runtimeSettings: {
                            environmentVariablesValues: ['3000', '4000']
                        }
                    },
                    computeSection: {
                        assetName: 'cpu-only'
                    },
                    dataSourceOptions
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.storage.nfs).to.not.be.undefined;
            });

            test('should create new flat inference with annotations and labels', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'inference any-workload submission path does not support exposedUrls yet'
                );
                const servingPort: IServingPortOptions = {
                    port: 4000
                };

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                        tool: 'Jupyter',
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                        },
                        servingPort
                    },
                    generalSectionOptions: {
                        annotations: [
                            { name: 'ann1', value: 'ann-val1' },
                            { name: 'ann2', value: 'ann-val2' }
                        ],
                        labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                expect(findItem(inferenceWorkload.spec.annotations, 'ann1', 'ann-val1')).to.be.true;
                expect(findItem(inferenceWorkload.spec.annotations, 'ann2', 'ann-val2')).to.be.true;
                expect(findItem(inferenceWorkload.spec.labels, 'lbl1', 'lbl-val1')).to.be.true;
            });

            supportedFrom(SupportedVersions.V2_20)(function () {
                test('should create new workload with new environment asset and compute resource asset and data sources', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'inference any-workload submission path does not support storage fields (new PVC / NFS / hostPath / configMap volume) yet'
                    );
                    const environmentName = objectGenerator.generateNameByPrefix('custom-environment');
                    const computeName = objectGenerator.generateNameByPrefix('custom-compute');
                    const servingPort: IServingPortOptions = {
                        protocol: 'HTTP',
                        port: 8000
                    };
                    const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                    const volumeOptions: IVolumeOptions = {
                        accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                        sizeUnit: 'MB',
                        path: '/home'
                    };

                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            newEnvironmentResource: {
                                assetName: environmentName,
                                projectName: project.name,
                                image: config.quickStartImage.inferenceHelloWorld,
                                servingPort: {
                                    container: servingPort.port
                                }
                            }
                        },
                        computeSection: {
                            newComputeResource: {
                                assetName: computeName,
                                projectName: project.name,
                                gpu: {
                                    devicesInput: 3
                                }
                            }
                        },
                        dataSourceOptions: {
                            nfs: dataSourceName,
                            volume: volumeOptions
                        }
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    expect(inferenceWorkload.spec.image).to.be.equal(
                        config.quickStartImage.inferenceHelloWorld
                    );
                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(servingPort.port);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal(
                        servingPort.protocol.toLowerCase()
                    );
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.be.equal(1);
                    expect(inferenceWorkload.spec.compute.gpuPortionRequest).to.be.equal(0.03);
                    expect(inferenceWorkload.spec.compute.gpuRequestType).to.be.equal('portion');
                    expect(inferenceWorkload.spec.storage.nfs).to.deep.equal([
                        {
                            path: '/test/sample',
                            readOnly: false,
                            server: 'test-server',
                            mountPath: '/home/local/data',
                            name: undefined,
                            exclude: undefined
                        }
                    ]);

                    const pvc = inferenceWorkload.spec.storage.pvc[0];
                    expect(pvc).to.be.exist;
                    expect(pvc.path).to.equal('/home');
                    expect(pvc.claimInfo.size).to.equal('1M');
                    expect(pvc.claimInfo.storageClass).to.be.exist;
                    expect(pvc.claimInfo.accessModes.readWriteOnce).to.equal(true);
                    expect(pvc.claimInfo.accessModes.readOnlyMany).to.be.undefined;
                    expect(pvc.claimInfo.accessModes.readWriteMany).to.be.undefined;

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);

                    // cleanup
                    const computeId = await computeResourceUtils.getId(computeName);
                    await computeResourceUtils.delete(computeId);
                    await computeResourceUtils.waitForComputeResourceToBeDeleted(computeName);
                });

                test('should create new flat inference with volume', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'inference any-workload submission path does not support storage fields (new PVC / NFS / hostPath / configMap volume) yet'
                    );
                    const volumeOptions: IVolumeOptions = {
                        accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                        sizeUnit: 'MB',
                        path: '/home'
                    };

                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            assetName: 'llm-server',
                            servingPort: {
                                protocol: 'HTTP',
                                port: 4000
                            },
                            runtimeSettings: {
                                environmentVariablesValues: ['3000', '4000']
                            }
                        },
                        computeSection: {
                            assetName: 'cpu-only'
                        },
                        dataSourceOptions: {
                            volume: volumeOptions
                        }
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    expect(inferenceWorkload.spec.storage.pvc).to.not.be.undefined;
                });

                test('should copy flat inference', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'inference any-workload submission path does not support storage fields (new PVC / NFS / hostPath / configMap volume) yet'
                    );
                    const copiedInferenceName = objectGenerator.generateShortNameByPrefix('copied-inference');

                    const servingPort: IServingPortOptions = {
                        protocol: 'HTTP',
                        port: 4000
                    };
                    const autoScaling: IAutoScalingOptions = {
                        minReplicas: 2,
                        maxReplicas: 3
                    };

                    const dataSourceOptions: IDataSourceOptions = {
                        nfs: objectGenerator.generateNameByPrefix('inf-nfs-storage'),
                        volume: {
                            accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                            sizeUnit: 'MB',
                            path: '/home'
                        }
                    };

                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            assetName: 'llm-server',
                            servingPort,
                            runtimeSettings: {
                                environmentVariablesValues: ['3000', '4000']
                            }
                        },
                        computeSection: {
                            assetName: 'cpu-only',
                            autoScaling
                        },
                        dataSourceOptions
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`originInferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    await workload.toggleWorkload(inferenceName);
                    await workload.clickOnCopyAndEditWorkload();

                    await inference.fillInferenceName(copiedInferenceName);
                    await inference.clickOnCreateInferenceBtn();

                    await waitUtils.waitForElementToExist(workload, `td:has-text("${copiedInferenceName}")`);

                    const inferenceCopiedWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        copiedInferenceName,
                        project.id.toString()
                    );
                    expect(inferenceCopiedWorkload).to.not.be.undefined;
                    logger.debug(`inferenceCopiedWorkload: ${JSON.stringify(inferenceCopiedWorkload)}`);

                    // this is a temp solution until we fix the security defaults
                    // for llm-server and huggingface-chat-ui
                    inferenceCopiedWorkload.spec.security = undefined;
                    inferenceCopiedWorkload.spec.storage.pvc[0].claimName = undefined;
                    inferenceCopiedWorkload.spec.storage.pvc[0].ephemeral = undefined;

                    logger.debug(
                        `inferenceCopiedWorkload (after cleanup): ${JSON.stringify(inferenceCopiedWorkload)}`
                    );
                    expect(inferenceWorkload.spec).to.deep.equal(inferenceCopiedWorkload.spec);
                    logger.debug(`inference ${copiedInferenceName} copied test ended successfully!`);
                });

                test('should update flat inference', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'inference any-workload submission path does not support storage fields (new PVC / NFS / hostPath / configMap volume) yet'
                    );
                    const autoScaling: IAutoScalingOptions = {
                        minReplicas: 1,
                        maxReplicas: 3
                    };

                    const dataSourceOptions: IDataSourceOptions = {
                        nfs: objectGenerator.generateNameByPrefix('inf-nfs-storage'),
                        volume: {
                            accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                            sizeUnit: 'MB',
                            path: '/home'
                        }
                    };

                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            assetName: 'llm-server',
                            runtimeSettings: {
                                environmentVariablesValues: ['3000', '4000']
                            }
                        },
                        computeSection: {
                            assetName: 'cpu-only',
                            autoScaling
                        },
                        dataSourceOptions
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    const envUpdates: IEnvironmentOptions = {
                        image: 'ubuntu',
                        imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                        runtimeSettings: {
                            environmentVariablesValues: ['3001', '4001']
                        }
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
                        },
                        autoScaling: {
                            minReplicas: 0,
                            maxReplicas: 5,
                            thresholdValue: 50,
                            autoScalingPeriod: '1m'
                        },
                        nodes: {
                            nodeType: 'some-node'
                        }
                    };

                    const workloadData: IUpdateWorkload = {
                        environmentSection: envUpdates,
                        computeSection: computeUpdates
                    };

                    await inferenceFlatUIUtil.updateWorkload(inferenceName, workloadData);
                    const updatedInferenceWorkload: Inference = (
                        await workloadsCrudUtils.getInference(inferenceWorkload.workloadId)
                    ).body;
                    logger.debug(`updatedInferenceWorkload: ${JSON.stringify(updatedInferenceWorkload)}`);
                    expect(updatedInferenceWorkload.spec.image).to.equal(envUpdates.image);
                    expect(updatedInferenceWorkload.spec.imagePullPolicy).to.equal('Always');
                    expect(updatedInferenceWorkload.spec.compute.gpuDevicesRequest).to.equal(3);
                    expect(updatedInferenceWorkload.spec.compute.gpuRequestType).to.equal('portion');
                    expect(updatedInferenceWorkload.spec.compute.gpuPortionRequest).to.equal(0.3);
                    expect(updatedInferenceWorkload.spec.compute.cpuCoreRequest).to.equal(0.004);
                    expect(updatedInferenceWorkload.spec.compute.cpuMemoryRequest).to.equal('5M');
                    expect(updatedInferenceWorkload.spec.compute.largeShmRequest).to.equal(true);
                    const resources = computeUpdates.extendedResource.resources.map((resource) => {
                        return {
                            ...resource,
                            exclude: undefined
                        };
                    });
                    expect(updatedInferenceWorkload.spec.compute.extendedResources).to.deep.equal(resources);
                    expect(updatedInferenceWorkload.spec.autoscaling.minReplicas).to.equal(0);
                    expect(updatedInferenceWorkload.spec.autoscaling.maxReplicas).to.equal(
                        computeUpdates.autoScaling.maxReplicas
                    );
                    expect(updatedInferenceWorkload.spec.autoscaling.metric).to.equal('throughput');

                    expect(updatedInferenceWorkload.spec.autoscaling.metricThreshold).to.equal(
                        computeUpdates.autoScaling.thresholdValue
                    );

                    expect(updatedInferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(
                        60
                    ); // computeUpdates.autoScaling.autoScalingPeriod

                    logger.debug(`inference ${inferenceName} updated successfully!`);
                });


                test('should copy inference workload with volume correctly', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'inference any-workload submission path does not support storage fields (new PVC / NFS / hostPath / configMap volume) yet'
                    );
                    const copiedInferenceName = objectGenerator.generateShortNameByPrefix('copied-inference');
                    const volumeOptions: IVolumeOptions = {
                        accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                        sizeUnit: 'MB',
                        path: '/home'
                    };

                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            assetName: 'llm-server',
                            servingPort: {
                                protocol: 'HTTP',
                                port: 4000
                            },
                            runtimeSettings: {
                                environmentVariablesValues: ['3000', '4000']
                            }
                        },
                        computeSection: {
                            assetName: 'cpu-only'
                        },
                        dataSourceOptions: {
                            volume: volumeOptions
                        }
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    expect(inferenceWorkload.spec.storage.pvc).to.not.be.undefined;

                    await workload.toggleWorkload(inferenceName);
                    await workload.clickOnCopyAndEditWorkload();

                    await inference.fillInferenceName(copiedInferenceName);
                    await inference.clickOnCreateInferenceBtn();

                    await waitUtils.waitForElementToExist(workload, `td:has-text("${copiedInferenceName}")`);

                    const inferenceCopiedWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        copiedInferenceName,
                        project.id.toString()
                    );
                    expect(inferenceCopiedWorkload).to.not.be.undefined;
                    logger.debug(`inferenceCopiedWorkload: ${JSON.stringify(inferenceCopiedWorkload)}`);
                    expect(inferenceCopiedWorkload.spec.storage.pvc).to.not.be.undefined;
                    logger.debug(`inference ${copiedInferenceName} copied test ended successfully!`);
                });
                
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test('should create new flat inference with custom request timeout values', async function () {
                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            image: 'ubuntu',
                            servingPort: {
                                port: 4000
                            }
                        },
                        generalSectionOptions: {
                            initializationTimeout: {
                                minutes: 5,
                                seconds: 7
                            },
                            requestTimeout: {
                                minutes: 2,
                                seconds: 30
                            }
                        }
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    expect(
                        inferenceWorkload.spec.servingConfiguration?.initializationTimeoutSeconds
                    ).to.equal(307); // 5 minutes * 60 (seconds) + 7 seconds
                    expect(inferenceWorkload.spec.servingConfiguration?.requestTimeoutSeconds).to.equal(150); // 2 minutes * 60 (seconds) + 30 seconds
                });

                test('should update request timeout values', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'update is disabled for inference created via the any-workload/v2 path (cannot update through the legacy path)'
                    );
                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            image: 'ubuntu',
                            servingPort: {
                                port: 4000
                            }
                        },
                        generalSectionOptions: {
                            initializationTimeout: {
                                minutes: 10
                            },
                            requestTimeout: {
                                minutes: 5
                            }
                        }
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    const defaultInitializationTimeoutSeconds = 600; // 10 minutes in seconds
                    const defaultRequestTimeoutSeconds = 300; // 5 minutes in seconds
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    expect(
                        inferenceWorkload.spec.servingConfiguration?.initializationTimeoutSeconds
                    ).to.equal(defaultInitializationTimeoutSeconds);
                    expect(inferenceWorkload.spec.servingConfiguration?.requestTimeoutSeconds).to.equal(
                        defaultRequestTimeoutSeconds
                    );

                    const generalSectionOptions: IGeneralSectionOptions = {
                        initializationTimeout: {
                            minutes: 5,
                            seconds: 7
                        },
                        requestTimeout: {
                            minutes: 2,
                            seconds: 30
                        }
                    };

                    await inferenceFlatUIUtil.updateWorkload(inferenceName, {
                        generalSectionOptions
                    });
                    const updatedInferenceWorkload: Inference = (
                        await workloadsCrudUtils.getInference(inferenceWorkload.workloadId)
                    ).body;

                    expect(updatedInferenceWorkload).to.not.be.undefined;
                    logger.debug(`updatedInferenceWorkload: ${JSON.stringify(updatedInferenceWorkload)}`);
                    expect(
                        updatedInferenceWorkload.spec.servingConfiguration?.initializationTimeoutSeconds
                    ).to.equal(307); // 5 minutes * 60 (seconds) + 7 seconds
                    expect(
                        updatedInferenceWorkload.spec.servingConfiguration?.requestTimeoutSeconds
                    ).to.equal(150); // 2 minutes * 60 (seconds) + 30 seconds
                });

                test('should copy and edit request timeout values', async function () {
                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.Custom,
                        environmentSection: {
                            image: 'ubuntu',
                            servingPort: {
                                port: 4000
                            }
                        },
                        generalSectionOptions: {
                            initializationTimeout: {
                                minutes: 5,
                                seconds: 7
                            },
                            requestTimeout: {
                                minutes: 2,
                                seconds: 30
                            }
                        }
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    const initializationTimeoutSeconds = 307; // 5 minutes * 60 (seconds) + 7 seconds
                    const requestTimeoutSeconds = 150; // 2 minutes * 60 (seconds) + 30 seconds
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    expect(
                        inferenceWorkload.spec.servingConfiguration?.initializationTimeoutSeconds
                    ).to.equal(initializationTimeoutSeconds);
                    expect(inferenceWorkload.spec.servingConfiguration?.requestTimeoutSeconds).to.equal(
                        requestTimeoutSeconds
                    );

                    await workload.toggleWorkload(inferenceName);
                    await workload.clickOnCopyAndEditWorkload();

                    const copiedInferenceName = objectGenerator.generateNameByPrefix('inference-copied');
                    await inference.fillInferenceName(copiedInferenceName);
                    await inference.clickOnCreateInferenceBtn();

                    const inferenceCopiedWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        copiedInferenceName,
                        project.id.toString()
                    );
                    expect(inferenceCopiedWorkload).to.not.be.undefined;
                    logger.debug(`inferenceCopiedWorkload: ${JSON.stringify(inferenceCopiedWorkload)}`);
                    expect(
                        inferenceCopiedWorkload.spec.servingConfiguration?.initializationTimeoutSeconds
                    ).to.equal(initializationTimeoutSeconds);
                    expect(inferenceCopiedWorkload.spec.servingConfiguration?.requestTimeoutSeconds).to.equal(
                        requestTimeoutSeconds
                    );
                });
            });
        });
    }
);
